package goQemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (q *Qemu) generateCloudInit(config Config, cloudInit CloudInit) (string, error) {
	if config.Options.UUID == "" {
		return "", fmt.Errorf("UUID is required for cloud-init")
	}

	if !map[string]bool{
		"ubuntu":     true,
		"debian":     true,
		"centos":     true,
		"rockylinux": true,
		"almalinux":  true,
	}[strings.ToLower(config.OS)] {
		return "", fmt.Errorf("unsupported OS: %s", config.OS)
	}

	if cloudInit.Hostname == "" {
		cloudInit.Hostname = config.OS
	}

	if cloudInit.Username == "" {
		cloudInit.Username = config.OS
	}

	if cloudInit.Password == "" {
		cloudInit.Password = "passwd"
	}

	tmpFolder := fmt.Sprintf(".cloudinit-%d", config.ID)
	tmpFolderPath := filepath.Join(q.Folder.VM, tmpFolder)
	if err := os.MkdirAll(tmpFolderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	if os.Getenv("TEST_MODE") != "true" {
		defer os.RemoveAll(tmpFolderPath)
	}

	// * generate meta-data
	metaData := fmt.Sprintf(`
instance-id: %s
local-hostname: %s
`, config.Options.UUID, cloudInit.Hostname)
	metaDataPath := filepath.Join(tmpFolderPath, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// * generate user-data
	sshKey := cloudInit.AuthorizedKey
	if sshKey == "" {
		homeDir, _ := os.UserHomeDir()
		keyPaths := []string{
			filepath.Join(homeDir, ".ssh", "id_ed25519.pub"),
			filepath.Join(homeDir, ".ssh", "id_rsa.pub"),
			filepath.Join(homeDir, ".ssh", "id_ecdsa.pub"),
		}

		for _, keyPath := range keyPaths {
			if data, err := os.ReadFile(keyPath); err == nil {
				sshKey = strings.TrimSpace(string(data))
				break
			}
		}

		// * pubkey not exist, then generate
		if sshKey == "" {
			privateKeyPath := filepath.Join(homeDir, ".ssh", "id_ed25519")
			publicKeyPath := privateKeyPath + ".pub"

			cmd := exec.Command("ssh-keygen",
				"-t", "ed25519",
				"-f", privateKeyPath,
				"-N", "",
			)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to generate SSH key: %w", err)
			}
			if data, err := os.ReadFile(publicKeyPath); err == nil {
				sshKey = strings.TrimSpace(string(data))
			}
		}
	}

	upgradePackages := "false"
	if cloudInit.UpgradePackages {
		upgradePackages = "true"
	}

	userData := fmt.Sprintf(`#cloud-config
users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - %s
    shell: /bin/bash

ssh_pwauth: true

chpasswd:
  list: |
    %s:%s
  expire: false

package_upgrade: %s

packages:
  - qemu-guest-agent
`, cloudInit.Username, sshKey, cloudInit.Username, cloudInit.Password, upgradePackages)

	// if cloudInit.NetworkConfig != nil {
	dnsConfig := q.generateDNSConfig(&cloudInit)
	if dnsConfig != "" {
		userData += "\n" + dnsConfig
	}
	// }

	userData += `
runcmd:
  - [ sh, -c, 'ping -c 3 $(ip route | grep default | awk "{print \$3}") >/dev/null 2>&1 &' ]
  - [ sh, -c, '(sleep 3 && rm -rf /var/lib/cloud/instance /var/lib/cloud/instances/*) &' ]
  - [ systemctl, enable, qemu-guest-agent ]
  - [ systemctl, start, qemu-guest-agent ]
`

	userDataPath := filepath.Join(tmpFolderPath, "user-data")
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	isoFiles := []string{metaDataPath, userDataPath}

	// if cloudInit.NetworkConfig != nil {
	networkConfigData := q.generateNetworkConfigFile(&cloudInit)
	if networkConfigData != "" {
		networkConfigPath := filepath.Join(tmpFolderPath, "network-config")
		if err := os.WriteFile(networkConfigPath, []byte(networkConfigData), 0644); err != nil {
			return "", fmt.Errorf("failed to write network-config: %w", err)
		}
		isoFiles = append(isoFiles, networkConfigPath)
	}
	// }

	ISO := fmt.Sprintf("%d-cloud-init.iso", config.ID)
	ISOPath := filepath.Join(q.Folder.VM, ISO)

	var cmd *exec.Cmd
	if _, err := exec.LookPath("genisoimage"); err == nil {
		args := []string{
			"-output", ISOPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
		}
		args = append(args, isoFiles...)
		cmd = exec.Command("genisoimage", args...)
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		args := []string{
			"-output", ISOPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
		}
		args = append(args, isoFiles...)
		cmd = exec.Command("mkisofs", args...)
	} else {
		return "", fmt.Errorf("failed to create cloud-init ISO: neither genisoimage nor mkisofs found in system")
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create cloud-init ISO: %w", err)
	}

	fmt.Printf("[*] created cloud-init ISO: %s\n", ISOPath)
	return ISOPath, nil
}

func (q *Qemu) generateDNSConfig(netConfig *CloudInit) string {
	if netConfig == nil {
		return ""
	}

	if len(netConfig.DNSServers) == 0 && netConfig.DNSDomain == "" {
		return ""
	}

	config := "manage_resolv_conf: true\n"
	config += "resolv_conf:\n"

	if len(netConfig.DNSServers) > 0 {
		config += "  nameservers:\n"
		for _, dns := range netConfig.DNSServers {
			config += fmt.Sprintf("    - %s\n", dns)
		}
	}

	if netConfig.DNSDomain != "" {
		config += "  searchdomains:\n"
		config += fmt.Sprintf("    - %s\n", netConfig.DNSDomain)
	}

	return config
}

func (q *Qemu) generateNetworkConfigFile(netConfig *CloudInit) string {
	if netConfig == nil {
		return ""
	}

	hasIPv4 := netConfig.IPv4 != nil && netConfig.IPv4.Mode == "static" && netConfig.IPv4.Address != ""
	hasIPv6 := netConfig.IPv6 != nil && netConfig.IPv6.Mode == "static" && netConfig.IPv6.Address != ""

	if !hasIPv4 && !hasIPv6 {
		return ""
	}

	config := "version: 2\nethernets:\n  eth0:\n"
	addresses := []string{}

	if hasIPv4 && netConfig.IPv4.Address != "" {
		addresses = append(addresses, netConfig.IPv4.Address)
	}

	if hasIPv6 && netConfig.IPv6.Address != "" {
		addresses = append(addresses, netConfig.IPv6.Address)
	}

	if len(addresses) > 0 {
		config += "    addresses:\n"
		for _, addr := range addresses {
			config += fmt.Sprintf("      - %s\n", addr)
		}
	}

	if hasIPv4 && netConfig.IPv4.Gateway != "" {
		config += fmt.Sprintf("    gateway4: %s\n", netConfig.IPv4.Gateway)
	}

	if hasIPv6 && netConfig.IPv6 != nil && netConfig.IPv6.Gateway != "" {
		config += fmt.Sprintf("    gateway6: %s\n", netConfig.IPv6.Gateway)
	}

	dhcp4 := "no"
	dhcp6 := "no"

	if netConfig.IPv4 == nil || netConfig.IPv4.Mode == "dhcp" {
		if !hasIPv4 {
			dhcp4 = "yes"
		}
	}

	if netConfig.IPv6 != nil {
		if netConfig.IPv6.Mode == "dhcp" {
			dhcp6 = "yes"
		} else if netConfig.IPv6.Mode == "slaac" {
			dhcp6 = "no"
			config += "    accept-ra: yes\n"
		}
	}

	config += fmt.Sprintf("    dhcp4: %s\n", dhcp4)
	config += fmt.Sprintf("    dhcp6: %s\n", dhcp6)

	if len(netConfig.DNSServers) > 0 {
		config += "    nameservers:\n"
		config += "      addresses:\n"
		for _, dns := range netConfig.DNSServers {
			config += fmt.Sprintf("        - %s\n", dns)
		}

		if netConfig.DNSDomain != "" {
			config += "      search:\n"
			config += fmt.Sprintf("        - %s\n", netConfig.DNSDomain)
		}
	}

	return config
}

func (q *Qemu) removeCloudInit(vmid int) {
	ISO := fmt.Sprintf("%d-cloud-init.iso", vmid)
	ISOPath := filepath.Join(q.Folder.VM, ISO)
	os.Remove(ISOPath)
}

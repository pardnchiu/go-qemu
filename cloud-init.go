package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (m *Folder) NewCloudInit(vmid int, config CloudInit) (string, error) {
	if config.UUID == "" {
		return "", fmt.Errorf("UUID is required for cloud-init")
	} else if !map[string]bool{
		"ubuntu":     true,
		"debian":     true,
		"rockylinux": true,
	}[strings.ToLower(config.OS)] {
		return "", fmt.Errorf("unsupported OS: %s", config.OS)
	}

	if config.Hostname == "" {
		config.Hostname = config.OS
	}

	if config.Username == "" {
		config.Username = config.OS
	}

	if config.Password == "" {
		config.Password = "passwd"
	}

	tmpFolder := fmt.Sprintf(".cloudinit-%d", vmid)
	tmpFolderPath := filepath.Join(m.VM, tmpFolder)
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
`, config.UUID, config.Hostname)
	metaDataPath := filepath.Join(tmpFolderPath, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// * generate user-data
	sshKey := config.SSHAuthorizedKey
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

	userData := fmt.Sprintf(`
users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - %s
    shell: /bin/bash

ssh_pwauth: false

chpasswd:
  list: |
    %s:%s
  expire: false

runcmd:
  - [ sh, -c, '(sleep 3 && rm -rf /var/lib/cloud/instance /var/lib/cloud/instances/*) &' ]
`, config.Username, sshKey, config.Username, config.Password)
	userDataPath := filepath.Join(tmpFolderPath, "user-data")
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	ISO := fmt.Sprintf("%d-cloud-init.iso", vmid)
	ISOPath := filepath.Join(m.VM, ISO)

	var cmd *exec.Cmd
	if _, err := exec.LookPath("genisoimage"); err == nil {
		cmd = exec.Command("genisoimage",
			"-output", ISOPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
			metaDataPath,
			userDataPath,
		)
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		cmd = exec.Command("mkisofs",
			"-output", ISOPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
			metaDataPath,
			userDataPath,
		)
	} else {
		return "", fmt.Errorf("failed to create cloud-init ISO: neither genisoimage nor mkisofs found in system")
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create cloud-init ISO: %w", err)
	}

	fmt.Printf("successfully created cloud-init ISO: %s\n", ISOPath)
	return ISOPath, nil
}

func (m *Folder) RemoveCloudInit(vmid int) {
	ISO := fmt.Sprintf("%d-cloud-init.iso", vmid)
	ISOPath := filepath.Join(m.VM, ISO)
	os.Remove(ISOPath)
}

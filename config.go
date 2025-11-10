package goQemu

import (
	"fmt"
	"os"
	"strconv"
)

func (q *Qemu) verifyConfig(config Config) (*Config, error) {
	vmidStart, err := strconv.Atoi(os.Getenv("GO_QEMU_VMID_START"))
	if err != nil {
		vmidStart = 100
	}

	vmidEnd, err := strconv.Atoi(os.Getenv("GO_QEMU_VMID_END"))
	if err != nil {
		vmidEnd = 999
	}

	if config.ID == 0 {
		return nil, fmt.Errorf("VMID must be specified")
	} else if config.ID < vmidStart || config.ID > vmidEnd {
		return nil, fmt.Errorf("VMID must be between %d and %d", vmidStart, vmidEnd)
	}

	if config.Hostname == "" {
		return nil, fmt.Errorf("hostname must be specified")
	}

	// if config.Username == "" {
	// 	return nil, fmt.Errorf("username must be specified")
	// }

	if config.UUID == "" {
		return nil, fmt.Errorf("UUID must be specified")
	}

	if config.DiskPath == "" {
		return nil, fmt.Errorf("disk_path must be specified")
	}

	config.VNCPort = 59000 + config.ID

	if len(config.Network) == 0 {
		config.Network = []Network{
			{
				Bridge:     "vmbr0",
				Model:      "virtio-net-pci",
				Vlan:       0,
				MACAddress: generateMAC(config.ID),
				Firewall:   false,
				Disconnect: false,
				MTU:        1500,
				RateLimit:  0,
				Multiqueue: 0,
			},
		}
	}

	cloudInitConfig := config.CloudInit
	// if config.CloudInitPath == "" && config.OS != "" {
	// 	cloudInitConfig = CloudInit{
	// 		UUID: config.UUID,
	// 		// OS:               config.OS,
	// 		Hostname:         config.Hostname,
	// 		Username:         config.Username,
	// 		Password:         config.Password,
	// 		SSHAuthorizedKey: config.SSHAuthorizedKey,
	// 		UpgradePackages:  true,
	// 		NetworkConfig:    nil,
	// 	}
	// }

	cloudInitPath, err := q.generateCloudInit(config, cloudInitConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cloud-init: %w", err)
	}
	config.CloudInit = cloudInitConfig
	config.CloudInitPath = cloudInitPath

	return &config, nil
}

func generateMAC(vmid int) string {
	// use QEMU OUI 52:54:00
	return fmt.Sprintf("52:54:00:00:%02X:%02X", (vmid>>8)&0xFF, vmid&0xFF)
}

package goQemu

import (
	"fmt"
	"os"
	"strconv"
)

func (q *Qemu) assignVMID() (int, error) {
	ids, err := os.ReadDir(q.Folder.Config)
	if err != nil {
		return 0, fmt.Errorf("failed to read go-qemu/configs: %w", err)
	}

	ary := make(map[int]bool, len(ids))
	for _, id := range ids {
		if id.IsDir() {
			continue
		}

		var vmid int
		if _, err := fmt.Sscanf(id.Name(), "%d.json", &vmid); err == nil {
			ary[vmid] = true
		}
	}

	for id := 100; id <= 999; id++ {
		if !ary[id] {
			return id, nil
		}
	}

	return 0, fmt.Errorf("no available VMID can be assigned")
}

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

	if config.Username == "" {
		return nil, fmt.Errorf("username must be specified")
	}

	if config.UUID == "" {
		return nil, fmt.Errorf("UUID must be specified")
	}

	if config.DiskPath == "" {
		return nil, fmt.Errorf("disk_path must be specified")
	}

	config.VNCPort = 59000 + config.ID

	// TODO: update this after update network
	config.SSHPort = 22000 + config.ID

	if config.CloudInit == "" && config.OS != "" {
		cloudInitConfig := CloudInit{
			UUID:             config.UUID,
			OS:               config.OS,
			Hostname:         config.Hostname,
			Username:         config.Username,
			Password:         config.Password,
			SSHAuthorizedKey: config.SSHAuthorizedKey,
		}

		cloudInitPath, err := q.generateCloudInit(config.ID, cloudInitConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cloud-init: %w", err)
		}
		config.CloudInit = cloudInitPath
	}

	return &config, nil
}

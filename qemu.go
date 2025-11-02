package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func NewQemu() (*Folder, error) {
	mainPath := os.Getenv("GO_QEMU_PATH")
	if mainPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if usr.HomeDir == "" {
			return nil, fmt.Errorf("user home directory is empty")
		}
		mainPath = filepath.Join(usr.HomeDir, "go-qemu")
	}

	if err := os.MkdirAll(mainPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu: %w", err)
	}

	vmsPath := filepath.Join(mainPath, "vms")
	if err := os.MkdirAll(vmsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/vms: %w", err)
	}

	configsPath := filepath.Join(mainPath, "configs")
	if err := os.MkdirAll(configsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/configs: %w", err)
	}

	logsPath := filepath.Join(mainPath, "logs")
	if err := os.MkdirAll(logsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/logs: %w", err)
	}

	pidsPath := filepath.Join(mainPath, "pids")
	if err := os.MkdirAll(pidsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/pids: %w", err)
	}

	monitorsPath := filepath.Join(mainPath, "monitors")
	if err := os.MkdirAll(monitorsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/monitors: %w", err)
	}

	imagesPath := filepath.Join(mainPath, "images")
	if err := os.MkdirAll(imagesPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder go-qemu/images: %w", err)
	}

	folder := &Folder{
		VM:      vmsPath,
		Config:  configsPath,
		Log:     logsPath,
		PID:     pidsPath,
		Monitor: monitorsPath,
		Image:   imagesPath,
	}

	return folder, nil
}

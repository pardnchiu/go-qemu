package goQemu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
)

func NewQemu() (*Qemu, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Info(".env not found, use system env")
	}

	mainPath := os.Getenv("GO_QEMU_PATH")
	if mainPath == "" {
		// * not assigned, use user home
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

	var binary string
	switch runtime.GOARCH {
	case "amd64", "386":
		binary = "qemu-system-x86_64"
	case "arm64", "arm":
		binary = "qemu-system-aarch64"
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	qemu := &Qemu{
		Folder: Folder{
			VM:      vmsPath,
			Config:  configsPath,
			Log:     logsPath,
			PID:     pidsPath,
			Monitor: monitorsPath,
			Image:   imagesPath,
		},
		Binary: binary,
	}

	return qemu, nil
}

func (q *Qemu) saveConfig(config Config) error {
	targetName := fmt.Sprintf("%d.json", config.ID)
	targetPath := filepath.Join(q.Folder.Config, targetName)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, data, 0644)
}

func (q *Qemu) loadConfig(vmid int) (*Config, error) {
	targetName := fmt.Sprintf("%d.json", vmid)
	targetPath := filepath.Join(q.Folder.Config, targetName)
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	verifyConfig, err := q.verifyConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return verifyConfig, nil
}

func (q *Qemu) deleteConfig(vmid int) error {
	targetName := fmt.Sprintf("%d.json", vmid)
	targetPath := filepath.Join(q.Folder.Config, targetName)
	return os.Remove(targetPath)
}

func (q *Qemu) isRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (q *Qemu) Cleanup() error {
	ids, err := os.ReadDir(q.Folder.Config)
	if err != nil {
		return err
	}

	cleaned := 0
	for _, entry := range ids {
		var vmid int
		if _, err := fmt.Sscanf(entry.Name(), "%d.json", &vmid); err == nil {
			if pidFilePath, pidData, err := q.getFile(q.Folder.PID, vmid); err == nil {
				var pid int
				fmt.Sscanf(pidData, "%d", &pid)

				if !q.isRunning(pid) {
					os.Remove(pidFilePath)
					cleaned++
				}
			}
		}
	}

	fmt.Printf("[*] cleaned up %d unused VM(s)\n", cleaned)
	return nil
}

func (q *Qemu) getFile(folderPath string, vmid int) (string, string, error) {
	var targetName string
	switch folderPath {
	case q.Folder.PID:
		targetName = fmt.Sprintf("%d.pid", vmid)
	case q.Folder.Monitor:
		targetName = fmt.Sprintf("%d.sock", vmid)
	case q.Folder.Config:
		targetName = fmt.Sprintf("%d.json", vmid)
	case q.Folder.Log:
		targetName = fmt.Sprintf("%d.log", vmid)
	default:
		return "", "", fmt.Errorf("unsupported folder path: %s", folderPath)
	}

	targetPath := filepath.Join(folderPath, targetName)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return targetPath, "", fmt.Errorf("file does not exist: %s/%s", targetPath, targetName)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		return targetPath, "", err
	}

	return targetPath, string(content), nil
}

func (q *Qemu) diskPathAll(vmid int) ([]string, error) {
	extensions := []string{"img", "qcow2"}
	var matches []string

	for _, ext := range extensions {
		pattern := fmt.Sprintf("%d-*.%s", vmid, ext)
		path := filepath.Join(q.Folder.VM, pattern)
		files, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("failed to search for disk files: %w", err)
		}

		matches = append(matches, files...)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no disk files found for VM %d", vmid)
	}

	return matches, nil
}

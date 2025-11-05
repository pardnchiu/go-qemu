package goQemu

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func (q *Qemu) Start(vmid int) error {
	q.Cleanup()

	config, err := q.loadConfig(vmid)
	if err != nil {
		return fmt.Errorf("failed to get vm-%d config: %w", vmid, err)
	}

	if pidFilepath, pidContent, err := q.getFile(q.Folder.PID, vmid); err == nil {
		var pid int
		fmt.Sscanf(pidContent, "%d", &pid)

		if q.isRunning(pid) {
			return fmt.Errorf("VM %d is already running: PID %d", vmid, pid)
		}
		os.Remove(pidFilepath)
	}

	if _, err := os.Stat(config.DiskPath); err != nil {
		return fmt.Errorf("disk not found: %s", config.DiskPath)
	}

	if _, err := os.Stat(config.BIOSPath); err != nil {
		return fmt.Errorf("BIOS not found: %s", config.BIOSPath)
	}

	pid, err := q.runVM(config, vmid)
	if err != nil {
		return err
	}

	fmt.Printf("VM %d started with PID %d\n", vmid, pid)
	return nil
}

func (q *Qemu) checkHVFSlots() error {
	cmd := exec.Command("pgrep", "-c", "qemu-system-aarch64")
	output, _ := cmd.Output()

	count := 0
	fmt.Sscanf(string(output), "%d", &count)

	if count >= 60 {
		return fmt.Errorf("too many VMs running (%d), HVF slots may be exhausted", count)
	}

	return nil
}

func (q *Qemu) runVM(config *Config, vmid int) (int, error) {
	if err := q.checkHVFSlots(); err != nil {
		return 0, fmt.Errorf("HVF resource issue: %w\nTry: ./qmac cleanup", err)
	}

	logFile := fmt.Sprintf("%d.log", vmid)
	logFilePath := filepath.Join(q.Folder.Log, logFile)
	logOut, err := os.Create(logFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logOut.Close()

	var binary string
	switch runtime.GOARCH {
	case "amd64", "386":
		binary = "qemu-system-x86_64"
	case "arm64", "arm":
		binary = "qemu-system-aarch64"
	default:
		return 0, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	args := q.verifyArgs(*config)
	cmd := exec.Command(binary, args...)
	cmd.Dir = q.Folder.VM
	cmd.Stdout = logOut
	cmd.Stderr = logOut
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start VM: %w", err)
	}

	pid := cmd.Process.Pid
	pidFile := fmt.Sprintf("%d.pid", vmid)
	pidFilePath := filepath.Join(q.Folder.PID, pidFile)
	if err := os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return 0, fmt.Errorf("failed to save PID: %w", err)
	}

	time.Sleep(1 * time.Second)
	if err := q.setVNCPassword(vmid, config.Password); err != nil {
		slog.Warn("failed to set VNC password", "error", err)
	}

	go func() {
		cmd.Wait()
	}()

	fmt.Printf("log file: %s\n", logFilePath)

	return pid, nil
}

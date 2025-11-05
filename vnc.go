package goQemu

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (q *Qemu) OpenVNC(vmid int) error {
	config, err := q.loadConfig(vmid)
	if err != nil {
		return fmt.Errorf("failed to get VM %d config: %w", vmid, err)
	}

	if config.VNCPort == 0 {
		return fmt.Errorf("VNC is not enabled for VM %d", vmid)
	}

	pidFile := filepath.Join(q.Folder.PID, fmt.Sprintf("%d.pid", vmid))
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("VM %d is not running", vmid)
	}

	var pid int
	fmt.Sscanf(string(pidData), "%d", &pid)
	if !q.isRunning(pid) {
		return fmt.Errorf("VM %d is not running", vmid)
	}

	vncURL := fmt.Sprintf("vnc://localhost:%d", config.VNCPort)
	fmt.Printf("connection to VM vnc %d on port %d...\n", vmid, config.VNCPort)

	cmd := exec.Command("open", vncURL)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open VNC viewer: %v\n", err)
		return nil
	}

	fmt.Printf("VNC viewer opened. Connect to: %s\n", vncURL)
	return nil
}

func (q *Qemu) setVNCPassword(vmid int, password string) error {
	monitorPath := filepath.Join(q.Folder.Monitor, fmt.Sprintf("%d.sock", vmid))

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(monitorPath); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	conn, err := net.Dial("unix", monitorPath)
	if err != nil {
		return fmt.Errorf("failed to connect to monitor: %w", err)
	}
	// defer conn.Close()

	buf := make([]byte, 1024)
	conn.Read(buf)

	cmd := fmt.Sprintf("change vnc password %s\n", password)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	conn.Read(buf)

	return nil
}

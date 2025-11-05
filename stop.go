package goQemu

import (
	"fmt"
	"os"
	"syscall"
)

func (q *Qemu) Stop(vmid int) error {
	_, err := q.loadConfig(vmid)
	if err != nil {
		return fmt.Errorf("failed to get VM %d config: %w", vmid, err)
	}

	var pid int
	pidFilepath, pidBody, err := q.getFile(q.Folder.PID, vmid)
	if err == nil {
		fmt.Sscanf(pidBody, "%d", &pid)
	}
	if !q.isRunning(pid) {
		return fmt.Errorf("VM %d is not running", vmid)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	os.Remove(pidFilepath)

	q.Cleanup()

	return nil
}

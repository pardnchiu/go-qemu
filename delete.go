package goQemu

import (
	"log/slog"
	"os"
	"time"
)

func (q *Qemu) Delete(vmid int) error {
	_, err := q.loadConfig(vmid)
	if err != nil {
		slog.Error("Failed to get VM config", "vmid", vmid, "error", err)
	}

	if pidFilePath, _, err := q.getFile(q.Folder.PID, vmid); err == nil {
		q.Stop(vmid)
		os.Remove(pidFilePath)
	}

	if configPath, _, err := q.getFile(q.Folder.Config, vmid); err == nil {
		os.Remove(configPath)
	}

	if diskPaths, err := q.diskPathAll(vmid); err == nil {
		for _, path := range diskPaths {
			slog.Info("Deleting disk file", "vmid", vmid, "path", path)
			os.Remove(path)
		}
		q.removeCloudInit(vmid)
	}

	if logPath, _, err := q.getFile(q.Folder.Log, vmid); err == nil {
		os.Remove(logPath)
	}

	time.Sleep(1 * time.Second)

	q.Cleanup()

	return nil
}

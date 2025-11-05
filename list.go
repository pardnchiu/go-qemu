package goQemu

import (
	"fmt"
	"os"
)

func (q *Qemu) List() []*Instance {
	vms := make([]*Instance, 0)
	ids, err := os.ReadDir(q.Folder.Config)
	if err != nil {
		return vms
	}

	for _, id := range ids {
		if id.IsDir() {
			continue
		}

		var vmid int
		if _, err := fmt.Sscanf(id.Name(), "%d.json", &vmid); err != nil {
			continue
		}

		config, err := q.loadConfig(vmid)
		if err != nil {
			continue
		}

		instance := &Instance{
			Config: *config,
			Status: "stopped",
		}

		if _, pidData, err := q.getFile(q.Folder.PID, vmid); err == nil {
			var pid int
			fmt.Sscanf(pidData, "%d", &pid)
			instance.PID = pid

			if q.isRunning(pid) {
				instance.Status = "running"
			}
		}

		vms = append(vms, instance)
	}

	return vms
}

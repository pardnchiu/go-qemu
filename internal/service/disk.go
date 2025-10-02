package service

import (
	"fmt"
	"strconv"
)

func (s *Service) Disk(vmid int, disk string) error {
	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"disk", "resize", strconv.Itoa(vmid),
		"scsi0", fmt.Sprintf("+%s", disk),
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("failed to append disk: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

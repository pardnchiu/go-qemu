package service

import (
	"fmt"
	"strconv"
)

func (s *Service) Destroy(vmid int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"destroy", strconv.Itoa(vmid),
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("failed to destroy VM: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

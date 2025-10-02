package service

import (
	"fmt"
	"strconv"
)

func (s *Service) Stop(vmid int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"stop", strconv.Itoa(vmid),
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("[-] failed to stop VM: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

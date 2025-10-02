package service

import (
	"fmt"
	"strconv"
)

func (s *Service) CPU(vmid, cpu int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"set", strconv.Itoa(vmid),
		"--cores", strconv.Itoa(cpu),
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("[-] failed to set CPU: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

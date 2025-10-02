package service

import (
	"fmt"
	"os"
	"strconv"
)

func (s *Service) Memory(vmid, memory int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"set", strconv.Itoa(vmid),
		"--memory", strconv.Itoa(memory),
	}

	envBalloonMin := os.Getenv("VM_BALLOON_MIN")
	balloonMin, err := strconv.Atoi(envBalloonMin)
	if err != nil {
		balloonMin = 0
	}

	if balloonMin != 0 && memory >= balloonMin+1024 {

		sharingMemory := memory - balloonMin
		if memory >= 65536 {
			sharingMemory = 49152
		}

		cmdArgs = append(cmdArgs, "--numa", "1")
		cmdArgs = append(cmdArgs, "--balloon", strconv.Itoa(sharingMemory))
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("[-] failed to set memory: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

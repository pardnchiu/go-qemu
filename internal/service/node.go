package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Service) Node(c *gin.Context, vmid int, node string) error {
	origin := c.Request.Header.Get("Origin")
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	isMain, _, ip := s.getVMIDsNode(vmid)

	cmdArgs := []string{
		"migrate", strconv.Itoa(vmid), node,
		"--with-local-disks",
	}

	cmd := s.getCommand(isMain, ip, cmdArgs...)

	if err := s.runCommandSSE(c, cmd, "migrating", "processing"); err != nil {
		return err
	}

	return nil
}

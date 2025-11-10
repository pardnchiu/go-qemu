package goQemu

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (q *Qemu) OpenVNC(vmid int) error {
	config, err := q.loadConfig(vmid)
	if err != nil {
		return fmt.Errorf("failed to get VM (%d): %w", vmid, err)
	}

	if config.VNCPort == 0 {
		return fmt.Errorf("VM (%d) is not enabled", vmid)
	}

	var pid int
	if _, data, err := q.getFile(q.Folder.PID, vmid); err == nil {
		fmt.Sscanf(data, "%d", &pid)
		if !q.isRunning(pid) {
			return fmt.Errorf("VM (%d) is not running", vmid)
		}
	} else {
		return fmt.Errorf("VM (%d) is not running", vmid)
	}

	ip, err := q.getHostIP()
	if err != nil {
		ip = "localhost"
	}

	fmt.Printf("vnc://%s:%d\n", ip, config.VNCPort)

	return nil
}

func (q *Qemu) getHostIP() (string, error) {
	cmd := exec.Command("ip", "addr", "show", "vmbr0")
	output, err := cmd.Output()
	if err == nil {
		/*
					3: vmbr0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
			    link/ether ae:68:b4:7e:23:cd brd ff:ff:ff:ff:ff:ff
			    inet 10.7.22.180/24 scope global vmbr0
			       valid_lft forever preferred_lft forever
			    inet6 fe80::ac68:b4ff:fe7e:23cd/64 scope link
			       valid_lft forever preferred_lft forever
		*/
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "inet ") {
				// inet 10.7.22.180/24 scope global vmbr0
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					ipCIDR := parts[1]
					ip := strings.Split(ipCIDR, "/")[0]
					return ip, nil
				}
			}
		}
	}

	cmd = exec.Command("ip", "route", "get", "1.1.1.1")
	output, err = cmd.Output()
	if err == nil {
		/*
					1.1.1.1 via 10.7.22.1 dev vmbr0 src 10.7.22.180 uid 1000
			    cache
		*/
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "src") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "src" && i+1 < len(parts) {
						return parts[i+1], nil
					}
				}
			}
		}
	}

	cmd = exec.Command("hostname", "-I")
	output, err = cmd.Output()
	if err == nil {
		ips := strings.Fields(string(output))
		for _, ip := range ips {
			if !strings.HasPrefix(ip, "127.") && strings.Count(ip, ".") == 3 {
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("could not determine host IP")
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

	buf := make([]byte, 4096)
	_, err = conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read from monitor: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	cmd := fmt.Sprintf("change vnc password %s\n", password)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	response := ""
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read monitor response: %w", err)
		}
		response += string(buf[:n])
		if strings.Contains(response, "(qemu)") {
			break
		}
	}

	if !isSuccess(response) {
		return fmt.Errorf("failed to set VNC password, monitor response: %s", response)
	}

	return nil
}

func isSuccess(resp string) bool {
	return !(strings.Contains(resp, "error") || strings.Contains(resp, "failed"))
}

package goQemu

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// TODO: append ssh string for cloud-init config
func (q *Qemu) Create(config Config, ssh string) error {
	q.Cleanup()

	// assign VMID if not provided
	if config.ID == 0 {
		vmid, err := q.assignVMID()
		if err != nil {
			return fmt.Errorf("failed to assign VMID: %w", err)
		}
		config.ID = vmid
	}

	// check if VMID already exists
	_, configBody, err := q.getFile(q.Folder.Config, config.ID)
	if err == nil && configBody != "" {
		return fmt.Errorf("VMID %d already exists", config.ID)
	}

	if config.OS != "" && config.Version != "" {
		if config.Hostname == "" {
			config.Hostname = fmt.Sprintf("%s-%d.vm", config.OS, config.ID)
		}

		img, err := q.getOSImageInfo(config.OS, config.Version)
		if err != nil {
			return fmt.Errorf("failed to get OS image info: %w", err)
		}

		imagePath, err := q.downloadOSImage(img)
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}

		diskSize := config.DiskSize
		if diskSize == "" {
			diskSize = "16G"
		}

		diskPath, err := q.generateVMDisk(config.ID, imagePath, diskSize)
		if err != nil {
			return fmt.Errorf("failed to generate VM disk: %w", err)
		}

		config.DiskPath = diskPath
	} else {
		return fmt.Errorf("either disk_path or (os and version) must be specified")
	}

	username := config.OS
	if config.OS == "rockylinux" {
		username = "rocky"
	} else if config.OS == "almalinux" {
		username = "alma"
	}

	passwd := "passwd"

	if config.Options.UUID == "" {
		config.Options = Options{
			UUID: uuid.New().String(),
		}

		config.CloudInit = CloudInit{
			Hostname:        config.Hostname,
			Username:        username,
			Password:        passwd,
			AuthorizedKey:   ssh,
			UpgradePackages: true,
			IPv4:            "mode=dhcp,address=,gateway=",
			IPv6:            "mode=dhcp,address=,gateway=",
		}
	}

	verifyConfig, err := q.verifyConfig(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := q.saveConfig(*verifyConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	pid, err := q.runVM(verifyConfig, verifyConfig.ID)
	if err != nil {
		return err
	}

	fmt.Printf("[*] VM %d created with PID %d\n", verifyConfig.ID, pid)
	return nil
}

func (q *Qemu) verifyArgs(config Config) []string {
	vncDisplay := config.VNCPort - 5900
	monitorPath := filepath.Join(q.Folder.Monitor, fmt.Sprintf("%d.sock", config.ID))

	machineType := runtime.GOARCH
	switch runtime.GOARCH {
	case "arm64", "arm":
		machineType = "virt"
	default:
		machineType = "pc"
	}

	// seabios / ovmf(uefi)
	biosPath := "/usr/share/seabios/bios.bin"
	if config.BIOS == "ovmf" || config.BIOS == "OVMF" {
		for _, e := range []string{
			"/usr/share/OVMF/OVMF_CODE.fd",
			"/usr/share/OVMF/OVMF.fd",
			"/usr/share/ovmf/OVMF_CODE.fd",
			"/usr/share/ovmf/OVMF.fd",
		} {
			if _, err := os.Stat(e); err == nil {
				biosPath = e
				break
			}
		}
	}

	// * for apple silicon
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		biosPath = "/opt/homebrew/share/qemu/edk2-aarch64-code.fd"
	}

	args := []string{
		"-accel", config.Accelerator,
		"-m", fmt.Sprintf("%d", config.Memory),
		"-smp", fmt.Sprintf("%d,sockets=%d,cores=%d,threads=1", config.CPUs, 1, config.CPUs),
		"-cpu", "host",
		"-M", machineType,
		"-bios", biosPath,
		"-device", "qemu-xhci",
		"-device", "usb-kbd",
		"-device", "usb-tablet",
		"-audiodev", "none,id=audio0",
		"-device", "intel-hda",
		"-device", "hda-duplex,audiodev=audio0",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", config.DiskPath),
		"-drive", fmt.Sprintf("file=%s,format=raw,media=cdrom,readonly=on", config.CloudInitPath),
		"-rtc", "base=utc,clock=host",
		"-vnc", fmt.Sprintf("0.0.0.0:%d,password=on", vncDisplay),
		"-monitor", fmt.Sprintf("unix:%s,server,nowait", monitorPath),
		// "-chardev", fmt.Sprintf("socket,id=mon0,path=%s,server=on,wait=off", monitorPath),
		// "-mon", "chardev=mon0,mode=control",
		// "-netdev", fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", config.SSHPort),
		// "-device", "virtio-net-pci,netdev=net0",

		"-smbios", fmt.Sprintf("type=1,uuid=%s", config.Options.UUID),
		"-device", "virtio-gpu-pci",
		// "-display", "none",
		// "-nographic",
		// "-display", "cocoa,show-cursor=on",
		// "-serial", "null", // Disable serial console
	}

	for i, e := range config.Network {
		net := getNetwork(e)
		slog.Info("Parsed network", "value", e, "network", net)
		if net.Disconnect {
			continue
		}

		netdevID := fmt.Sprintf("net%d", i)

		netdevArgs := fmt.Sprintf("bridge,id=%s,br=%s", netdevID, net.Bridge)
		args = append(args, "-netdev", netdevArgs)

		deviceArgs := fmt.Sprintf("%s,netdev=%s", net.Model, netdevID)

		if net.MACAddress != "" {
			deviceArgs += fmt.Sprintf(",mac=%s", net.MACAddress)
		}

		if net.Multiqueue > 0 {
			deviceArgs += fmt.Sprintf(",mq=on,vectors=%d", net.Multiqueue*2+2)
		}

		args = append(args, "-device", deviceArgs)
	}

	slog.Info("config", "arg", args)

	return args
}

func (q *Qemu) assignVMID() (int, error) {
	ids, err := os.ReadDir(q.Folder.Config)
	if err != nil {
		return 0, fmt.Errorf("failed to read go-qemu/configs: %w", err)
	}

	ary := make(map[int]bool, len(ids))
	for _, id := range ids {
		if id.IsDir() {
			continue
		}

		var vmid int
		if _, err := fmt.Sscanf(id.Name(), "%d.json", &vmid); err == nil {
			ary[vmid] = true
		}
	}

	for id := 100; id <= 999; id++ {
		if !ary[id] {
			return id, nil
		}
	}

	return 0, fmt.Errorf("no available VMID can be assigned")
}

func getNetwork(value string) Network {
	network := Network{MTU: 1500}

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key, val := kv[0], kv[1]
		switch key {
		case "bridge":
			network.Bridge = val
		case "model":
			network.Model = val
		case "vlan":
			network.Vlan, _ = strconv.Atoi(val)
		case "mac_address":
			network.MACAddress = val
		case "firewall":
			network.Firewall = val != "0"
		case "disconnect":
			network.Disconnect = val != "0"
		case "mtu":
			if mtu, err := strconv.Atoi(val); err == nil && mtu > 0 {
				network.MTU = mtu
			}
		case "rate_limit":
			network.RateLimit, _ = strconv.Atoi(val)
		case "multiqueue":
			network.Multiqueue, _ = strconv.Atoi(val)
		}
	}

	return network
}

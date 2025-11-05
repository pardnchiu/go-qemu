package goQemu

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/google/uuid"
)

func (q *Qemu) Create(config Config) error {
	q.Cleanup()

	config.UUID = uuid.New().String()

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

	config.Username = config.OS

	if config.OS == "rockylinux" {
		config.Username = "rocky"
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

	args := []string{
		"-accel", config.Accelerator,
		"-m", fmt.Sprintf("%d", config.Memory),
		"-smp", fmt.Sprintf("%d,sockets=%d,cores=%d,threads=1", config.CPUs, 1, config.CPUs),
		"-cpu", "host",
		"-M", "virt",
		"-bios", config.BIOSPath,
		"-device", "qemu-xhci",
		"-device", "usb-kbd",
		"-device", "usb-tablet",
		"-audiodev", "none,id=audio0",
		"-device", "intel-hda",
		"-device", "hda-duplex,audiodev=audio0",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", config.DiskPath),
		"-drive", fmt.Sprintf("file=%s,format=raw,media=cdrom,readonly=on", config.CloudInit),
		"-rtc", "base=utc,clock=host",
		"-vnc", fmt.Sprintf("127.0.0.1:%d,password=on", vncDisplay),
		"-monitor", fmt.Sprintf("unix:%s,server,nowait", monitorPath),
		// "-chardev", fmt.Sprintf("socket,id=mon0,path=%s,server=on,wait=off", monitorPath),
		// "-mon", "chardev=mon0,mode=control",
		"-netdev", fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", config.SSHPort),
		"-device", "virtio-net-pci,netdev=net0",

		"-smbios", fmt.Sprintf("type=1,uuid=%s", config.UUID),
		"-device", "virtio-gpu-pci",
		// "-display", "none",
		// "-nographic",
		// "-display", "cocoa,show-cursor=on",
		// "-serial", "null", // Disable serial console
	}

	slog.Info("config", "arg", args)

	return args
}

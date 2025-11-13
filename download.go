package goQemu

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func isExists(ary []string, item string) bool {
	for _, v := range ary {
		if v == item {
			return true
		}
	}
	return false
}

func (q *Qemu) getOSImageInfo(osName, version string) (*Image, error) {
	var (
		debianVersions = strings.Split(os.Getenv("GO_QEMU_DEBIAN_VERSION"), ",")
		ubuntuVersions = strings.Split(os.Getenv("GO_QEMU_UBUNTU_VERSION"), ",")
		centosVersions = strings.Split(os.Getenv("GO_QEMU_CENTOS_VERSION"), ",")
		rockyVersions  = strings.Split(os.Getenv("GO_QEMU_ROCKYLINUX_VERSION"), ",")
		almaVersions   = strings.Split(os.Getenv("GO_QEMU_ALMALINUX_VERSION"), ",")
		img            Image
	)
	img.OS = osName
	img.Version = version

	arch := runtime.GOARCH
	switch arch {
	case "amd64", "arm64":
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}

	switch osName {
	case "debian":
		versionNames := map[string]string{
			"11": "bullseye",
			"12": "bookworm",
			"13": "trixie",
		}
		if !isExists(debianVersions, version) {
			return nil, fmt.Errorf("unsupported version: Debian %s", version)
		}
		versionName := versionNames[version]
		img.URL = fmt.Sprintf("https://cloud.debian.org/images/cloud/%s/latest/debian-%s-generic-%s.qcow2", versionName, version, arch)
		img.Filename = fmt.Sprintf("debian-%s-generic-%s.qcow2", version, arch)
	case "ubuntu":
		if !isExists(ubuntuVersions, version) {
			return nil, fmt.Errorf("unsupported version: Ubuntu %s", version)
		}
		img.URL = fmt.Sprintf("https://cloud-images.ubuntu.com/releases/%s/release/ubuntu-%s-server-cloudimg-%s.img", version, version, arch)
		img.Filename = fmt.Sprintf("ubuntu-%s-server-cloudimg-%s.img", version, arch)
	case "centos":
		switch arch {
		case "arm64":
			arch = "aarch64"
		case "amd64":
			arch = "x86_64"
		}
		if !isExists(centosVersions, version) {
			return nil, fmt.Errorf("unsupported version: CentOS %s", version)
		}
		img.URL = fmt.Sprintf("https://cloud.centos.org/centos/%s-stream/%s/images/CentOS-Stream-GenericCloud-%s-latest.%s.qcow2", version, arch, version, arch)
		img.Filename = fmt.Sprintf("CentOS-Stream-GenericCloud-%s-latest.%s.qcow2", version, arch)
	case "rockylinux":
		switch arch {
		case "arm64":
			arch = "aarch64"
		case "amd64":
			arch = "x86_64"
		}
		if !isExists(rockyVersions, version) {
			return nil, fmt.Errorf("unsupported version: RockyLinux %s", version)
		}
		img.URL = fmt.Sprintf("https://dl.rockylinux.org/pub/rocky/%s/images/%s/Rocky-%s-GenericCloud-Base.latest.%s.qcow2", version, arch, version, arch)
		img.Filename = fmt.Sprintf("Rocky-%s-GenericCloud-Base.latest.%s.qcow2", version, arch)
	case "almalinux":
		switch arch {
		case "arm64":
			arch = "aarch64"
		case "amd64":
			arch = "x86_64"
		}
		if !isExists(almaVersions, version) {
			return nil, fmt.Errorf("unsupported version: AlmaLinux %s", version)
		}
		img.URL = fmt.Sprintf("https://repo.almalinux.org/almalinux/%s/cloud/%s/images/AlmaLinux-%s-GenericCloud-latest.%s.qcow2", version, arch, version, arch)
		img.Filename = fmt.Sprintf("AlmaLinux-%s-GenericCloud-latest.%s.qcow2", version, arch)
	default:
		return nil, fmt.Errorf("unsupported OS: %s", osName)
	}

	return &img, nil
}

func (q *Qemu) downloadOSImage(image *Image) (string, error) {
	imagePath := filepath.Join(q.Folder.Image, image.Filename)
	if _, err := os.Stat(imagePath); err == nil {
		// * image exists
		return imagePath, nil
	}

	fmt.Printf("[*] download: %s (%s)\n", image.OS, image.Version)
	resp, err := http.Get(image.URL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// * HTTP error
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	size := resp.ContentLength
	tmpFile := imagePath + ".tmp"
	imageFile, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to create: %w", err)
	}
	defer imageFile.Close()

	Progress := &Progress{
		Total:     size,
		Completed: 0,
	}

	reader := io.TeeReader(resp.Body, Progress)

	if size > 0 {
		fmt.Printf("[*] total size: %.2f MB\n", float64(size)/1024/1024)
	}

	_, err = io.Copy(imageFile, reader)
	if err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := os.Rename(tmpFile, imagePath); err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Printf("\n[*] filePath: %s\n", imagePath)
	return imagePath, nil
}

func (q *Qemu) generateVMDisk(vmid int, imagePath, size string) (string, error) {
	ext := filepath.Ext(imagePath)
	if ext == "" {
		return "", fmt.Errorf("invalid image file")
	}

	fmt.Printf("[*] copying image to VM directory\n")
	target := fmt.Sprintf("%d-0%s", vmid, ext)
	targetPath := filepath.Join(q.Folder.VM, target)
	if err := copy(imagePath, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	if size == "" {
		return "", fmt.Errorf("disk size is required")
	}

	fmt.Printf("\n[*] resizing disk to %s\n", size)
	cmd := exec.Command("qemu-img", "resize", targetPath, size)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to resize: %w", err)
	}

	return targetPath, nil
}

func copy(imagePath, targetPath string) error {
	fromPath, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer fromPath.Close()

	toPath, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer toPath.Close()

	imageInfo, err := fromPath.Stat()
	if err != nil {
		return fmt.Errorf("failed to get image info: %w", err)
	}
	size := imageInfo.Size()

	progress := &Progress{
		Total:     size,
		Completed: 0,
	}

	if size > 0 {
		fmt.Printf("[*] total size: %.2f MB\n", float64(size)/1024/1024)
	}

	reader := io.TeeReader(fromPath, progress)

	_, err = io.Copy(toPath, reader)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return toPath.Sync()
}

// * for io.TeeReader to track download progress
func (d *Progress) Write(progress []byte) (int, error) {
	bytes := len(progress)
	d.Completed += int64(bytes)

	if d.Total > 0 {
		percent := float64(d.Completed) / float64(d.Total) * 100
		completed := float64(d.Completed) / 1024 / 1024
		total := float64(d.Total) / 1024 / 1024

		fmt.Printf("\r[*] progress: %.2f MB / %.2f MB (%.2f%%)", completed, total, percent)
	} else {
		completed := float64(d.Completed) / 1024 / 1024
		fmt.Printf("\r[*] progress: %.2f MB", completed)
	}

	return bytes, nil
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func (m *Folder) getOSImageInfo(osName, version string) (*Image, error) {
	var (
		debianVersions = strings.Split(os.Getenv("GO_QEMU_DEBIAN_VERSION"), ",")
		ubuntuVersions = strings.Split(os.Getenv("GO_QEMU_UBUNTU_VERSION"), ",")
		rockyVersions  = strings.Split(os.Getenv("GO_QEMU_ROCKYLINUX_VERSION"), ",")
		img            Image
	)
	img.OS = osName
	img.Version = version

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
		img.URL = fmt.Sprintf("https://cloud.debian.org/images/cloud/%s/latest/debian-%s-generic-arm64.qcow2", versionName, version)
		img.Filename = fmt.Sprintf("debian-%s-generic-arm64.qcow2", version)
	case "ubuntu":
		if !isExists(ubuntuVersions, version) {
			return nil, fmt.Errorf("unsupported version: Ubuntu %s", version)
		}
		img.URL = fmt.Sprintf("https://cloud-images.ubuntu.com/releases/%s/release/ubuntu-%s-server-cloudimg-arm64.img", version, version)
		img.Filename = fmt.Sprintf("ubuntu-%s-server-cloudimg-arm64.img", version)
	case "rockylinux":
		if !isExists(rockyVersions, version) {
			return nil, fmt.Errorf("unsupported version: RockyLinux %s", version)
		}
		img.URL = fmt.Sprintf("https://dl.rockylinux.org/pub/rocky/%s/images/aarch64/Rocky-%s-GenericCloud-Base.latest.aarch64.qcow2", version, version)
		img.Filename = fmt.Sprintf("Rocky-%s-GenericCloud-Base.latest.aarch64.qcow2", version)
	default:
		return nil, fmt.Errorf("unsupported OS: %s", osName)
	}

	return &img, nil
}

func (m *Folder) downloadOSImage(image *Image) (string, error) {
	imagePath := filepath.Join(m.Image, image.Filename)
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

	fmt.Printf("[*] filePath: %s\n", imagePath)
	return imagePath, nil
}

func (m *Folder) generateVMDisk(vmid int, imagePath, size string) (string, error) {
	ext := filepath.Ext(imagePath)
	if ext == "" {
		return "", fmt.Errorf("invalid image file")
	}

	fmt.Printf("[*] copying image to VM directory\n")
	target := fmt.Sprintf("%d-0%s", vmid, ext)
	targetPath := filepath.Join(m.Main, target)
	if err := copy(imagePath, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	if size == "" {
		return "", fmt.Errorf("disk size is required")
	}

	fmt.Printf("[*] resizing disk to %s\n", size)
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

		fmt.Printf("\r[*] progress: %.2f MB / %.2f MB (%.2f%%)\n", completed, total, percent)
	} else {
		completed := float64(d.Completed) / 1024 / 1024
		fmt.Printf("\r[*] progress: %.2f MB\n", completed)
	}

	return bytes, nil
}

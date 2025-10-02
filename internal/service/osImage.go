package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func (s *Service) getOSImage(os, version string) (string, string, error) {
	var url, file string
	debianSupported := []string{"11", "12", "13"}
	debianOSName := map[string]string{
		"11": "bullseye",
		"12": "bookworm",
		"13": "trixie",
	}
	rockySupported := []string{"8", "9", "10"}
	ubuntuSupported := []string{"20.04", "22.04", "24.04"}

	switch os {
	case "debian":
		if !slices.Contains(debianSupported, version) {
			return "", "", fmt.Errorf("unsupported Debian version: %s", version)
		}
		url = fmt.Sprintf("https://cloud.debian.org/images/cloud/%s/latest/debian-%s-generic-amd64.qcow2", debianOSName[version], version)
		file = fmt.Sprintf("/tmp/debian-%s-generic-amd64.qcow2", version)
	case "rockylinux":
		if !slices.Contains(rockySupported, version) {
			return "", "", fmt.Errorf("unsupported RockyLinux version: %s", version)
		}
		url = fmt.Sprintf("https://dl.rockylinux.org/pub/rocky/%s/images/x86_64/Rocky-%s-GenericCloud-Base.latest.x86_64.qcow2", version, version)
		file = fmt.Sprintf("/tmp/Rocky-%s-GenericCloud-Base.latest.x86_64.qcow2", version)
	case "ubuntu":
		if !slices.Contains(ubuntuSupported, version) {
			return "", "", fmt.Errorf("unsupported Ubuntu Server version: %s", version)
		}
		url = fmt.Sprintf("https://mirror.twds.com.tw/ubuntu-cloud-images/releases/%s/release/ubuntu-%s-server-cloudimg-amd64.img", version, version)
		file = fmt.Sprintf("/tmp/ubuntu-%s-server-cloudimg-amd64.img", version)
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", os)
	}

	return url, file, nil
}

func (s *Service) checkOSImageURL(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to access URL: %s", resp.Status)
	}

	return nil
}

func (s *Service) downloadOSImage(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// generate by ChatGPT
package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCloudInit(t *testing.T) {
	os.Setenv("TEST_MODE", "true")
	defer os.Unsetenv("TEST_MODE")

	// Setup test folder
	testFolder := Folder{VM: "./testdata"}
	if err := os.MkdirAll(testFolder.VM, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}
	defer os.RemoveAll(testFolder.VM)

	// Test configuration
	config := CloudInit{
		UUID:             "test-uuid",
		OS:               "debian",
		Hostname:         "test-host",
		Username:         "test-user",
		Password:         "test-pass",
		SSHAuthorizedKey: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA7...",
	}

	// Test NewCloudInit
	isoPath, err := testFolder.NewCloudInit(101, config)
	if err != nil {
		t.Fatalf("NewCloudInit failed: %v", err)
	}

	// Verify ISO file creation
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		t.Errorf("Expected ISO file not found: %s", isoPath)
	}

	// Verify meta-data file
	metaDataPath := filepath.Join(testFolder.VM, ".cloudinit-101", "meta-data")
	if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
		t.Errorf("Expected meta-data file not found: %s", metaDataPath)
	}

	// Verify user-data file
	userDataPath := filepath.Join(testFolder.VM, ".cloudinit-101", "user-data")
	if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
		t.Errorf("Expected user-data file not found: %s", userDataPath)
	}

	slog.Info("TestNewCloudInit completed successfully", slog.String("ISOPath", isoPath))
}

func TestGetOSImageInfo(t *testing.T) {
	// Setup environment variables
	os.Setenv("GO_QEMU_DEBIAN_VERSION", "11,12,13")
	os.Setenv("GO_QEMU_UBUNTU_VERSION", "20.04,22.04")
	os.Setenv("GO_QEMU_ROCKYLINUX_VERSION", "8,9")
	defer func() {
		os.Unsetenv("GO_QEMU_DEBIAN_VERSION")
		os.Unsetenv("GO_QEMU_UBUNTU_VERSION")
		os.Unsetenv("GO_QEMU_ROCKYLINUX_VERSION")
	}()

	m := &Folder{}

	tests := []struct {
		name      string
		osName    string
		version   string
		expectErr bool
	}{
		{"Valid Debian Version", "debian", "11", false},
		{"Invalid Debian Version", "debian", "10", true},
		{"Valid Ubuntu Version", "ubuntu", "22.04", false},
		{"Invalid Ubuntu Version", "ubuntu", "18.04", true},
		{"Valid RockyLinux Version", "rockylinux", "8", false},
		{"Invalid RockyLinux Version", "rockylinux", "7", true},
		{"Unsupported OS", "fedora", "35", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := m.getOSImageInfo(tt.osName, tt.version)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for OS: %s, Version: %s, but got none", tt.osName, tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for OS: %s, Version: %s: %v", tt.osName, tt.version, err)
				}
				if img == nil {
					t.Errorf("Expected valid Image struct, got nil")
				} else {
					if img.OS != tt.osName || img.Version != tt.version {
						t.Errorf("Image struct mismatch. Got OS: %s, Version: %s, Expected OS: %s, Version: %s",
							img.OS, img.Version, tt.osName, tt.version)
					}
				}
			}

			slog.Info("Test completed", slog.String("OS", tt.osName), slog.String("Version", tt.version), slog.Bool("ExpectErr", tt.expectErr))
		})
	}
}

func TestDownloadOSImage(t *testing.T) {
	// Setup temporary folder for testing
	tempDir := t.TempDir()
	m := &Folder{Image: tempDir}

	tests := []struct {
		name         string
		img          *Image
		setup        func() *httptest.Server
		expectErr    bool
		expectedFile string
	}{
		{
			name: "Image already exists",
			img: &Image{
				OS:       "debian",
				Version:  "11",
				Filename: "debian-11-generic-arm64.qcow2",
			},
			setup: func() *httptest.Server {
				existingFile := filepath.Join(tempDir, "debian-11-generic-arm64.qcow2")
				os.WriteFile(existingFile, []byte{}, 0644)
				return nil
			},
			expectErr:    false,
			expectedFile: "debian-11-generic-arm64.qcow2",
		},
		{
			name: "Successful download",
			img: &Image{
				OS:       "ubuntu",
				Version:  "22.04",
				Filename: "ubuntu-22.04-server-cloudimg-arm64.img",
			},
			setup: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("dummy image data"))
				}))
				return server
			},
			expectErr:    false,
			expectedFile: "ubuntu-22.04-server-cloudimg-arm64.img",
		},
		{
			name: "HTTP error during download",
			img: &Image{
				OS:       "rockylinux",
				Version:  "8",
				Filename: "Rocky-8-GenericCloud-Base.latest.aarch64.qcow2",
			},
			setup: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				return server
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setup != nil {
				server = tt.setup()
				if server != nil {
					defer server.Close()
					tt.img.URL = server.URL
				}
			}

			result, err := m.downloadOSImage(tt.img)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				expectedPath := filepath.Join(tempDir, tt.expectedFile)
				if result != expectedPath {
					t.Errorf("Expected file path %s, got %s", expectedPath, result)
				}
			}
		})
	}
}

func TestGenerateVMDisk(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_generate_vmdisk")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary source file
	srcFile, err := os.CreateTemp(tempDir, "source-*.img")
	if err != nil {
		t.Fatalf("Failed to create temp source file: %v", err)
	}
	defer srcFile.Close()

	// Write some data to the source file
	if _, err := srcFile.Write([]byte("test data")); err != nil {
		t.Fatalf("Failed to write to source file: %v", err)
	}

	// Create a Folder struct with the temporary directory as the main directory
	folder := &Folder{
		VM: tempDir,
	}

	// Call generateVMDisk
	vmid := 1
	imagePath := srcFile.Name()
	size := "20M"
	targetPath, err := folder.generateVMDisk(vmid, imagePath, size)
	if err != nil {
		t.Fatalf("generateVMDisk failed: %v", err)
	}

	// Check if the target file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatalf("Target file does not exist: %s", targetPath)
	}

	// Check if the file was resized (this is a basic check, actual resizing depends on qemu-img)
	if filepath.Ext(targetPath) != ".img" {
		t.Errorf("Unexpected file extension: %s", filepath.Ext(targetPath))
	}
}

func TestNewVMManager(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir := t.TempDir()
	os.Setenv("GO_QEMU_PATH", tempDir)
	defer os.Unsetenv("GO_QEMU_PATH")

	// Call NewVMManager
	folder, err := NewQemu()
	if err != nil {
		t.Fatalf("NewVMManager failed: %v", err)
	}

	// Verify that the folder paths are correctly set
	expectedPaths := []struct {
		name string
		path string
	}{
		{"VM", folder.VM},
		{"Config", folder.Config},
		{"Log", folder.Log},
		{"PID", folder.PID},
		{"Monitor", folder.Monitor},
		{"Image", folder.Image},
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(p.path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s not found: %s", p.name, p.path)
		}
	}
}

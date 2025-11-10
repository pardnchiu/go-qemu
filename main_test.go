// generate by ChatGPT
package goQemu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// func TestNewCloudInit(t *testing.T) {
// 	os.Setenv("TEST_MODE", "true")
// 	defer os.Unsetenv("TEST_MODE")

// 	// Setup test folder
// 	testFolder := Qemu{Folder: Folder{VM: "./testdata"}}
// 	if err := os.MkdirAll(testFolder.Folder.VM, 0755); err != nil {
// 		t.Fatalf("Failed to create test folder: %v", err)
// 	}
// 	defer os.RemoveAll(testFolder.Folder.VM)

// 	// Test configuration
// 	config := CloudInit{
// 		UUID:             "test-uuid",
// 		// OS:               "debian",
// 		Hostname:         "test-host",
// 		Username:         "test-user",
// 		Password:         "test-pass",
// 		SSHAuthorizedKey: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA7...",
// 	}

// 	// Test NewCloudInit
// 	isoPath, err := testFolder.generateCloudInit(101, config)
// 	if err != nil {
// 		t.Fatalf("NewCloudInit failed: %v", err)
// 	}

// 	// Verify ISO file creation
// 	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
// 		t.Errorf("Expected ISO file not found: %s", isoPath)
// 	}

// 	// Verify meta-data file
// 	metaDataPath := filepath.Join(testFolder.Folder.VM, ".cloudinit-101", "meta-data")
// 	if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
// 		t.Errorf("Expected meta-data file not found: %s", metaDataPath)
// 	}

// 	// Verify user-data file
// 	userDataPath := filepath.Join(testFolder.Folder.VM, ".cloudinit-101", "user-data")
// 	if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
// 		t.Errorf("Expected user-data file not found: %s", userDataPath)
// 	}

// 	slog.Info("TestNewCloudInit completed successfully", slog.String("ISOPath", isoPath))
// }

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

	m := &Qemu{}

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
	m := &Qemu{Folder: Folder{Image: tempDir}}

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
	folder := &Qemu{
		Folder: Folder{
			VM: tempDir,
		},
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

func TestNewQemu(t *testing.T) {
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
		{"VM", folder.Folder.VM},
		{"Config", folder.Folder.Config},
		{"Log", folder.Folder.Log},
		{"PID", folder.Folder.PID},
		{"Monitor", folder.Folder.Monitor},
		{"Image", folder.Folder.Image},
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(p.path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s not found: %s", p.name, p.path)
		}
	}
}

func TestAssignVMID(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a Folder struct with the temporary directory as the Config path
	folder := &Qemu{
		Folder: Folder{
			Config: tempDir,
		},
	}

	// Test case: No existing VM IDs
	t.Run("No existing VM IDs", func(t *testing.T) {
		vmid, err := folder.assignVMID()
		if err != nil {
			t.Fatalf("generateVMID failed: %v", err)
		}
		if vmid < 100 || vmid > 999 {
			t.Errorf("Generated VM ID out of range: %d", vmid)
		}
	})

	// Test case: Some existing VM IDs
	t.Run("Some existing VM IDs", func(t *testing.T) {
		// Create mock configuration files to simulate existing VM IDs
		existingIDs := []int{101, 102, 103}
		for _, id := range existingIDs {
			filePath := filepath.Join(tempDir, fmt.Sprintf("%d.json", id))
			if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
				t.Fatalf("Failed to create mock config file: %v", err)
			}
		}

		vmid, err := folder.assignVMID()
		if err != nil {
			t.Fatalf("generateVMID failed: %v", err)
		}
		for _, id := range existingIDs {
			if vmid == id {
				t.Errorf("Generated VM ID conflicts with existing ID: %d", vmid)
			}
		}
		if vmid < 100 || vmid > 999 {
			t.Errorf("Generated VM ID out of range: %d", vmid)
		}
	})

	// Test case: No available VM IDs
	t.Run("No available VM IDs", func(t *testing.T) {
		// Fill the range 100-999 with mock configuration files
		for id := 100; id <= 999; id++ {
			filePath := filepath.Join(tempDir, fmt.Sprintf("%d.json", id))
			if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
				t.Fatalf("Failed to create mock config file: %v", err)
			}
		}

		_, err := folder.assignVMID()
		if err == nil {
			t.Fatalf("Expected error when no VM IDs are available, but got none")
		}
		expectedErr := "no available VMID can be assigned"
		if err.Error() != expectedErr {
			t.Errorf("Unexpected error message. Got: %s, Expected: %s", err.Error(), expectedErr)
		}
	})
}

func TestVerifyArgs(t *testing.T) {
	q := &Qemu{
		Folder: Folder{
			Monitor: "/tmp/monitor",
		},
	}

	tests := []struct {
		name     string
		config   Config
		expected []string
	}{
		{
			name: "Basic configuration",
			config: Config{
				ID:            1,
				Accelerator:   "kvm",
				Memory:        2048,
				CPUs:          2,
				BIOSPath:      "/usr/share/qemu/bios.bin",
				DiskPath:      "/tmp/disk.qcow2",
				CloudInitPath: "/tmp/cloud-init.iso",
				VNCPort:       5901,
				// SSHPort:     2222,
				UUID: "123e4567-e89b-12d3-a456-426614174000",
			},
			expected: []string{
				"-accel", "kvm",
				"-m", "2048",
				"-smp", "2,sockets=1,cores=2,threads=1",
				"-cpu", "host",
				"-M", "virt",
				"-bios", "/usr/share/qemu/bios.bin",
				"-device", "qemu-xhci",
				"-device", "usb-kbd",
				"-device", "usb-tablet",
				"-audiodev", "none,id=audio0",
				"-device", "intel-hda",
				"-device", "hda-duplex,audiodev=audio0",
				"-drive", "file=/tmp/disk.qcow2,format=qcow2,if=virtio",
				"-drive", "file=/tmp/cloud-init.iso,format=raw,media=cdrom,readonly=on",
				"-rtc", "base=utc,clock=host",
				"-vnc", "127.0.0.1:1,password=on",
				"-monitor", "unix:/tmp/monitor/1.sock,server,nowait",
				"-netdev", "user,id=net0,hostfwd=tcp::2222-:22",
				"-device", "virtio-net-pci,netdev=net0",
				"-smbios", "type=1,uuid=123e4567-e89b-12d3-a456-426614174000",
				"-device", "virtio-gpu-pci",
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := q.verifyArgs(tt.config)
			if !reflect.DeepEqual(args, tt.expected) {
				t.Errorf("verifyArgs() = %v, want %v", args, tt.expected)
			}
		})
	}
}

func TestGetOSImageInfo_AdditionalCases(t *testing.T) {
	// Setup environment variables
	os.Setenv("GO_QEMU_DEBIAN_VERSION", "11,12,13")
	os.Setenv("GO_QEMU_UBUNTU_VERSION", "20.04,22.04,24.04")
	os.Setenv("GO_QEMU_ROCKYLINUX_VERSION", "8,9,10")
	defer func() {
		os.Unsetenv("GO_QEMU_DEBIAN_VERSION")
		os.Unsetenv("GO_QEMU_UBUNTU_VERSION")
		os.Unsetenv("GO_QEMU_ROCKYLINUX_VERSION")
	}()

	m := &Qemu{}

	tests := []struct {
		name      string
		osName    string
		version   string
		expectErr bool
	}{
		{"Valid Debian Version 12", "debian", "12", false},
		{"Valid Ubuntu Version 24.04", "ubuntu", "24.04", false},
		{"Valid RockyLinux Version 10", "rockylinux", "10", false},
		{"Invalid Debian Version 14", "debian", "14", true},
		{"Invalid Ubuntu Version 19.10", "ubuntu", "19.10", true},
		{"Invalid RockyLinux Version 7", "rockylinux", "7", true},
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
		})
	}
}

// func TestCreate_AdditionalCases(t *testing.T) {
// 	// Setup environment variables
// 	os.Setenv("GO_QEMU_DEBIAN_VERSION", "11,12,13")
// 	os.Setenv("GO_QEMU_UBUNTU_VERSION", "20.04,22.04,24.04")
// 	os.Setenv("GO_QEMU_ROCKYLINUX_VERSION", "8,9,10")
// 	defer func() {
// 		os.Unsetenv("GO_QEMU_DEBIAN_VERSION")
// 		os.Unsetenv("GO_QEMU_UBUNTU_VERSION")
// 		os.Unsetenv("GO_QEMU_ROCKYLINUX_VERSION")
// 	}()
// 	// Setup temporary directory for testing
// 	tempDir := t.TempDir()
// 	q := &Qemu{
// 		Folder: Folder{
// 			Config:  filepath.Join(tempDir, "config"),
// 			Monitor: filepath.Join(tempDir, "monitor"),
// 		},
// 	}

// 	// Create necessary directories
// 	os.MkdirAll(q.Folder.Config, 0755)
// 	os.MkdirAll(q.Folder.Monitor, 0755)

// 	tests := []struct {
// 		name      string
// 		config    Config
// 		setup     func() // Optional setup for the test
// 		expectErr bool
// 	}{
// 		{
// 			name: "Valid Debian VM Creation",
// 			config: Config{
// 				OS:      "debian",
// 				Version: "12",
// 				Memory:  2048,
// 				CPUs:    2,
// 				VNCPort: 5902,
// 				// SSHPort:  2223,
// 				DiskSize: "16G",
// 			},
// 			expectErr: false,
// 		},
// 		{
// 			name: "Valid Ubuntu VM Creation",
// 			config: Config{
// 				OS:      "ubuntu",
// 				Version: "24.04",
// 				Memory:  4096,
// 				CPUs:    4,
// 				VNCPort: 5903,
// 				// SSHPort:  2224,
// 				DiskSize: "32G",
// 			},
// 			expectErr: false,
// 		},
// 		{
// 			name: "Valid RockyLinux VM Creation",
// 			config: Config{
// 				OS:      "rockylinux",
// 				Version: "10",
// 				Memory:  1024,
// 				CPUs:    1,
// 				VNCPort: 5904,
// 				// SSHPort:  2225,
// 				DiskSize: "16G",
// 			},
// 			expectErr: false,
// 		},
// 		{
// 			name: "Invalid OS Version",
// 			config: Config{
// 				OS:      "ubuntu",
// 				Version: "19.10",
// 				Memory:  2048,
// 				CPUs:    2,
// 				VNCPort: 5905,
// 				// SSHPort:  2226,
// 				DiskSize: "16G",
// 			},
// 			expectErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.setup != nil {
// 				tt.setup()
// 			}

// 			slog.Info("Starting test case", "config", tt.config)

// 			err := q.Create(tt.config)

// 			if tt.expectErr {
// 				if err == nil {
// 					t.Errorf("Expected error but got none")
// 				}
// 			} else {
// 				if err != nil {
// 					t.Errorf("Unexpected error: %v", err)
// 				}
// 			}
// 		})
// 	}
// }

func TestList_AdditionalCases(t *testing.T) {
	// Setup temporary directory for testing
	tempDir := t.TempDir()
	q := &Qemu{
		Folder: Folder{
			Config: filepath.Join(tempDir, "config"),
			PID:    filepath.Join(tempDir, "pid"),
		},
	}

	// Create necessary directories
	os.MkdirAll(q.Folder.Config, 0755)
	os.MkdirAll(q.Folder.PID, 0755)

	// Create mock VM configurations and PID files
	mockVMs := []struct {
		id        int
		status    string
		createPID bool
	}{
		{101, "running", true},
		{102, "running", true},
		{103, "stopped", false},
	}

	for _, vm := range mockVMs {
		// Create mock config file
		configPath := filepath.Join(q.Folder.Config, fmt.Sprintf("%d.json", vm.id))
		config := Config{
			ID:      vm.id,
			Memory:  2048,
			CPUs:    2,
			VNCPort: 5900 + vm.id,
			// SSHPort: 2200 + vm.id,
		}
		data, _ := json.Marshal(config)
		os.WriteFile(configPath, data, 0644)

		// Create mock PID file if VM is running
		if vm.createPID {
			pidPath := filepath.Join(q.Folder.PID, fmt.Sprintf("%d.pid", vm.id))
			os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", 1000+vm.id)), 0644)
		}
	}

	tests := []struct {
		name      string
		expectVMs int
	}{
		{
			name:      "List all VMs",
			expectVMs: len(mockVMs),
		},
		{
			name:      "List only running VMs",
			expectVMs: 2, // Only VMs with running status
		},
		{
			name:      "List only stopped VMs",
			expectVMs: 1, // Only VMs with stopped status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vms := q.List()

			if tt.name == "List only running VMs" {
				vms = filterVMsByStatus(vms, "running")
			} else if tt.name == "List only stopped VMs" {
				vms = filterVMsByStatus(vms, "stopped")
			}

			if len(vms) != tt.expectVMs {
				t.Errorf("Expected %d VMs, got %d", tt.expectVMs, len(vms))
			}

			for _, vm := range vms {
				expectedStatus := "stopped"
				for _, mockVM := range mockVMs {
					if mockVM.id == vm.Config.ID && mockVM.status == "running" {
						expectedStatus = "running"
						break
					}
				}
				if vm.Status != expectedStatus {
					t.Errorf("VM ID %d: Expected status %s, got %s", vm.Config.ID, expectedStatus, vm.Status)
				}
			}
		})
	}
}

// Helper function to filter VMs by status
func filterVMsByStatus(vms []*Instance, status string) []*Instance {
	var filtered []*Instance
	for _, vm := range vms {
		if vm.Status == status {
			filtered = append(filtered, vm)
		}
	}
	return filtered
}

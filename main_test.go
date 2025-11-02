package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCloudInit(t *testing.T) {
	os.Setenv("TEST_MODE", "true")
	defer os.Unsetenv("TEST_MODE")

	// Setup test folder
	testFolder := Folder{Main: "./testdata"}
	if err := os.MkdirAll(testFolder.Main, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}
	defer os.RemoveAll(testFolder.Main)

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
	metaDataPath := filepath.Join(testFolder.Main, ".cloudinit-101", "meta-data")
	if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
		t.Errorf("Expected meta-data file not found: %s", metaDataPath)
	}

	// Verify user-data file
	userDataPath := filepath.Join(testFolder.Main, ".cloudinit-101", "user-data")
	if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
		t.Errorf("Expected user-data file not found: %s", userDataPath)
	}
}

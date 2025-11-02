package main

import "time"

type Config struct {
	ID               int    `json:"id"`
	Hostname         string `json:"hostname"`
	Accelerator      string `json:"accelerator"`
	Memory           int    `json:"memory"`
	CPUs             int    `json:"cpus"` // TODO: expand to sockets, cores, threads
	BIOSPath         string `json:"bios_path"`
	DiskPath         string `json:"disk_path"`
	DiskSize         string `json:"disk_size"`
	CloudInit        string `json:"cloud_init"`
	OS               string `json:"os"`
	Version          string `json:"version"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	SSHAuthorizedKey string `json:"ssh_key"`
	SSHPort          int    `json:"ssh_port"`
	VNCPort          int    `json:"vnc_port"`
}

type Instance struct {
	Config    Config     `json:"config"`
	PID       int        `json:"pid"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

type Image struct {
	OS       string
	Version  string
	URL      string
	Filename string
}

type Folder struct {
	VM      string
	Config  string
	Log     string
	PID     string
	Monitor string
	Image   string
}

type CloudInit struct {
	UUID             string
	OS               string
	Hostname         string
	Username         string
	Password         string
	SSHAuthorizedKey string
}

type Progress struct {
	Total     int64
	Completed int64
}

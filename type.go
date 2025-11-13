package goQemu

import "time"

type Config struct {
	ID            int    `json:"id"`
	Hostname      string `json:"hostname"`
	Accelerator   string `json:"accelerator"`
	Memory        int    `json:"memory"`
	CPUs          int    `json:"cpus"` // TODO: expand to sockets, cores, threads
	BIOS          string `json:"bios"`
	DiskPath      string `json:"disk_path"`
	DiskSize      string `json:"disk_size"`
	CloudInitPath string `json:"cloud_init_path"`
	OS            string `json:"os"`
	Version       string `json:"version"`
	// Username         string    `json:"username"`
	// Password         string    `json:"password"`
	// SSHAuthorizedKey string    `json:"ssh_key"`
	VNCPort int `json:"vnc_port"`
	// UUID      string    `json:"uuid"`
	Network   []Network `json:"network"`
	CloudInit CloudInit `json:"cloud_init"`
	Options   Options   `json:"options"`
}

type Network struct {
	Bridge     string `json:"bridge"`
	Model      string `json:"model"`
	Vlan       int    `json:"vlan"`
	MACAddress string `json:"mac_address"`
	Firewall   bool   `json:"firewall"`
	Disconnect bool   `json:"disconnect"`
	MTU        int    `json:"mtu"`
	RateLimit  int    `json:"rate_limit"`
	Multiqueue int    `json:"multiqueue"`
	// IPv4       *IPConfig `json:"ipv4"`
	// IPv6       *IPConfig `json:"ipv6"`
}

type IPConfig struct {
	Mode    string `json:"mode"`
	Address string `json:"address"`
	Gateway string `json:"gateway"`
}

type CloudInit struct {
	// OS               string         `json:"os"`
	Hostname        string    `json:"hostname"`
	Username        string    `json:"username"`
	Password        string    `json:"passwd"`
	AuthorizedKey   string    `json:"authorized_key"`
	UpgradePackages bool      `json:"upgrade_packages"`
	DNSDomain       string    `json:"dns_domain"`
	DNSServers      []string  `json:"dns_servers"`
	IPv4            *IPConfig `json:"ipv4"`
	IPv6            *IPConfig `json:"ipv6"`
	// NetworkConfig   *NetworkConfig `json:"network_config,omitempty"`
}

type Options struct {
	UUID string `json:"uuid"`
}

// type NetworkConfig struct {
// 	IPv4 *IPConfig `json:"ipv4"`
// 	IPv6 *IPConfig `json:"ipv6"`
// }

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

type Qemu struct {
	Folder Folder
	Binary string
}

type Folder struct {
	VM      string
	Config  string
	Log     string
	PID     string
	Monitor string
	Image   string
}

type Progress struct {
	Total     int64
	Completed int64
}

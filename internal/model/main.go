package model

type Config struct {
	ID      int    `json:"id,omitempty"`                                         // VM ID
	Name    string `json:"name"`                                                 // VM Name
	Node    string `json:"node,omitempty"`                                       // Proxmox Node
	Storage string `json:"storage,omitempty"`                                    // Storage Pool
	OS      string `json:"os" binding:"required,oneof=debian ubuntu rockylinux"` // OS Type
	Version string `json:"version" binding:"required"`                           // OS Version
	CPU     int    `json:"cpu"`                                                  // Number of vCPU Cores
	Disk    string `json:"disk"`                                                 // Disk Size
	RAM     int    `json:"ram"`                                                  // RAM Size
	IP      string `json:"ip,omitempty"`                                         // Static IP Address
	Gateway string `json:"gateway,omitempty"`                                    // Gateway
	User    string `json:"user"`                                                 // SSH Username
	Passwd  string `json:"passwd"`                                               // SSH Password
	Pubkey  string `json:"pubkey,omitempty"`                                     // SSH Public Key
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	VMID    int    `json:"vm_id,omitempty"`
	IP      string `json:"ip,omitempty"`
}

type SSE struct {
	Step    string `json:"step"`
	Status  string `json:"status"`
	Message string `json:"message"`
	VMID    int    `json:"vm_id,omitempty"`
	IP      string `json:"ip,omitempty"`
}

type Status struct {
	IP        string `json:"ip"`
	Available bool   `json:"available"`
	VMID      int    `json:"vm_id"`
}

type VM struct {
	ID         int    `json:"vmid"`
	Name       string `json:"name"`
	OS         string `json:"os"`
	Running    bool   `json:"running"`
	Node       string `json:"node"`
	CPU        int    `json:"cpu"`
	Disk       int    `json:"disk"`
	Memory     int    `json:"memory"`
	MemoryUsed int    `json:"memory_used"`
}

type Node struct {
	Node       string  `json:"node"`
	MaxCPU     float64 `json:"max_cpu"`
	MaxMemory  float64 `json:"max_memory"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	MemoryUsed float64 `json:"memory_used"`
	Disk       float64 `json:"disk"`
	Running    bool    `json:"running"`
}

package models

// Service represents a monitored service
type Service struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      string          `json:"status"` // RUNNING, STOPPED, ERROR, MAINTENANCE
	Port        int             `json:"port"`
	URL         string          `json:"url"`
	Configs     []ServiceConfig `json:"configs"`
	Uptime      string          `json:"uptime"`
	CPUUsage    float64         `json:"cpuUsage"`    // Percent
	MemoryUsage float64         `json:"memoryUsage"` // MB
	Host        string          `json:"host,omitempty"`
}

// ServiceConfig represents a configuration file for a service
type ServiceConfig struct {
	Type       string `json:"type"` // YAML, JSON, INI, DOCKERFILE
	Path       string `json:"path"`
	Content    string `json:"content,omitempty"`
	LastEdited string `json:"lastEdited"`
}

// HostStats represents system resource usage
type HostStats struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Storage StorageStats `json:"storage"`
}

type CPUStats struct {
	Usage   float64 `json:"usage"`
	Cores   int     `json:"cores"`
	Threads int     `json:"threads"`
}

type MemoryStats struct {
	UsedGB  float64 `json:"usedGB"`
	TotalGB float64 `json:"totalGB"`
}

type StorageStats struct {
	UsedGB  float64 `json:"usedGB"`
	TotalGB float64 `json:"totalGB"`
}

package config

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Auth        AuthConfig        `yaml:"auth"`
	UptimeKuma  *UptimeKumaConfig `yaml:"uptime_kuma,omitempty"`
	Services    []ServiceConfig   `yaml:"services"`
	RemoteHosts []RemoteHost      `yaml:"remote_hosts,omitempty"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Port            int    `yaml:"port"`
	SessionSecret   string `yaml:"session_secret"`
	CORSAllowOrigin string `yaml:"cors_allow_origin"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	APIToken string `yaml:"api_token"`
}

// UptimeKumaConfig contains Uptime Kuma integration settings
type UptimeKumaConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	APIKey   string `yaml:"api_key,omitempty"`
}

// ServiceConfig defines a service to monitor
type ServiceConfig struct {
	Name          string   `yaml:"name"`
	URL           string   `yaml:"url"`
	Port          int      `yaml:"port"`
	Backend       string   `yaml:"backend"` // docker, uptime_kuma
	ContainerName string   `yaml:"container_name,omitempty"`
	KumaMonitorID int      `yaml:"kuma_monitor_id,omitempty"`
	Configs       []string `yaml:"configs,omitempty"`
}

// RemoteHost defines a remote instance for federation
type RemoteHost struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
}

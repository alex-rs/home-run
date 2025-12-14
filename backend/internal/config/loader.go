package config

import (
	"errors"
	"fmt"
	"os"

	"home-run-backend/internal/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	logger.WithField("path", path).Debug("Loading configuration file")

	data, err := os.ReadFile(path)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("Failed to read config file")
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.WithFields(logrus.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("Failed to parse config file")
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate
	if err := validate(&cfg); err != nil {
		logger.WithFields(logrus.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("Config validation failed")
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"services":     len(cfg.Services),
		"remote_hosts": len(cfg.RemoteHosts),
		"port":         cfg.Server.Port,
		"uptime_kuma":  cfg.UptimeKuma != nil,
	}).Info("Configuration loaded successfully")

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.SessionSecret == "" {
		cfg.Server.SessionSecret = "change-me-in-production-32chars"
	}
	if cfg.Server.CORSAllowOrigin == "" {
		cfg.Server.CORSAllowOrigin = "*"
	}
}

func validate(cfg *Config) error {
	// Required auth fields
	if cfg.Auth.Username == "" {
		return errors.New("auth.username is required")
	}
	if cfg.Auth.Password == "" {
		return errors.New("auth.password is required")
	}
	if cfg.Auth.APIToken == "" {
		return errors.New("auth.api_token is required for federation")
	}

	// Validate services
	for i, svc := range cfg.Services {
		if svc.Name == "" {
			return fmt.Errorf("services[%d].name is required", i)
		}
		if svc.Backend != "docker" && svc.Backend != "uptime_kuma" {
			return fmt.Errorf("services[%d].backend must be 'docker' or 'uptime_kuma', got '%s'", i, svc.Backend)
		}
		if svc.Backend == "docker" && svc.ContainerName == "" {
			return fmt.Errorf("services[%d].container_name is required for docker backend", i)
		}
		if svc.Backend == "uptime_kuma" {
			if cfg.UptimeKuma == nil {
				return fmt.Errorf("services[%d] uses uptime_kuma backend but uptime_kuma config is missing", i)
			}
			if svc.KumaMonitorID == 0 {
				return fmt.Errorf("services[%d].kuma_monitor_id is required for uptime_kuma backend", i)
			}
		}
	}

	// Validate Uptime Kuma config if present
	if cfg.UptimeKuma != nil && cfg.UptimeKuma.URL == "" {
		return errors.New("uptime_kuma.url is required when uptime_kuma is configured")
	}

	// Validate remote hosts
	for i, host := range cfg.RemoteHosts {
		if host.Name == "" {
			return fmt.Errorf("remote_hosts[%d].name is required", i)
		}
		if host.Endpoint == "" {
			return fmt.Errorf("remote_hosts[%d].endpoint is required", i)
		}
		if host.Token == "" {
			return fmt.Errorf("remote_hosts[%d].token is required", i)
		}
	}

	return nil
}

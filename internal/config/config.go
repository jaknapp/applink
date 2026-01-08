package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the user's applink configuration
type Config struct {
	Services map[string]ClientCredentials `yaml:"services"`
	Settings Settings                     `yaml:"settings"`
}

// ClientCredentials holds OAuth client credentials for a service
type ClientCredentials struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// Settings holds general applink settings
type Settings struct {
	CallbackPort int `yaml:"callback_port"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Services: make(map[string]ClientCredentials),
		Settings: Settings{
			CallbackPort: 8888,
		},
	}
}

// ConfigDir returns the path to the applink config directory
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "applink"), nil
}

// ConfigPath returns the path to the main config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file, creating defaults if it doesn't exist
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Ensure callback port has a default
	if cfg.Settings.CallbackPort == 0 {
		cfg.Settings.CallbackPort = 8888
	}

	return cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	AuthType  AuthType `json:"auth_type"`
	APIKey    string   `json:"api_key,omitempty"`
	ProjectID string   `json:"project_id,omitempty"`
	Region    string   `json:"region,omitempty"`
}

// getConfigDir returns the application config directory path
func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, AppConfigDir), nil
}

// getConfigPath returns the main config file path
func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// LoadConfig reads and parses the configuration file
func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), DirPerms); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, ConfigFilePerms)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	switch c.AuthType {
	case AuthTypeAPIKey:
		if c.APIKey == "" {
			return ErrMissingAPIKey
		}
	case AuthTypeVertex:
		if c.ProjectID == "" {
			return ErrMissingProjectID
		}
		if c.Region == "" {
			return ErrMissingRegion
		}
	default:
		return ErrInvalidAuthType
	}
	return nil
}

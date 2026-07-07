package internal

import (
	"encoding/json"
	"os"
)

type Config struct {
	Folders []string `json:"folders"`
}

const configFile = "config.json"

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.Folders = []string{}
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Folders == nil {
		cfg.Folders = []string{}
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

func (c *Config) AddFolder(path string) {
	for _, f := range c.Folders {
		if f == path {
			return
		}
	}
	c.Folders = append(c.Folders, path)
}

func (c *Config) RemoveFolder(path string) {
	var updated []string
	for _, f := range c.Folders {
		if f != path {
			updated = append(updated, f)
		}
	}
	c.Folders = updated
}

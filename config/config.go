package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme string `yaml:"theme"`
}

func configPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")

	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}

	dir := filepath.Join(base, "kanba")

	for _, name := range []string{"config.yaml", "config.yml"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", nil
}

func Load() (*Config, error) {
	cfg := &Config{}

	p, err := configPath()
	if err != nil || p == "" {
		return cfg, nil
	}

	f, err := os.Open(p)
	if err != nil {
		return cfg, nil
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", p, err)
	}

	return cfg, nil
}

package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml: %w", err)
	}

	return &cfg, nil
}

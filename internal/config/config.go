package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Parse(configPath string, conf interface{}) error {
	configYml, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading error: %w", err)
	}

	if err := yaml.Unmarshal(configYml, conf); err != nil {
		return fmt.Errorf("can't unmarshal %s: %w", configPath, err)
	}

	return nil
}

package build

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
	Build   Build  `yaml:"build"`
}

// FromYaml parses a yaml file content string into a Config struct
func FromYaml(content []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %s", err)
	}
	return &config, nil
}

package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dep      DepConfig   `yaml:"dep"`
	GoMod    GoModConfig `yaml:"gomod"`
	Yarn     YarnConfig  `yaml:"yarn"`
	Patterns []string    `yaml:"patterns"`
}

type GoModConfig struct {
	GoModFile string `yaml:"mod-file"`
}

type DepConfig struct {
	VendorDirs []string `yaml:"vendor-dirs"`
	Lockfile   string   `yaml:"lockfile"`
}

type YarnConfig struct {
	NodeModulesDirs []string `yaml:"node-modules-dirs"`
	Lockfile        string   `yaml:"lockfile"`
}

func Load(filename string) (*Config, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.UnmarshalStrict(body, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

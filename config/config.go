package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Resolvers ResolverConfig `yaml:"resolvers"`
}

type ResolverConfig struct {
	Dep  *DepResolver  `yaml:"dep"`
	Yarn *YarnResolver `yaml:"yarn"`
}

type DepResolver struct {
	VendorDir string `yaml:"vendor-dir"`
	LockFile  string `yaml:"lock-file"`
}

type YarnResolver struct {
	ModuleDir string `yaml:"module-dir"`
	LockFile  string `yaml:"lock-file"`
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

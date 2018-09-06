package config

import (
	"io/ioutil"

	"github.com/stackrox/ossls/resolver"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Resolvers ResolverConfig `yaml:"resolvers"`
}

type ResolverConfig struct {
	Dep *resolver.DepResolver `yaml:"dep"`
	Js  *resolver.JsResolver  `yaml:"js"`
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

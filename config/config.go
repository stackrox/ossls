package config

import (
	"io/ioutil"

	"github.com/stackrox/ossls/resolver"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Resolvers    ResolverConfig        `yaml:"resolvers"`
	Dependencies map[string]Dependency `yaml:"dependencies"`
}

type ResolverConfig struct {
	Dep *resolver.DepResolver `yaml:"dep"`
	Js  *resolver.JsResolver  `yaml:"js"`
}

type ContentConfig struct {
	FileHash    string
	FieldHashes map[string]string
}

func (e *ContentConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		fieldHashes = make(map[string]string)
		fileHash    string
	)

	*e = ContentConfig{}

	if err := unmarshal(&fieldHashes); err == nil {
		e.FieldHashes = fieldHashes
		return nil
	}

	if err := unmarshal(&fileHash); err != nil {
		return err
	}

	e.FileHash = fileHash
	return nil
}

func (e ContentConfig) MarshalYAML() (interface{}, error) {
	switch {
	case len(e.FileHash) == 0 && len(e.FieldHashes) == 0:
		panic("both empty")
	case len(e.FileHash) != 0 && len(e.FieldHashes) != 0:
		panic("both non-empty")
	case len(e.FileHash) != 0:
		return e.FileHash, nil
	case len(e.FieldHashes) != 0:
		return e.FieldHashes, nil
	default:
		panic("impossible")
	}
}

type Dependency struct {
	URL         string                   `yaml:"url"`
	License     string                   `yaml:"license"`
	Files       map[string]ContentConfig `yaml:"files"`
	Attribution []string                 `yaml:"attribution"`
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

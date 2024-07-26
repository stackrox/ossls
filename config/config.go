package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type PatternConfig struct {
	Patterns        []string `json:"patterns"`
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
}

type Config struct {
	Dep   DepConfig   `json:"dep"`
	GoMod GoModConfig `json:"gomod"`
	Yarn  YarnConfig  `json:"yarn"`
	Npm   NpmConfig   `json:"npm"`
	PatternConfig
}

type GoModConfig struct {
	GoModFile string `json:"mod-file"`
}

type DepConfig struct {
	VendorDirs []string `json:"vendor-dirs"`
	Lockfile   string   `json:"lockfile"`
}

type YarnConfig struct {
	NodeModulesDirs []string `json:"node-modules-dirs"`
	Lockfile        string   `json:"lockfile"`
}

type NpmConfig struct {
	NodeModulesDir string `json:"node-modules-dir"`
	Lockfile       string `json:"lockfile"`
}

func Load(filename string) (*Config, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := yaml.YAMLToJSON(body)
	if err != nil {
		return nil, err
	}
	jsonDec := json.NewDecoder(bytes.NewReader(jsonBytes))
	jsonDec.DisallowUnknownFields()

	var config Config
	if err := jsonDec.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

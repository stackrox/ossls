package resolver

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type constraint struct {
	Name     string `toml:"name"`
	Revision string `toml:"revision"`
	Version  string `toml:"version"`
	Branch   string `toml:"branch"`
}

type DepResolver struct {
	VendorDir string `yaml:"vendor-dir"`
	Manifest  string `yaml:"manifest"`
}

func (r *DepResolver) Repos() ([]string, error) {
	constraints, err := parseManifest(r.Manifest)
	if err != nil {
		return nil, err
	}

	repos := make([]string, len(constraints))

	for index, constraint := range constraints {
		repos[index] = filepath.Join(r.VendorDir, constraint.Name)
	}

	return repos, nil
}

func parseManifest(filename string) ([]constraint, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type lock struct {
		Constraints []constraint `toml:"constraint"`
	}

	var l lock
	if err := toml.Unmarshal(body, &l); err != nil {
		return nil, err
	}

	return l.Constraints, nil
}

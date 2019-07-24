package resolver

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type constraint struct {
	Name     string `toml:"name"`
	Revision string `toml:"revision"`
	Version  string `toml:"version"`
	Branch   string `toml:"branch"`
}

type DepProject struct {
	name    string
	version string
}

var _ Project = (*DepProject)(nil)

func (p DepProject) Name() string {
	return p.name
}

func (p DepProject) Optional() bool {
	return false
}

func (p DepProject) Version() string {
	return p.version
}

func ProjectsFromDepLockfile(filename string) ([]Project, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type lock struct {
		Constraints []constraint `toml:"projects"`
	}

	var l lock
	if err := toml.Unmarshal(body, &l); err != nil {
		return nil, err
	}

	projects := make([]Project, len(l.Constraints))

	for index, constraint := range l.Constraints {
		version := constraint.Version
		if version == "" {
			version = constraint.Revision
		}
		projects[index] = DepProject{
			name:    constraint.Name,
			version: version,
		}
	}

	return projects, nil
}

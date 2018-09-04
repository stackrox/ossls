package resolver

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type project struct {
	Name     string `toml:"name"`
	Revision string `toml:"revision"`
	Version  string `toml:"version"`
}

type DepResolver struct {
	VendorDir string `yaml:"vendor-dir"`
	LockFile  string `yaml:"lock-file"`
}

func (r *DepResolver) Repos() ([]Dependency, error) {
	projects, err := parseLockfile(r.LockFile)
	if err != nil {
		return nil, err
	}

	repos := make([]Dependency, len(projects))

	for index, project := range projects {
		repos[index] = Dependency{
			Name:      project.Name,
			Version:   project.Version,
			Reference: project.Revision,
			Path:      filepath.Join(r.VendorDir, project.Name),
		}
	}

	return repos, nil
}

func parseLockfile(filename string) ([]project, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type lock struct {
		Projects []project `toml:"projects"`
	}

	var l lock
	if err := toml.Unmarshal(body, &l); err != nil {
		return nil, err
	}

	return l.Projects, nil
}

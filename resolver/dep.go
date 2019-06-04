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

//
//func (r *DepResolver) Repos() ([]string, error) {
//	constraints, err := parseManifest(r.Manifest)
//	if err != nil {
//		return nil, err
//	}
//
//	repos := make([]string, len(constraints))
//
//	for index, constraint := range constraints {
//		repos[index] = filepath.Join(r.VendorDir, constraint.Name)
//	}
//
//	return repos, nil
//}

type DepProject struct {
	name string
}

func (p DepProject) Name() string {
	return p.name
}

func (p DepProject) Optional() bool {
	return false
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
		projects[index] = DepProject{constraint.Name}
	}

	return projects, nil
}

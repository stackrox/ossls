package resolver

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type JsResolver struct {
	ModuleDir string `yaml:"module-dir"`
	Manifest  string `yaml:"manifest"`
}

func (r *JsResolver) Repos() ([]Dependency, error) {
	deps, err := parsePackageJson(r.Manifest)
	if err != nil {
		return nil, err
	}

	repos := make([]Dependency, 0, len(deps))

	for name, version := range deps {
		dep := Dependency{
			Name:    name,
			Version: version,
			Path:    filepath.Join(r.ModuleDir, name),
		}
		repos = append(repos, dep)
	}

	return repos, nil
}

func parsePackageJson(filename string) (map[string]string, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type packageFile struct {
		Dependencies map[string]string `json:"dependencies"`
	}

	var pkg packageFile
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, err
	}

	return pkg.Dependencies, nil
}

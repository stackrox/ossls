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

func (r *JsResolver) Repos() ([]string, error) {
	deps, err := parsePackageJson(r.Manifest)
	if err != nil {
		return nil, err
	}

	repos := make([]string, 0, len(deps))

	for name := range deps {
		repo := filepath.Join(r.ModuleDir, name)
		repos = append(repos, repo)
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

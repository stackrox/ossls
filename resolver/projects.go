package resolver

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Project interface {
	Name() string
	Optional() bool
	Version() string
}

type Dependency struct {
	Name      string
	Alias     string
	Files     []string
	SourceDir string
	Version   string
}

func LocateGoModProjects(projects []GoModProject) (map[string]Dependency, error) {
	goPath := GoPath()
	result := make(map[string]Dependency, len(projects))

	for _, project := range projects {
		result[project.Name()] = Dependency{
			Name:      project.Name(),
			Version:   project.Version(),
			SourceDir: filepath.Join(goPath, "pkg/mod", project.sourcePath()),
		}
	}

	return result, nil
}

func LocateProjects(root string, projects []Project) (map[string]Dependency, error) {
	locations := make(map[string]Dependency)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name() < projects[j].Name()
	})

	if err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrap(err, "failed to walk "+path)
			}
			if !info.IsDir() {
				return nil
			}

			for _, project := range projects {
				if !strings.HasSuffix(path, "/"+project.Name()) {
					continue
				}

				oldPath, found := locations[project.Name()]
				switch {
				case !found:
					locations[project.Name()] = Dependency{
						Name:      project.Name(),
						Version:   project.Version(),
						SourceDir: path,
					}
				case len(path) < len(oldPath.SourceDir):
					dep := locations[project.Name()]
					dep.SourceDir = path
					locations[project.Name()] = dep
				}
				return nil
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	// Sanity check that all projects were located successfully.
	for _, project := range projects {
		if project.Optional() {
			continue
		}
		_, found := locations[project.Name()]
		if !found {
			return nil, errors.New("failed to locate project " + project.Name())
		}
	}

	return locations, nil
}

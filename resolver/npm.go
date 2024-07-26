package resolver

import (
	"encoding/json"
	"os"
	"strings"
)

func ProjectsFromNpmLockfileV3(filename string) ([]Project, error) {
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var packageLock PackageLockV3

	json.Unmarshal(byteValue, &packageLock)

	return asNpmProjects(packageLock), nil
}

type NpmProject struct {
	name     string
	version  string
	optional bool
}

var _ Project = (*NpmProject)(nil)

func (p NpmProject) Name() string {
	return p.name
}

func (p NpmProject) Optional() bool {
	return p.optional
}

func (p NpmProject) Version() string {
	return p.version
}

type PackageLockV3 struct {
	Name     string                `json:"name"`
	Version  string                `json:"version"`
	Packages map[string]NpmPackage `json:"packages"`
}

type NpmPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func asNpmProjects(packageLock PackageLockV3) []Project {
	projectSet := make(map[string]NpmProject, len(packageLock.Packages)-1)

	for pkg, entry := range packageLock.Packages {
		if pkg == "" {
			// skip the top-level package as it refers to the project itself
			continue
		}

		projectSet[pkg] = NpmProject{
			// Remove the `node_modules/` root directory prefix from each `pkg`
			name:    strings.TrimPrefix(pkg, "node_modules/"),
			version: entry.Version,
			// optional packages that are included as dependencies in the build will be declared at the top
			// level of packageLock.Packages, so we can explicitly mark optional as false here
			optional: false,
		}
	}

	projectList := make([]Project, 0, len(projectSet))
	for _, project := range projectSet {
		projectList = append(projectList, project)
	}

	return projectList
}

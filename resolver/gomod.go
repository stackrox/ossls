package resolver

import (
	"fmt"
	"io/ioutil"

	"github.com/stackrox/ossls/gomod/modfile"
	"github.com/stackrox/ossls/gomod/module"
)

type GoModProject struct {
	pkg         module.Version
	replacement *module.Version
}

var _ Project = (*GoModProject)(nil)

func (p GoModProject) Name() string {
	return p.pkg.Path
}

func (p GoModProject) Optional() bool {
	return false
}

func (p GoModProject) Version() string {
	return p.pkg.Version
}

func (p GoModProject) sourcePath() string {
	effective := p.replacement
	if effective == nil {
		effective = &p.pkg
	}
	return fmt.Sprintf("%s@%s", effective.Path, effective.Version)
}

func ProjectsFromGoModFile(filename string) ([]GoModProject, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	goModFile, err := modfile.Parse(filename, contents, nil)

	replacementMap := make(map[module.Version]module.Version)
	for _, replace := range goModFile.Replace {
		replacementMap[replace.Old] = replace.New
	}

	var projects []GoModProject
	for _, requirement := range goModFile.Require {
		project := GoModProject{
			pkg: requirement.Mod,
		}
		if repl, ok := replacementMap[requirement.Mod]; ok {
			project.replacement = &repl
		}
		projects = append(projects, project)
	}

	return projects, nil
}

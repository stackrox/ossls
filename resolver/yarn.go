package resolver

import "github.com/stackrox/ossls/yarn"

func ProjectsFromYarnLockfile(filename string) ([]Project, error) {
	entries, err := yarn.Parse(filename)
	if err != nil {
		return nil, err
	}

	return asProjects(entries), nil
}

type YarnProject struct {
	name         string
	version      string
	optional     bool
	optionaldeps map[string]struct{}
	deps         map[string]struct{}
}

var _ Project = (*YarnProject)(nil)

func (p YarnProject) Name() string {
	return p.name
}

func (p YarnProject) Optional() bool {
	return p.optional
}
func (p YarnProject) Version() string {
	return p.version
}

func joinProjects(first YarnProject, second yarn.Entry) YarnProject {
	for _, dep := range second.Dependencies {
		first.deps[dep.Name] = struct{}{}
	}

	for _, dep := range second.OptionalDependencies {
		first.optionaldeps[dep.Name] = struct{}{}
	}
	return first
}

func depMap(deps []yarn.Dependency) map[string]struct{} {
	results := make(map[string]struct{}, len(deps))
	for _, dep := range deps {
		results[dep.Name] = struct{}{}
	}
	return results
}

func asProjects(entries []yarn.Entry) []Project {

	projectSet := make(map[string]YarnProject, len(entries))

	for _, entry := range entries {

		first, found := projectSet[entry.Name]
		if found {
			projectSet[entry.Name] = joinProjects(first, entry)
		} else {
			projectSet[entry.Name] = YarnProject{
				name:         entry.Name,
				optionaldeps: depMap(entry.OptionalDependencies),
				deps:         depMap(entry.Dependencies),
				version:      entry.Version,
			}
		}

	}

	for _, entry := range entries {
		for _, dep := range entry.OptionalDependencies {
			markOptional(projectSet, dep.Name)
		}
	}

	projectList := make([]Project, 0)
	for _, project := range projectSet {
		projectList = append(projectList, project)
	}

	return projectList
}

func markOptional(projects map[string]YarnProject, name string) {
	project, found := projects[name]
	if !found {
		return
	}

	if project.optional {
		return
	}

	project.optional = true
	projects[name] = project

	for dep := range project.deps {
		markOptional(projects, dep)
	}

	for dep := range project.optionaldeps {
		markOptional(projects, dep)
	}
}

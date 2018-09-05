package yarn

type Entry struct {
	Name                 string
	Specs                []string
	Resolved             string
	Version              string
	Dependencies         []Dependency
	OptionalDependencies []Dependency
}

type Dependency struct {
	Name   string
	Semver string
}

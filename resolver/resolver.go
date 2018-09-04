package resolver

type Dependency struct {
	Name      string
	Path      string
	Version   string
	Reference string
}

type Resolver interface {
	Repos() ([]Dependency, error)
}

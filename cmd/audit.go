package cmd

import (
	"fmt"

	"os"
	"path/filepath"
	"strings"

	"sort"

	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/integrity"
)

type Violation struct {
	Dependency string
	Message    string
}

func NewViolation(dependency string, message string) Violation {
	return Violation{
		Dependency: dependency,
		Message:    message,
	}
}

func DiffViolations(expected map[string]config.Dependency, actual map[string]struct{}) []Violation {
	violations := []Violation{}

	for name := range actual {
		if _, found := expected[name]; !found {
			violations = append(violations, NewViolation(name, "dependency added"))
		}
	}

	for name := range expected {
		if _, found := actual[name]; !found {
			violations = append(violations, NewViolation(name, "dependency deleted"))
		}
	}

	return violations
}

func DependencyViolations(name string, dependency config.Dependency) ([]Violation, error) {
	violations := []Violation{}

	if !strings.HasPrefix(dependency.URL, "http://") && !strings.HasPrefix(dependency.URL, "https://") {
		violations = append(violations, NewViolation(name, "invalid url "+dependency.URL))
	}

	if dependency.License == "" {
		violations = append(violations, NewViolation(name, "no license"))
	}

	if len(dependency.Attribution) < 1 {
		violations = append(violations, NewViolation(name, "no attribution"))
	}

	if len(dependency.Files) < 1 {
		violations = append(violations, NewViolation(name, "no files"))
	}

	for file, checksum := range dependency.Files {
		filename := filepath.Join(name, file)
		if _, err := os.Stat(filename); err != nil {
			violations = append(violations, NewViolation(name, fmt.Sprintf("file %s does not exist.", filename)))
			continue
		}

		matched, actual, err := integrity.Verify(filename, checksum)
		if err != nil {
			return nil, err
		}

		if !matched {
			violations = append(violations, NewViolation(name, fmt.Sprintf("checksum mismatch for %s. expected %s but got %s", filename, checksum, actual)))
		}
	}

	return violations, nil
}

func Audit(cfg *config.Config) ([]Violation, error) {

	var nothing = struct{}{}

	// Get list of Golang (via dep) dependencies
	goDeps, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return nil, err
	}

	// Get list of JavaScript (via package.json) dependencies
	jsDeps, err := cfg.Resolvers.Js.Repos()
	if err != nil {
		return nil, err
	}

	// Alias know dependencies map for ease of use
	expectedDeps := cfg.Dependencies
	actualDeps := make(map[string]struct{}, len(goDeps)+len(jsDeps))

	// Add dependency directory paths from both list together
	for _, dep := range goDeps {
		actualDeps[dep.Path] = nothing
	}
	for _, dep := range jsDeps {
		actualDeps[dep.Path] = nothing
	}

	violations := DiffViolations(expectedDeps, actualDeps)

	for name, dependency := range expectedDeps {
		if _, found := actualDeps[name]; !found {
			continue
		}

		vs, err := DependencyViolations(name, dependency)
		if err != nil {
			return nil, err
		}
		violations = append(violations, vs...)
	}

	return violations, nil
}

func AuditPrint(violations []Violation) {
	deps := make(map[string][]Violation)
	for _, violation := range violations {
		deps[violation.Dependency] = append(deps[violation.Dependency], violation)
	}

	names := make([]string, 0, len(deps))
	for name := range deps {
		names = append(names, name)
	}

	sort.Strings(names)

	for index, name := range names {
		if index != 0 {
			fmt.Println()
		}
		fmt.Printf("%s:\n", name)
		for _, issue := range deps[name] {
			fmt.Printf("  - %s\n", issue.Message)
		}
	}
}

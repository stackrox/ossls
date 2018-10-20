package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
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

func DiffViolations(expected map[string]config.Dependency, actual map[string]struct{}) map[string][]Violation {
	violations := make(map[string][]Violation)

	for name := range actual {
		if _, found := expected[name]; !found {
			violations[name] = []Violation{NewViolation(name, "dependency added")}
		}
	}

	for name := range expected {
		if _, found := actual[name]; !found {
			violations[name] = []Violation{NewViolation(name, "dependency deleted")}
		}
	}

	return violations
}

func DependencyViolations(name string, dependency config.Dependency) ([]Violation, error) {
	violations := make([]Violation, 0)

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

	for file, content := range dependency.Files {
		filename := filepath.Join(name, file)

		fvs, err := FileViolations(name, filename, content)
		if err != nil {
			return nil, err
		}

		violations = append(violations, fvs...)
	}

	return violations, nil
}

func FileViolations(dependency string, filename string, content config.ContentConfig) ([]Violation, error) {
	violations := make([]Violation, 0)

	if _, err := os.Stat(filename); err != nil {
		violations = append(violations, NewViolation(dependency, fmt.Sprintf("file %s does not exist.", filename)))
		return violations, nil
	}

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if len(content.FileHash) != 0 {
		matched, actual := integrity.VerifyBytes(body, content.FileHash)
		if !matched {
			violations = append(violations, NewViolation(dependency, fmt.Sprintf("checksum mismatch for %s. expected %s but got %s", filename, content.FileHash, actual)))
		}
		return violations, nil
	}

	fields := make(map[string]interface{})

	if err := json.Unmarshal(body, &fields); err != nil {
		violations = append(violations, NewViolation(dependency, fmt.Sprintf("unable to unmarshal %s. invalid json", filename)))
		return violations, nil
	}

	for key, expected := range content.FieldHashes {
		value, found := fields[key]

		if !found {
			violations = append(violations, NewViolation(dependency, fmt.Sprintf("missing field %s in %s.", key, filename)))
			continue
		}

		matched, actual, err := integrity.VerifyField(value, expected)
		if err != nil {
			violations = append(violations, NewViolation(dependency, fmt.Sprintf("unable to checksum field %s in %s.", key, filename)))
			continue
		}

		if !matched {
			violations = append(violations, NewViolation(dependency, fmt.Sprintf("checksum mismatch for %s field %s. expected %s but got %s", filename, key, expected, actual)))
			continue
		}
	}

	return violations, nil
}

func Audit(cfg *config.Config) (map[string][]Violation, int, error) {
	var nothing = struct{}{}

	// Get list of Golang (via dep) dependencies
	goRepos, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return nil, 0, err
	}

	// Get list of JavaScript (via package.json) dependencies
	jsRepos, err := cfg.Resolvers.Js.Repos()
	if err != nil {
		return nil, 0, err
	}

	// Alias known dependencies map for ease of use
	expectedDeps := cfg.Dependencies
	actualDeps := make(map[string]struct{}, len(goRepos)+len(jsRepos))

	// Add dependency directory paths from both list together
	for _, repo := range goRepos {
		actualDeps[repo] = nothing
	}
	for _, repo := range jsRepos {
		actualDeps[repo] = nothing
	}

	violations := DiffViolations(expectedDeps, actualDeps)

	for name, dependency := range expectedDeps {
		if _, found := actualDeps[name]; !found {
			continue
		}

		vs, err := DependencyViolations(name, dependency)
		if err != nil {
			return nil, 0, err
		}

		violations[name] = vs
	}

	total := 0
	for _, issues := range violations {
		total += len(issues)
	}

	return violations, total, nil
}

func AuditPrint(violations map[string][]Violation) {
	var (
		total = 0
		names = make([]string, 0, len(violations))
	)

	for name := range violations {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		total += len(violations[name])
		switch len(violations[name]) {
		case 0:
			color.Green("✓ %s\n", name)
		default:
			color.Red("✗ %s\n", name)
		}
		for _, issue := range violations[name] {
			color.HiBlack("  ↳ %s\n", issue.Message)
		}
	}
}

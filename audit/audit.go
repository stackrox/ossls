package audit

import (
	"path/filepath"

	"fmt"

	"github.com/stackrox/ossls/resolver"
)

func findLicense(dirname string) []string {
	var (
		foundFiles = []string{}
		patterns   = []string{
			"package.json",
			"LICENSE*",
			"COPYING*",
			"license*",
			"copying*",
		}
	)

	for _, pattern := range patterns {
		glob := filepath.Join(dirname, pattern)
		matches, _ := filepath.Glob(glob)
		for _, match := range matches {
			foundFiles = append(foundFiles, match)
		}
	}

	return foundFiles
}

func Dependencies(dependencies []resolver.Dependency) {
	for index, dep := range dependencies {
		fmt.Printf("[%d/%d] %s\n", index+1, len(dependencies), dep.Name)

		foundFiles := findLicense(dep.Path)

		if len(foundFiles) == 0 {
			fmt.Printf("  No license files found\n")
			continue
		}

		for _, file := range foundFiles {
			fmt.Printf("  Found file %s\n", file)
		}
	}
}

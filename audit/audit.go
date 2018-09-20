package audit

import (
	"path/filepath"

	"os"

	"github.com/joshdk/licensor/spdx"
)

func FindLicenseFiles(dirname string) []string {
	var (
		foundFiles = []string{}
		patterns   = []string{
			"*AUTHOR*",
			"*COPYING*",
			"*LICENSE*",
			"*NOTICE*",
			"*author*",
			"*copying*",
			"*license*",
			"*notice*",
			"package.json",
		}
	)

	for _, pattern := range patterns {
		glob := filepath.Join(dirname, pattern)
		matches, _ := filepath.Glob(glob)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			foundFiles = append(foundFiles, match)
		}
	}

	return foundFiles
}

func extractLicense(filename string) ([]spdx.License, error) {
	switch filepath.Base(filename) {
	case "package.json":
		return extractLicenseFromPackageJson(filename)
	default:
		return extractLicenseFromFileBody(filename)
	}
}

type info struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func Dependencies(dependencies []string) (map[string]map[string][]spdx.License, error) {
	results := make(map[string]map[string][]spdx.License, len(dependencies))

	for _, dep := range dependencies {
		foundFiles := FindLicenseFiles(dep)

		if len(foundFiles) == 0 {
			continue
		}

		result := make(map[string][]spdx.License, len(foundFiles))

		for _, file := range foundFiles {
			licenses, err := extractLicense(file)
			if err != nil {
				return nil, err
			}

			if len(licenses) == 0 {
				continue
			}

			result[file] = licenses
		}

		results[dep] = result
	}

	return results, nil
}

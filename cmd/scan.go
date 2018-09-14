package cmd

import (
	"fmt"

	"io/ioutil"

	"path/filepath"

	"github.com/joshdk/licensor"
	"github.com/joshdk/licensor/spdx"
	"github.com/stackrox/ossls/audit"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/integrity"
	"gopkg.in/yaml.v2"
)

func Scan(directories []string) (map[string]config.Dependency, error) {
	dependencies := make(map[string]config.Dependency, len(directories))

	for _, directory := range directories {
		dependency, err := ScanSingle(directory)
		if err != nil {
			return nil, err
		}

		dependencies[directory] = *dependency
	}

	return dependencies, nil
}

func ScanSingle(directory string) (*config.Dependency, error) {
	var (
		licenseFiles   = audit.FindLicenseFiles(directory)
		bestConfidence = 0.8
		bestLicense    *spdx.License
		dependency     = &config.Dependency{
			Files: make(map[string]string),
		}
	)

	// Calculate checksums for the found files
	for _, file := range licenseFiles {
		checksum, err := integrity.Checksum(file)
		if err != nil {
			return nil, err
		}

		frag, err := filepath.Rel(directory, file)
		if err != nil {
			return nil, err
		}

		dependency.Files[frag] = checksum
	}

	// Attempt to divine the best license from the set of found files
	for _, file := range licenseFiles {
		body, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		match := licensor.Best(body)

		if match.Confidence > bestConfidence {
			bestConfidence = match.Confidence
			bestLicense = &match.License
		}
	}

	if bestLicense != nil {
		dependency.License = bestLicense.Identifier
	}

	return dependency, nil
}

func ScanPrint(configFile string, dependencies map[string]config.Dependency) {
	dummy := map[string]map[string]config.Dependency{
		"dependencies": dependencies,
	}

	out, _ := yaml.Marshal(dummy)

	fmt.Printf("The given directories were scanned\n")
	fmt.Printf("for potentially relevant license and attribution files.\n")
	fmt.Printf("This set may include too many or too few files. Please audit carefully.\n")
	fmt.Println()
	fmt.Printf("To pin this dependency, add the following to %s and fill in missing information.\n", configFile)
	fmt.Println()
	fmt.Printf("%s\n", string(out))
}

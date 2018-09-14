package cmd

import (
	"fmt"

	"io/ioutil"

	"github.com/joshdk/licensor"
	"github.com/joshdk/licensor/spdx"
	"github.com/stackrox/ossls/audit"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/integrity"
	"gopkg.in/yaml.v2"
)

func Scan(directory string) (*config.Dependency, error) {
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

		dependency.Files[file] = checksum
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

func ScanPrint(configFile string, directory string, dependency *config.Dependency) {
	dummy := map[string]map[string]config.Dependency{
		"dependencies": {
			directory: *dependency,
		},
	}

	out, _ := yaml.Marshal(dummy)

	fmt.Printf("The directory %s was scanned\n", directory)
	fmt.Printf("for potentially relevant license and attribution files.\n")
	fmt.Printf("This set may include too many or too few files. Please audit carefully.\n")
	fmt.Println()
	fmt.Printf("To pin this dependency, add the following to %s and fill in missing information.\n", configFile)
	fmt.Println()
	fmt.Printf("%s\n", string(out))
}

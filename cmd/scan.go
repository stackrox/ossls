package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/joshdk/licensor"
	"github.com/joshdk/licensor/spdx"
	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/audit"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/integrity"
	"gopkg.in/yaml.v2"
)

func ScanCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan",
		Short: "Scan a single dependency",
		RunE: func(_ *cobra.Command, args []string) error {
			dependencies, err := Scan(args)
			if err != nil {
				return err
			}

			ScanPrint(dependencies)
			return nil
		},
	}

	return c
}

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
			Files: make(map[string]config.ContentConfig),
		}
	)

	// Calculate checksums for the found files
	for _, file := range licenseFiles {
		frag, err := filepath.Rel(directory, file)
		if err != nil {
			return nil, err
		}

		content := config.ContentConfig{}

		body, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		switch filepath.Base(file) {
		case "package.json":
			content.FieldHashes = make(map[string]string)

			fields := make(map[string]interface{})

			if err := json.Unmarshal(body, &fields); err != nil {
				return nil, err
			}

			for _, key := range []string{"author", "authors", "license", "licenses", "contributors"} {
				data, found := fields[key]
				if !found {
					continue
				}

				checksum, err := integrity.ChecksumField(data)
				if err != nil {
					return nil, err
				}

				content.FieldHashes[key] = checksum
			}

		default:
			checksum := integrity.ChecksumBytes(body)
			content.FileHash = checksum
		}

		dependency.Files[frag] = content
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

func ScanPrint(dependencies map[string]config.Dependency) {
	dummy := map[string]map[string]config.Dependency{
		"dependencies": dependencies,
	}

	out, _ := yaml.Marshal(dummy)

	fmt.Printf("The given directories were scanned\n")
	fmt.Printf("for potentially relevant license and attribution files.\n")
	fmt.Printf("This set may include too many or too few files. Please audit carefully.\n")
	fmt.Println()
	fmt.Printf("To pin this dependency, add the following to the config file and fill in missing information.\n")
	fmt.Println()
	fmt.Printf("%s\n", string(out))
}

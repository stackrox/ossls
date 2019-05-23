package cmd

import (
	"fmt"
	"sort"

	"github.com/joshdk/licensor/spdx"
	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/config"
)

func NoticeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "notice",
		Short: "Generate license notice",
		RunE: func(c *cobra.Command, _ []string) error {
			configFlag, _ := c.Flags().GetString("config")
			cfg, err := config.Load(configFlag)
			if err != nil {
				return err
			}

			return PrintNotice(cfg)
		},
	}
}

func PrintNotice(cfg *config.Config) error {
	// Create a map of SPDX license identifiers (like "MIT") to a list of all
	// dependencies that use that license.
	byType := make(map[string][]config.Dependency)
	for _, dependency := range cfg.Dependencies {
		byType[dependency.License] = append(byType[dependency.License], dependency)
	}

	// Extract (and sort) the names of all used license identifiers. This is to
	// provide determinism in notice output.
	names := make([]string, 0, len(byType))
	for key, dependencies := range byType {
		sort.Slice(dependencies, func(i, j int) bool {
			return dependencies[i].URL < dependencies[j].URL
		})
		names = append(names, key)
	}
	sort.Strings(names)

	// Build a map of SPDX identifiers to their corresponding license, for
	// efficiency later on.
	var licenseMap = make(map[string]spdx.License)
	for _, license := range spdx.All() {
		licenseMap[license.Identifier] = license
	}

	for _, license := range names {
		var (
			dependencies = byType[license]
			body         = licenseMap[license].Text
		)

		// Print header and full body for this licence.
		fmt.Printf("The StackRox Platform uses the following components which are subject to the terms and conditions of the %s License:\n\n", license)
		fmt.Printf("%s\n\n", body)

		// Print each dependency url, and its list of copyright holders.
		for _, dependency := range dependencies {
			fmt.Printf("%s\n", dependency.URL)
			for _, attr := range dependency.Attribution {
				fmt.Printf("  Copyright %s\n", attr)
			}
		}
		fmt.Println()
	}

	return nil
}

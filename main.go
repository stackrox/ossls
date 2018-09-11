package main

import (
	"os"

	"flag"
	"fmt"

	"github.com/joshdk/licensor/spdx"
	"github.com/stackrox/ossls/audit"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/resolver"
)

var (
	version = "development"
)

func mainCmd() error {
	var (
		configFlag  = flag.String("config", ".ossls.yml", "Path to configuration file.")
		versionFlag = flag.Bool("version", false, "Displays the version and exits.")
	)
	flag.Parse()

	if *versionFlag == true {
		fmt.Println(version)
		return nil
	}

	cfg, err := config.Load(*configFlag)
	if err != nil {
		return err
	}

	goDeps, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return err
	}

	jsDeps, err := cfg.Resolvers.Js.Repos()
	if err != nil {
		return err
	}

	deps := make([]resolver.Dependency, 0, len(goDeps)+len(jsDeps))

	for _, dep := range goDeps {
		deps = append(deps, dep)
	}

	for _, dep := range jsDeps {
		deps = append(deps, dep)
	}

	results, err := audit.Dependencies(deps)
	if err != nil {
		return err
	}

	report(*configFlag, results)

	return nil
}

func report(header string, results map[resolver.Dependency]map[string][]spdx.License) {
	tree := func(index int) (string, string) {
		switch index {
		case 0:
			return "└── ", "    "
		default:
			return "├── ", "│   "
		}
	}

	fmt.Println(header)

	depIndex := len(results)
	for dep, files := range results {
		depIndex--
		depPrefix, depCont := tree(depIndex)
		fmt.Printf("%s%s\n", depPrefix, dep.Path)

		fileIndex := len(files)
		for file, licenses := range files {
			fileIndex--
			filePrefix, fileCont := tree(fileIndex)
			fmt.Printf("%s%s%s\n", depCont, filePrefix, file)

			licenseIndex := len(licenses)
			for _, license := range licenses {
				licenseIndex--
				licensePrefix, _ := tree(licenseIndex)
				fmt.Printf("%s%s%s%s\n", depCont, fileCont, licensePrefix, license.Name)
			}
		}
	}
}

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

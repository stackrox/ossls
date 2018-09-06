package main

import (
	"os"

	"flag"
	"fmt"

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

	fmt.Printf("Config: %+v\n", cfg)

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

	for index, dep := range deps {
		ref := dep.Version
		if ref == "" {
			ref = dep.Reference
		}
		fmt.Printf("[%d/%d] %s @%s (%s)\n", index+1, len(deps), dep.Name, ref, dep.Path)
	}

	return nil
}

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

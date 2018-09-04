package main

import (
	"os"

	"flag"
	"fmt"

	"github.com/stackrox/ossls/config"
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

	deps, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return err
	}

	for index, dep := range deps {
		ref := dep.Version
		if ref == "" {
			ref = dep.Reference
		}
		fmt.Printf("[%d/%d] %s @%s\n", index+1, len(deps), dep.Name, ref)
	}

	return nil
}

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

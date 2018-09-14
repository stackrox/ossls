package main

import (
	"os"

	"flag"
	"fmt"

	"errors"

	"github.com/stackrox/ossls/cmd"
	"github.com/stackrox/ossls/config"
)

var (
	version = "development"
)

func mainCmd() error {
	var (
		configFlag  = flag.String("config", ".ossls.yml", "Path to configuration file.")
		listFlag    = flag.Bool("list", false, "List all dependencies")
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

	switch {
	case *listFlag:
		names, err := cmd.List(cfg)
		if err != nil {
			return err
		}
		cmd.ListPrint(names)
		return nil

	default:
		return errors.New("no action given")
	}
}

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

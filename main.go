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
		checksumFlag = flag.Bool("checksum", false, "Calculate checksum for a file.")
		configFlag   = flag.String("config", ".ossls.yml", "Path to configuration file.")
		listFlag     = flag.Bool("list", false, "List all dependencies.")
		scanFlag     = flag.Bool("scan", false, "Scan single dependency.")
		versionFlag  = flag.Bool("version", false, "Displays the version and exits.")
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
	case *checksumFlag:
		filename := flag.Arg(0)
		checksum, err := cmd.Checksum(filename)
		if err != nil {
			return err
		}
		cmd.ChecksumPrint(filename, checksum)
		return nil

	case *listFlag:
		names, err := cmd.List(cfg)
		if err != nil {
			return err
		}
		cmd.ListPrint(names)
		return nil

	case *scanFlag:
		directory := flag.Arg(0)
		dep, err := cmd.Scan(directory)
		if err != nil {
			return err
		}
		cmd.ScanPrint(*configFlag, directory, dep)
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

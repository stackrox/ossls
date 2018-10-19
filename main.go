package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/stackrox/ossls/cmd"
	"github.com/stackrox/ossls/config"
)

var (
	version = "development"
)

func mainCmd() error {
	var (
		configFlag   = flag.String("config", ".ossls.yml", "Path to configuration file.")
		auditFlag    = flag.Bool("audit", false, "Audit all dependencies.")
		checksumFlag = flag.Bool("checksum", false, "Calculate checksum for a file.")
		listFlag     = flag.Bool("list", false, "List all dependencies.")
		scanFlag     = flag.Bool("scan", false, "Scan single dependency.")
		versionFlag  = flag.Bool("version", false, "Displays the version and exits.")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return nil
	}

	cfg, err := config.Load(*configFlag)
	if err != nil {
		return err
	}

	switch {
	case *auditFlag:
		violations, count, err := cmd.Audit(cfg)
		if err != nil {
			return err
		}
		cmd.AuditPrint(violations)

		switch count {
		case 0:
			return nil
		default:
			return errors.New("violations found")
		}

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
		directories := flag.Args()
		dependencies, err := cmd.Scan(directories)
		if err != nil {
			return err
		}
		cmd.ScanPrint(*configFlag, dependencies)
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

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/stackrox/ossls/cmd"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/integrity"
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
		noticeFlag   = flag.Bool("notice", false, "Generate license notice.")
		scanFlag     = flag.Bool("scan", false, "Scan single dependency.")
		versionFlag  = flag.Bool("version", false, "Displays the version and exits.")
		quietFlag    = flag.Bool("quiet", false, "Only print audit entries that fail.")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return nil
	}

	switch {
	case *auditFlag:
		cfg, err := config.Load(*configFlag)
		if err != nil {
			return err
		}

		violations, count, err := cmd.Audit(cfg)
		if err != nil {
			return err
		}
		cmd.AuditPrint(violations, *quietFlag)

		switch count {
		case 0:
			return nil
		default:
			return errors.New("violations found")
		}

	case *checksumFlag:
		var (
			filename = flag.Arg(0)
			field    = flag.Arg(1)
			checksum string
			err      error
		)
		switch field {
		case "":
			checksum, err = integrity.Checksum(filename)
		default:
			checksum, err = integrity.ChecksumFileField(filename, field)
		}
		if err != nil {
			return err
		}
		cmd.ChecksumPrint(filename, field, checksum)
		return nil

	case *listFlag:
		cfg, err := config.Load(*configFlag)
		if err != nil {
			return err
		}

		names, err := cmd.List(cfg)
		if err != nil {
			return err
		}
		cmd.ListPrint(names)
		return nil

	case *noticeFlag:
		cfg, err := config.Load(*configFlag)
		if err != nil {
			return err
		}

		err = cmd.PrintNotice(cfg)
		if err != nil {
			return err
		}
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

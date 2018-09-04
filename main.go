package main

import (
	"os"

	"flag"
	"fmt"

	"github.com/stackrox/ossls/config"
)

func mainCmd() error {
	configFlag := flag.String("config", ".ossls.yml", "")
	flag.Parse()

	cfg, err := config.Load(*configFlag)
	if err != nil {
		return err
	}

	fmt.Printf("Config: %+v\n", cfg)

	return nil
}

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

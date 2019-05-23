package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/cmd"
)

var (
	version = "development"
)

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "ossls: %s\n", err.Error())
		os.Exit(1)
	}
}

func mainCmd() error {
	c := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	c.PersistentFlags().StringP("config", "c", ".ossls.yml", "path to configuration file")

	c.AddCommand(cmd.AuditCommand())
	c.AddCommand(cmd.ChecksumCommand())
	c.AddCommand(cmd.ListCommand())
	c.AddCommand(cmd.NoticeCommand())
	c.AddCommand(cmd.ScanCommand())
	c.AddCommand(versionCommand())

	return c.Execute()
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Displays the version (%s) and exits", version),
		Run: func(*cobra.Command, []string) {
			fmt.Println(version)
		},
	}
}

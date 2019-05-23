package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/config"
)

func ListCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List all dependencies",
		RunE: func(c *cobra.Command, args []string) error {
			configFlag, _ := c.Flags().GetString("config")
			cfg, err := config.Load(configFlag)
			if err != nil {
				return err
			}

			names, err := List(cfg)
			if err != nil {
				return err
			}

			ListPrint(names)
			return nil
		},
	}

	return c
}

func List(cfg *config.Config) ([]string, error) {
	// Get list of Golang (via dep) dependencies
	goRepos, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return nil, err
	}

	// Get list of JavaScript (via package.json) dependencies
	jsRepos, err := cfg.Resolvers.Js.Repos()
	if err != nil {
		return nil, err
	}

	// Add dependency directory paths from both list together
	directories := []string{}
	for _, repo := range goRepos {
		directories = append(directories, repo)
	}
	for _, repo := range jsRepos {
		directories = append(directories, repo)
	}

	return directories, nil
}

func ListPrint(directories []string) {
	// Sort directory list
	sort.Strings(directories)

	for index, directory := range directories {
		fmt.Printf("[%d/%d] %s\n", index+1, len(directories), directory)
	}
}

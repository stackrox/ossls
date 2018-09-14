package cmd

import (
	"fmt"
	"sort"

	"github.com/stackrox/ossls/config"
)

func List(cfg *config.Config) ([]string, error) {
	// Get list of Golang (via dep) dependencies
	goDeps, err := cfg.Resolvers.Dep.Repos()
	if err != nil {
		return nil, err
	}

	// Get list of JavaScript (via package.json) dependencies
	jsDeps, err := cfg.Resolvers.Js.Repos()
	if err != nil {
		return nil, err
	}

	// Add dependency directory paths from both list together
	directories := []string{}
	for _, dep := range goDeps {
		directories = append(directories, dep.Path)
	}
	for _, dep := range jsDeps {
		directories = append(directories, dep.Path)
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

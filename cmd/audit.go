package cmd

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/config"
	"github.com/stackrox/ossls/resolver"
)

func AuditCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "audit",
		Short: "Audit all dependencies",
		RunE: func(c *cobra.Command, _ []string) error {
			quietFlag, _ := c.Flags().GetBool("quiet")
			configFlag, _ := c.Flags().GetString("config")
			exportFlag, _ := c.Flags().GetString("export")
			cfg, err := config.Load(configFlag)
			if err != nil {
				return errors.Wrap(err, "failed to load configuration file "+configFlag)
			}

			yarnProjects, err := resolver.ProjectsFromYarnLockfile(cfg.Yarn.Lockfile)
			if err != nil {
				return errors.Wrap(err, "failed to discover dependencies from yarn lockfile "+cfg.Yarn.Lockfile)
			}

			depProjects, err := resolver.ProjectsFromDepLockfile(cfg.Dep.Lockfile)
			if err != nil {
				return errors.Wrap(err, "failed to discover dependencies from dep lockfile "+cfg.Dep.Lockfile)
			}

			yarnResolved, err := resolver.LocateProjects(cfg.Yarn.NodeModulesDir, yarnProjects)
			if err != nil {
				return errors.Wrap(err, "failed to locate js dependencies in dir "+cfg.Yarn.NodeModulesDir)
			}

			depResolved, err := resolver.LocateProjects(cfg.Dep.VendorDir, depProjects)
			if err != nil {
				return errors.Wrap(err, "failed to locate go dependencies in dir "+cfg.Dep.VendorDir)
			}

			dependencies := joinDeps(cfg.Patterns, yarnResolved, depResolved)

			var failures bool
			for _, dependency := range dependencies {
				var err error
				if exportFlag != "" {
					err = export(dependency, exportFlag)
				}

				if err != nil {
					failures = true
					color.Red("✗ %s (%s)", dependency.Name, dependency.SourceDir)
					color.Yellow("  ↳ %v", err)
				} else if !quietFlag {
					color.Green("✓ %s (%s)", dependency.Name, dependency.SourceDir)
					for _, file := range dependency.Files {
						color.Blue("  ↳ %s/%s", dependency.Alias, filepath.Base(file))
					}
				}
			}

			if failures {
				return errors.New("failed to audit dependencies")
			}
			return nil
		},
	}

	c.Flags().BoolP("quiet", "q", false, "only display audit entries that fail")
	c.Flags().StringP("export", "x", "", "")

	return c
}

func joinDeps(patterns []string, sets ...map[string]string) []Dependency {
	var total int
	for _, set := range sets {
		total += len(set)
	}

	dependencies := make([]Dependency, 0, total)

	for _, set := range sets {
		for name, path := range set {
			files := resolver.FindLicenseFiles(path, patterns)
			dependency := Dependency{
				Name:      name,
				Alias:     flattenName(name),
				Files:     files,
				SourceDir: path,
			}
			dependencies = append(dependencies, dependency)
		}
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})

	for _, set := range sets {
		total += len(set)
	}

	return dependencies
}

func export(dependency Dependency, destination string) error {
	if err := os.MkdirAll(filepath.Join(destination, dependency.Alias), 0755); err != nil {
		return err
	}

	for _, file := range dependency.Files {
		if err := copyFileContents(
			file,
			filepath.Join(destination, dependency.Alias, filepath.Base(file)),
		); err != nil {
			return err
		}
	}

	return nil
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

type Dependency struct {
	Name      string
	Alias     string
	Files     []string
	SourceDir string
}

func flattenName(name string) string {
	return strings.Replace(name, "/", "-", -1)
}

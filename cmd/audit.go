package cmd

import (
	"fmt"
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

// AuditCommand implements an audit of dependencies
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

			var yarnProjects []resolver.Project
			if cfg.Yarn.Lockfile != "" {
				yarnProjects, err = resolver.ProjectsFromYarnLockfile(cfg.Yarn.Lockfile)
				if err != nil {
					return errors.Wrap(err, "failed to discover dependencies from yarn lockfile "+cfg.Yarn.Lockfile)
				}
			}

			var depProjects []resolver.Project
			if cfg.Dep.Lockfile != "" {
				depProjects, err = resolver.ProjectsFromDepLockfile(cfg.Dep.Lockfile)
				if err != nil {
					return errors.Wrap(err, "failed to discover dependencies from dep lockfile "+cfg.Dep.Lockfile)
				}
			}

			var goModProjects []resolver.GoModProject
			if cfg.GoMod.GoModFile != "" {
				goModProjects, err = resolver.ProjectsFromGoModFile(cfg.GoMod.GoModFile)
				if err != nil {
					return errors.Wrapf(err, "failed to discover dependencies from go.mod file %s", cfg.GoMod.GoModFile)
				}
			}

			var yarnResolved = make(map[string]resolver.Dependency)
			if len(yarnProjects) > 0 {
				dirList := cfg.Yarn.NodeModulesDirs
				fmt.Printf("Processing JS deps directories: %v \n", dirList)
				currentDeps, err := resolver.LocateProjects(dirList, yarnProjects)
				if err != nil {
					return errors.Wrapf(err, "failed to locate js dependencies in dirs %v", dirList)
				}
				for _, v := range currentDeps {
					fmt.Printf("Target dependency: %s \n", v)
					keyWithVersion := v.Name + v.Version
					fmt.Printf("Target dependency key with version: %s \n", keyWithVersion)

					yarnResolved[keyWithVersion] = v
				}

			}

			var depResolved map[string]resolver.Dependency
			if len(depProjects) > 0 {
				depResolved, err = resolver.LocateProjects(cfg.Dep.VendorDirs, depProjects)
				fmt.Printf("Resolved dependency: %s \n", depResolved)

				if err != nil {
					return errors.Wrapf(err, "failed to locate go dependencies in dirs %v", cfg.Dep.VendorDirs)
				}
			}

			var goModResolved map[string]resolver.Dependency
			if len(goModProjects) > 0 {
				goModResolved, err = resolver.LocateGoModProjects(goModProjects)
				if err != nil {
					return errors.Wrap(err, "failed to locate gomod dependencies")
				}
			}

			dependencies := joinDeps(cfg.Patterns, yarnResolved, depResolved, goModResolved)

			var failures bool
			for _, dependency := range dependencies {
				var err error
				if exportFlag != "" {
					err = export(dependency, exportFlag)
				}

				if err != nil {
					failures = true
					color.Red("✗ %s @%s (%s)", dependency.Name, dependency.Version, dependency.SourceDir)
					color.Yellow("  ↳ %v", err)
				} else if !quietFlag {
					color.Green("✓ %s @%s (%s)", dependency.Name, dependency.Version, dependency.SourceDir)
					for _, file := range dependency.Files {
						color.Blue("  ↳ %s/%s", dependency.Alias, filepath.Base(file))
					}
				}
			}
			if exportFlag != "" {
				if err := exportManifest(exportFlag, dependencies); err != nil {
					return err
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

func joinDeps(patterns []string, sets ...map[string]resolver.Dependency) []resolver.Dependency {
	var total int
	for _, set := range sets {
		total += len(set)
	}

	dependencies := make([]resolver.Dependency, 0, total)

	for _, set := range sets {
		for name, dependency := range set {
			files := resolver.FindLicenseFiles(dependency.SourceDir, patterns)
			dependency.Alias = flattenName(name)
			dependency.Files = files
			dependencies = append(dependencies, dependency)
		}
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	return dependencies
}

func export(dependency resolver.Dependency, destination string) error {
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

func exportManifest(destination string, dependencies []resolver.Dependency) error {
	manifestFile := filepath.Join(destination, "manifest.csv")
	file, err := os.OpenFile(manifestFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "Name,Version,Directory")
	for _, dep := range dependencies {
		fmt.Fprintf(file, "%s,%s,./%s\n", dep.Name, dep.Version, dep.Alias)
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

func flattenName(name string) string {
	return strings.Replace(name, "/", "-", -1)
}

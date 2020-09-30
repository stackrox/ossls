package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

			dependencies, err := joinDeps(cfg.PatternConfig, yarnResolved, depResolved, goModResolved)
			if err != nil {
				return errors.Wrap(err, "resolving dependencies")
			}

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

func joinDeps(patterns config.PatternConfig, sets ...map[string]resolver.Dependency) ([]resolver.Dependency, error) {
	var total int
	for _, set := range sets {
		total += len(set)
	}

	matcher, err := resolver.CompilePatternConfig(patterns)
	if err != nil {
		return nil, errors.Wrap(err, "compiling patterns")
	}

	dependencies := make([]resolver.Dependency, 0, total)

	for _, set := range sets {
		for name, dependency := range set {
			files, err := resolver.FindLicenseFiles(dependency.SourceDir, matcher)
			if err != nil {
				return nil, errors.Wrapf(err, "finding license files in directory %s", dependency.SourceDir)
			}
			dependency.Alias = flattenName(name)
			dependency.Files = files
			dependencies = append(dependencies, dependency)
		}
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	return dependencies, nil
}

func export(dependency resolver.Dependency, destination string) error {
	if err := os.MkdirAll(filepath.Join(destination, dependency.Alias), 0755); err != nil {
		return err
	}

	for _, file := range dependency.Files {
		if err := exportDependencyFile(
			file,
			filepath.Join(destination, dependency.Alias),
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

func exportDependencyFile(src, dstDir string) error {
	dstFile := filepath.Base(src)
	// Do not directly copy the package json file to avoid false positives
	// from image scanners for developer dependencies -- only export a subset of fields.
	// Multiple versions of the same package are exported into their own
	// directory, with separate licenses.
	if strings.ToLower(dstFile) == "package.json" {
		dstFile = "license-info.json"
		return copyPackageJsonContents(src, filepath.Join(dstDir, dstFile))
	}
	return copyFileContents(src, filepath.Join(dstDir, dstFile))
}

func copyJsonFieldIfExists(fieldName string, in, out map[string]interface{}) {
	if field, ok := in[fieldName]; ok {
		out[fieldName] = field
	}
}

func JSONMarshalIndentWithoutEscape(t interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	if err != nil {
		return nil, err
	}
	indentBuf := &bytes.Buffer{}
	err = json.Indent(indentBuf, buf.Bytes(), "", "  ")
	if err != nil {
		return nil, err
	}
	return indentBuf.Bytes(), err
}

func copyPackageJsonContents(packageJsonFile, licenseInfoJsonFile string) error {
	pkgJsonData, err := ioutil.ReadFile(packageJsonFile)
	if err != nil {
		return err
	}

	// Note: Unmarshal package.json file as unstructured json because some packages may represent license
	// in a deprecated form using an array of license objects instead of a SPDX format string
	// Format details: https://docs.npmjs.com/files/package.json#license
	var inputData map[string]interface{}
	if err = json.Unmarshal(pkgJsonData, &inputData); err != nil {
		return err
	}

	type licenseInfo struct {
		License  interface{} `json:"license"`
		Metadata interface{} `json:"metadata"`
	}

	metadata := make(map[string]interface{})
	copyJsonFieldIfExists("name", inputData, metadata)
	copyJsonFieldIfExists("author", inputData, metadata)
	copyJsonFieldIfExists("contributors", inputData, metadata)
	copyJsonFieldIfExists("repository", inputData, metadata)

	outputData := licenseInfo{}
	if license, ok := inputData["license"]; ok {
		outputData.License = license
	}
	outputData.Metadata = metadata

	bytes, err := JSONMarshalIndentWithoutEscape(outputData)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(licenseInfoJsonFile, bytes, 0644); err != nil {
		return err
	}
	return nil
}

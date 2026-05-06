package cmd

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stackrox/ossls/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions

func createTestDependency(t *testing.T, name, version string, files map[string]string) resolver.Dependency {
	t.Helper()
	sourceDir := t.TempDir()
	filePaths := make([]string, 0, len(files))

	for filename, content := range files {
		path := filepath.Join(sourceDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	return resolver.Dependency{
		Name:      name,
		Alias:     strings.ReplaceAll(name, "/", "-"),
		Version:   version,
		SourceDir: sourceDir,
		Files:     filePaths,
	}
}

func verifyFileTimestamp(t *testing.T, path string, expected time.Time) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, expected.Unix(), info.ModTime().Unix(),
		"timestamp mismatch for %s", path)
}

func readManifestCSV(t *testing.T, manifestPath string) [][]string {
	t.Helper()
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	require.NoError(t, err)

	return rows
}

func computeSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Tests

func TestGetExportTimestamp(t *testing.T) {
	tests := []struct {
		name            string
		sourceDateEpoch string
		setEnv          bool
		expectNow       bool
		expectedTime    time.Time
	}{
		{
			name:            "valid epoch returns correct time",
			sourceDateEpoch: "1609459200",
			setEnv:          true,
			expectNow:       false,
			expectedTime:    time.Unix(1609459200, 0),
		},
		{
			name:            "unset env var returns current time",
			sourceDateEpoch: "",
			setEnv:          false,
			expectNow:       true,
			expectedTime:    time.Time{},
		},
		{
			name:            "invalid epoch non-numeric",
			sourceDateEpoch: "not-a-number",
			setEnv:          true,
			expectNow:       true,
			expectedTime:    time.Time{},
		},
		{
			name:            "zero epoch returns Unix epoch",
			sourceDateEpoch: "0",
			setEnv:          true,
			expectNow:       false,
			expectedTime:    time.Unix(0, 0),
		},
		{
			name:            "large valid epoch",
			sourceDateEpoch: "2147483647",
			setEnv:          true,
			expectNow:       false,
			expectedTime:    time.Unix(2147483647, 0),
		},
		{
			name:            "invalid epoch with decimal",
			sourceDateEpoch: "160945920.5",
			setEnv:          true,
			expectNow:       true,
			expectedTime:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("SOURCE_DATE_EPOCH", tt.sourceDateEpoch)
			} else {
				os.Unsetenv("SOURCE_DATE_EPOCH")
			}

			result := getExportTimestamp()

			if tt.expectNow {
				assert.WithinDuration(t, time.Now(), result, 1*time.Second)
			} else {
				assert.Equal(t, tt.expectedTime.Unix(), result.Unix())
			}
		})
	}
}

func TestCopyPackageJsonContents(t *testing.T) {
	tests := []struct {
		name             string
		inputJSON        string
		expectedLicense  interface{}
		expectedMetadata map[string]interface{}
		excludedFields   []string
	}{
		{
			name:      "standard with all fields",
			inputJSON: "testdata/package.json",
			expectedLicense: "MIT",
			expectedMetadata: map[string]interface{}{
				"name":   "test-package",
				"author": "Test Author",
				"contributors": []interface{}{
					"Contributor 1",
					"Contributor 2",
				},
				"repository": map[string]interface{}{
					"type": "git",
					"url":  "https://github.com/test/test-package.git",
				},
			},
			excludedFields: []string{"devDependencies", "scripts", "version"},
		},
		{
			name: "minimal package.json",
			inputJSON: `{
				"name": "minimal-package",
				"license": "Apache-2.0"
			}`,
			expectedLicense: "Apache-2.0",
			expectedMetadata: map[string]interface{}{
				"name": "minimal-package",
			},
			excludedFields: []string{"version", "author", "devDependencies"},
		},
		{
			name: "without optional fields",
			inputJSON: `{
				"name": "no-optional-fields",
				"version": "2.0.0",
				"license": "BSD-3-Clause"
			}`,
			expectedLicense: "BSD-3-Clause",
			expectedMetadata: map[string]interface{}{
				"name": "no-optional-fields",
			},
			excludedFields: []string{"version", "author", "contributors", "repository"},
		},
		{
			name: "legacy license format array",
			inputJSON: `{
				"name": "legacy-license",
				"license": [
					{
						"type": "MIT",
						"url": "https://opensource.org/licenses/MIT"
					}
				]
			}`,
			expectedLicense: []interface{}{
				map[string]interface{}{
					"type": "MIT",
					"url":  "https://opensource.org/licenses/MIT",
				},
			},
			expectedMetadata: map[string]interface{}{
				"name": "legacy-license",
			},
			excludedFields: []string{"version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcFile := tt.inputJSON
			isPath := !strings.Contains(tt.inputJSON, "{")

			if !isPath {
				tmpFile := filepath.Join(t.TempDir(), "package.json")
				err := os.WriteFile(tmpFile, []byte(tt.inputJSON), 0644)
				require.NoError(t, err)
				srcFile = tmpFile
			}

			dstFile := filepath.Join(t.TempDir(), "license-info.json")

			err := copyPackageJsonContents(srcFile, dstFile)
			require.NoError(t, err)

			var output map[string]interface{}
			data, err := os.ReadFile(dstFile)
			require.NoError(t, err)
			err = json.Unmarshal(data, &output)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedLicense, output["license"])

			expectedMetadata := tt.expectedMetadata
			actualMetadata, ok := output["metadata"].(map[string]interface{})
			require.True(t, ok, "metadata should be a map")

			for key, expectedValue := range expectedMetadata {
				actualValue, exists := actualMetadata[key]
				assert.True(t, exists, "metadata should contain %s", key)
				assert.Equal(t, expectedValue, actualValue, "metadata.%s value mismatch", key)
			}

			for _, field := range tt.excludedFields {
				assert.NotContains(t, output, field, "excluded field %s should not be in root", field)
				assert.NotContains(t, actualMetadata, field, "excluded field %s should not be in metadata", field)
			}

			assert.NotContains(t, string(data), "\\u003c", "JSON should not contain HTML-escaped <")
			assert.NotContains(t, string(data), "\\u003e", "JSON should not contain HTML-escaped >")
			assert.Contains(t, string(data), "  ", "JSON should be indented")
		})
	}
}

func TestExportDependencyFile(t *testing.T) {
	testTimestamp := time.Unix(1609459200, 0)

	tests := []struct {
		name            string
		srcFilename     string
		srcContent      string
		expectedDstName string
		checkContent    func(t *testing.T, dstPath string)
	}{
		{
			name:            "regular license file",
			srcFilename:     "LICENSE",
			srcContent:      "MIT License\nCopyright (c) 2024",
			expectedDstName: "LICENSE",
			checkContent: func(t *testing.T, dstPath string) {
				content, err := os.ReadFile(dstPath)
				require.NoError(t, err)
				assert.Equal(t, "MIT License\nCopyright (c) 2024", string(content))
			},
		},
		{
			name:            "markdown license",
			srcFilename:     "LICENSE.md",
			srcContent:      "# MIT License\n\n**Copyright (c) 2024**",
			expectedDstName: "LICENSE.md",
			checkContent: func(t *testing.T, dstPath string) {
				content, err := os.ReadFile(dstPath)
				require.NoError(t, err)
				assert.Equal(t, "# MIT License\n\n**Copyright (c) 2024**", string(content))
			},
		},
		{
			name:            "package.json lowercase",
			srcFilename:     "package.json",
			srcContent:      `{"name": "test", "license": "MIT"}`,
			expectedDstName: "license-info.json",
			checkContent: func(t *testing.T, dstPath string) {
				var licenseInfo map[string]interface{}
				data, err := os.ReadFile(dstPath)
				require.NoError(t, err)
				err = json.Unmarshal(data, &licenseInfo)
				require.NoError(t, err)
				assert.Contains(t, licenseInfo, "license")
				assert.Contains(t, licenseInfo, "metadata")
			},
		},
		{
			name:            "PACKAGE.JSON uppercase",
			srcFilename:     "PACKAGE.JSON",
			srcContent:      `{"name": "test-upper", "license": "Apache-2.0"}`,
			expectedDstName: "license-info.json",
			checkContent: func(t *testing.T, dstPath string) {
				var licenseInfo map[string]interface{}
				data, err := os.ReadFile(dstPath)
				require.NoError(t, err)
				err = json.Unmarshal(data, &licenseInfo)
				require.NoError(t, err)
				assert.Equal(t, "Apache-2.0", licenseInfo["license"])
			},
		},
		{
			name:            "Package.json mixed case",
			srcFilename:     "Package.json",
			srcContent:      `{"name": "test-mixed", "license": "BSD"}`,
			expectedDstName: "license-info.json",
			checkContent: func(t *testing.T, dstPath string) {
				var licenseInfo map[string]interface{}
				data, err := os.ReadFile(dstPath)
				require.NoError(t, err)
				err = json.Unmarshal(data, &licenseInfo)
				require.NoError(t, err)
				assert.Equal(t, "BSD", licenseInfo["license"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcDir := t.TempDir()
			srcPath := filepath.Join(srcDir, tt.srcFilename)
			err := os.WriteFile(srcPath, []byte(tt.srcContent), 0644)
			require.NoError(t, err)

			dstDir := t.TempDir()

			err = exportDependencyFile(srcPath, dstDir, testTimestamp)
			require.NoError(t, err)

			dstPath := filepath.Join(dstDir, tt.expectedDstName)
			info, err := os.Stat(dstPath)
			require.NoError(t, err, "destination file should exist")

			assert.Equal(t, testTimestamp.Unix(), info.ModTime().Unix(),
				"file timestamp should match expected")

			if tt.checkContent != nil {
				tt.checkContent(t, dstPath)
			}
		})
	}
}

func TestExport(t *testing.T) {
	testTimestamp := time.Unix(1609459200, 0)

	tests := []struct {
		name       string
		dependency resolver.Dependency
		files      map[string]string
	}{
		{
			name: "single dependency with multiple files",
			files: map[string]string{
				"LICENSE":    "MIT License",
				"README.md":  "# Test Package",
				"LICENSE.md": "# MIT License",
			},
		},
		{
			name: "dependency with package.json",
			files: map[string]string{
				"package.json": `{"name": "test-pkg", "license": "MIT"}`,
				"LICENSE":      "MIT License",
			},
		},
		{
			name: "scoped package",
			files: map[string]string{
				"LICENSE": "Apache License",
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := createTestDependency(t, "test-package", "1.0.0", tt.files)

			if i == 2 {
				dep.Name = "@babel/core"
				dep.Alias = "@babel-core"
			}

			destRoot := t.TempDir()

			err := export(dep, destRoot, testTimestamp)
			require.NoError(t, err)

			destDir := filepath.Join(destRoot, dep.Alias)
			info, err := os.Stat(destDir)
			require.NoError(t, err)
			assert.True(t, info.IsDir())

			for srcFilename := range tt.files {
				expectedDst := srcFilename
				if strings.ToLower(srcFilename) == "package.json" {
					expectedDst = "license-info.json"
				}

				destPath := filepath.Join(destDir, expectedDst)
				info, err := os.Stat(destPath)
				require.NoError(t, err, "file %s should exist", expectedDst)
				assert.Equal(t, testTimestamp.Unix(), info.ModTime().Unix(),
					"timestamp for %s should match", expectedDst)
			}
		})
	}
}

func TestExportManifest(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []resolver.Dependency
		expectedCSV  string
	}{
		{
			name: "alphabetical sorting case insensitive",
			dependencies: []resolver.Dependency{
				{Name: "zebra", Version: "1.0.0", Alias: "zebra"},
				{Name: "Apple", Version: "2.0.0", Alias: "Apple"},
				{Name: "banana", Version: "1.5.0", Alias: "banana"},
			},
			expectedCSV: "Name,Version,Directory\nApple,2.0.0,./Apple\nbanana,1.5.0,./banana\nzebra,1.0.0,./zebra\n",
		},
		{
			name: "same name different versions",
			dependencies: []resolver.Dependency{
				{Name: "package", Version: "2.0.0", Alias: "package2.0.0"},
				{Name: "package", Version: "1.0.0", Alias: "package1.0.0"},
				{Name: "package", Version: "1.5.0", Alias: "package1.5.0"},
			},
			expectedCSV: "Name,Version,Directory\npackage,1.0.0,./package1.0.0\npackage,1.5.0,./package1.5.0\npackage,2.0.0,./package2.0.0\n",
		},
		{
			name: "mixed case names",
			dependencies: []resolver.Dependency{
				{Name: "React", Version: "1.0.0", Alias: "React"},
				{Name: "REACT", Version: "2.0.0", Alias: "REACT"},
				{Name: "react-dom", Version: "1.0.0", Alias: "react-dom"},
			},
			expectedCSV: "Name,Version,Directory\nReact,1.0.0,./React\nREACT,2.0.0,./REACT\nreact-dom,1.0.0,./react-dom\n",
		},
		{
			name: "scoped packages",
			dependencies: []resolver.Dependency{
				{Name: "@babel/core", Version: "7.0.0", Alias: "@babel-core"},
				{Name: "@angular/core", Version: "12.0.0", Alias: "@angular-core"},
				{Name: "lodash", Version: "4.17.21", Alias: "lodash"},
			},
			expectedCSV: "Name,Version,Directory\n@angular/core,12.0.0,./@angular-core\n@babel/core,7.0.0,./@babel-core\nlodash,4.17.21,./lodash\n",
		},
		{
			name:         "empty dependencies",
			dependencies: []resolver.Dependency{},
			expectedCSV:  "Name,Version,Directory\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destDir := t.TempDir()

			err := exportManifest(destDir, tt.dependencies)
			require.NoError(t, err)

			manifestPath := filepath.Join(destDir, "manifest.csv")
			content, err := os.ReadFile(manifestPath)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCSV, string(content))
		})
	}

	t.Run("determinism test", func(t *testing.T) {
		dependencies := []resolver.Dependency{
			{Name: "@emotion/hash", Version: "0.6.6", Alias: "@emotion-hash0.6.6"},
			{Name: "@emotion/hash", Version: "0.8.0", Alias: "@emotion-hash0.8.0"},
			{Name: "@emotion/memoize", Version: "0.6.6", Alias: "@emotion-memoize0.6.6"},
			{Name: "@emotion/memoize", Version: "0.7.1", Alias: "@emotion-memoize0.7.1"},
			{Name: "@emotion/memoize", Version: "0.7.4", Alias: "@emotion-memoize0.7.4"},
		}

		checksums := make(map[string]int)
		for i := 0; i < 5; i++ {
			destDir := t.TempDir()
			err := exportManifest(destDir, dependencies)
			require.NoError(t, err)

			manifestPath := filepath.Join(destDir, "manifest.csv")
			hash := computeSHA256(t, manifestPath)
			checksums[hash]++
		}

		assert.Equal(t, 1, len(checksums), "manifest must be deterministic - all runs should produce identical output")
	})
}

func TestReproducibleExport(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "1609459200")
	timestamp := getExportTimestamp()

	dependencies := []resolver.Dependency{
		createTestDependency(t, "package-a", "1.0.0", map[string]string{
			"LICENSE": "MIT License",
		}),
		createTestDependency(t, "package-b", "2.0.0", map[string]string{
			"LICENSE.md":   "# Apache License",
			"package.json": `{"name": "package-b", "license": "Apache-2.0"}`,
		}),
		createTestDependency(t, "@scoped/package", "3.0.0", map[string]string{
			"LICENSE": "BSD License",
		}),
	}
	dependencies[2].Alias = "@scoped-package"

	dest1 := t.TempDir()
	dest2 := t.TempDir()

	for _, dep := range dependencies {
		require.NoError(t, export(dep, dest1, timestamp))
		require.NoError(t, export(dep, dest2, timestamp))
	}

	require.NoError(t, exportManifest(dest1, dependencies))
	require.NoError(t, exportManifest(dest2, dependencies))

	err := filepath.Walk(dest1, func(path1 string, info1 os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info1.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(dest1, path1)
		path2 := filepath.Join(dest2, relPath)

		content1, err := os.ReadFile(path1)
		require.NoError(t, err)
		content2, err := os.ReadFile(path2)
		require.NoError(t, err)
		assert.Equal(t, content1, content2, "file %s differs", relPath)

		info2, err := os.Stat(path2)
		require.NoError(t, err)
		assert.Equal(t, info1.ModTime().Unix(), info2.ModTime().Unix(),
			"timestamp differs for %s", relPath)

		return nil
	})
	require.NoError(t, err)
}

func TestExportDependencyFileErrors(t *testing.T) {
	testTimestamp := time.Unix(1609459200, 0)

	tests := []struct {
		name        string
		setup       func(t *testing.T) (src, dst string)
		expectError bool
		errorMatch  string
	}{
		{
			name: "source file does not exist",
			setup: func(t *testing.T) (string, string) {
				return "/nonexistent/file.txt", t.TempDir()
			},
			expectError: true,
			errorMatch:  "no such file",
		},
		{
			name: "invalid package.json",
			setup: func(t *testing.T) (string, string) {
				srcDir := t.TempDir()
				srcFile := filepath.Join(srcDir, "package.json")
				err := os.WriteFile(srcFile, []byte("{ invalid json }"), 0644)
				require.NoError(t, err)
				return srcFile, t.TempDir()
			},
			expectError: true,
			errorMatch:  "invalid character",
		},
		{
			name: "valid regular file export",
			setup: func(t *testing.T) (string, string) {
				srcDir := t.TempDir()
				srcFile := filepath.Join(srcDir, "LICENSE")
				err := os.WriteFile(srcFile, []byte("MIT License"), 0644)
				require.NoError(t, err)
				return srcFile, t.TempDir()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup(t)
			err := exportDependencyFile(src, dst, testTimestamp)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMatch)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

package yarn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePropertyLine(t *testing.T) {
	tests := []struct {
		title         string
		line          string
		expectedName  string
		expectedValue string
	}{
		{
			title:         "quoted name",
			line:          `"@babel/types" "7.0.0-beta.31"`,
			expectedName:  "@babel/types",
			expectedValue: "7.0.0-beta.31",
		},
		{
			title:         "plain name",
			line:          `version "7.0.0-beta.31"`,
			expectedName:  "version",
			expectedValue: "7.0.0-beta.31",
		},
		{
			title:         "plain name with spaces",
			line:          `    version    "7.0.0-beta.31"    `,
			expectedName:  "version",
			expectedValue: "7.0.0-beta.31",
		},
		{
			title:         "quoted name with spaces",
			line:          `    "version"    7.0.0-beta.31    `,
			expectedName:  "version",
			expectedValue: "7.0.0-beta.31",
		},
		{
			title:         "property with url",
			line:          `resolved "https://registry.yarnpkg.com/@babel/types/-/types-7.0.0-beta.31.tgz#42c9c86784f674c173fb21882ca9643334029de4"`,
			expectedName:  "resolved",
			expectedValue: "https://registry.yarnpkg.com/@babel/types/-/types-7.0.0-beta.31.tgz#42c9c86784f674c173fb21882ca9643334029de4",
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d-%s", index, test.title)
		t.Run(name, func(t *testing.T) {
			actualName, actualValue := parsePropertyLine(test.line)
			assert.Equal(t, test.expectedName, actualName)
			assert.Equal(t, test.expectedValue, actualValue)
		})
	}
}

func TestParseSpecLine(t *testing.T) {
	tests := []struct {
		title            string
		line             string
		expectedName     string
		expectedVersions []string
	}{
		{
			title:            "quoted spec",
			line:             `"@babel/types@7.0.0-beta.31":`,
			expectedName:     "@babel/types",
			expectedVersions: []string{"7.0.0-beta.31"},
		},
		{
			title:            "quoted spec with spaces",
			line:             `  "@babel/types@7.0.0-beta.31"  :  `,
			expectedName:     "@babel/types",
			expectedVersions: []string{"7.0.0-beta.31"},
		},
		{
			title:            "quoted spec with more spaces",
			line:             `  "@babel/types   @   7.0.0-beta.31"  :  `,
			expectedName:     "@babel/types",
			expectedVersions: []string{"7.0.0-beta.31"},
		},
		{
			title:            "plain spec",
			line:             `argparse@^1.0.7:`,
			expectedName:     "argparse",
			expectedVersions: []string{"^1.0.7"},
		},
		{
			title:            "plain spec with spaces",
			line:             `  argparse  @  ^1.0.7  :  `,
			expectedName:     "argparse",
			expectedVersions: []string{"^1.0.7"},
		},
		{
			title:        "multi spec",
			line:         `async@^1.4.0, async@^1.5.2:`,
			expectedName: "async",
			expectedVersions: []string{
				"^1.4.0",
				"^1.5.2",
			},
		},
		{
			title:        "complex multi spec",
			line:         `"semver@2 || 3 || 4 || 5", semver@^5.0.3, semver@^5.1.0, semver@^5.3.0:`,
			expectedName: "semver",
			expectedVersions: []string{
				"2 || 3 || 4 || 5",
				"^5.0.3",
				"^5.1.0",
				"^5.3.0",
			},
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d-%s", index, test.title)
		t.Run(name, func(t *testing.T) {
			actualName, actualVersions := parseSpecLine(test.line)
			assert.Equal(t, test.expectedName, actualName)
			assert.Equal(t, test.expectedVersions, actualVersions)
		})
	}
}

func TestLineType(t *testing.T) {
	tests := []struct {
		title        string
		line         string
		expectedKind kind
	}{
		{
			title:        "empty line",
			expectedKind: Blank,
		},
		{
			title:        "line with only spaces",
			line:         "   ",
			expectedKind: Blank,
		},
		{
			title:        "comment line",
			line:         "# yarn lockfile v1",
			expectedKind: Comment,
		},
		{
			title:        "comment line with leading spaces",
			line:         "   # yarn lockfile v1",
			expectedKind: Comment,
		},
		{
			title:        "plain spec",
			line:         `accepts@~1.3.4:`,
			expectedKind: Spec,
		},
		{
			title:        "quoted spec",
			line:         `"@babel/code-frame@7.0.0-beta.31":`,
			expectedKind: Spec,
		},
		{
			title:        "version property",
			line:         `  version "7.0.0-beta.31"`,
			expectedKind: Property,
		},
		{
			title:        "dependencies header",
			line:         `  dependencies:`,
			expectedKind: Header,
		},
		{
			title:        "optional dependencies header",
			line:         `  optionalDependencies:`,
			expectedKind: Header,
		},
		{
			title:        "plain dependency",
			line:         `    chalk "^2.0.0"`,
			expectedKind: Dep,
		},
		{
			title:        "quoted dependency",
			line:         `    "@babel/types" "7.0.0-beta.31"`,
			expectedKind: Dep,
		},
		{
			title:        "resolved property",
			line:         `  resolved "https://registry.yarnpkg.com/accepts/-/accepts-1.3.4.tgz#86246758c7dd6d21a6474ff084a4740ec05eb21f"`,
			expectedKind: Property,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d-%s", index, test.title)
		t.Run(name, func(t *testing.T) {
			actualKind := lineType(test.line)
			assert.Equal(t, test.expectedKind, actualKind)
		})
	}
}

func TestFullParse(t *testing.T) {
	expectedEntries := []Entry{
		{
			Name:     "@babel/code-frame",
			Resolved: "https://registry.yarnpkg.com/@babel/code-frame/-/code-frame-7.0.0-beta.31.tgz#473d021ecc573a2cce1c07d5b509d5215f46ba35",
			Specs:    []string{"7.0.0-beta.31"},
			Dependencies: []Dependency{
				{"chalk", "^2.0.0"},
				{"esutils", "^2.0.2"},
				{"js-tokens", "^3.0.0"},
			},
			Version: "7.0.0-beta.31",
		},
		{
			Name:     "@babel/helper-function-name",
			Resolved: "https://registry.yarnpkg.com/@babel/helper-function-name/-/helper-function-name-7.0.0-beta.31.tgz#afe63ad799209989348b1109b44feb66aa245f57",
			Specs:    []string{"7.0.0-beta.31"},
			Dependencies: []Dependency{
				{"@babel/helper-get-function-arity", "7.0.0-beta.31"},
				{"@babel/template", "7.0.0-beta.31"},
				{"@babel/types", "7.0.0-beta.31"},
			},
			Version: "7.0.0-beta.31",
		},
		{
			Name:     "@babel/helper-module-imports",
			Resolved: "https://registry.yarnpkg.com/@babel/helper-module-imports/-/helper-module-imports-7.0.0-beta.32.tgz#8126fc024107c226879841b973677a4f4e510a03",
			Specs:    []string{"7.0.0-beta.32"},
			Dependencies: []Dependency{
				{"@babel/types", "7.0.0-beta.32"},
				{"lodash", "^4.2.0"},
			},
			Version: "7.0.0-beta.32",
		},
		{
			Name:     "abbrev",
			Resolved: "https://registry.yarnpkg.com/abbrev/-/abbrev-1.1.1.tgz#f8f2c887ad10bf67f634f005b6987fed3179aac8",
			Specs:    []string{"1"},
			Version:  "1.1.1",
		},
		{
			Name:     "accepts",
			Resolved: "https://registry.yarnpkg.com/accepts/-/accepts-1.3.4.tgz#86246758c7dd6d21a6474ff084a4740ec05eb21f",
			Specs:    []string{"~1.3.4"},
			Dependencies: []Dependency{
				{"mime-types", "~2.1.16"},
				{"negotiator", "0.6.1"},
			},
			Version: "1.3.4",
		},
		{
			Name:     "acorn",
			Resolved: "https://registry.yarnpkg.com/acorn/-/acorn-4.0.13.tgz#105495ae5361d697bd195c825192e1ad7f253787",
			Specs:    []string{"^4.0.3", "^4.0.4"},
			Version:  "4.0.13",
		},
		{
			Name:     "address",
			Resolved: "https://registry.yarnpkg.com/address/-/address-1.0.3.tgz#b5f50631f8d6cec8bd20c963963afb55e06cbce9",
			Specs:    []string{"1.0.3", "^1.0.1"},
			Version:  "1.0.3",
		},
		{
			Name:     "chokidar",
			Resolved: "https://registry.yarnpkg.com/chokidar/-/chokidar-1.7.0.tgz#798e689778151c8076b4b360e5edd28cda2bb468",
			Specs:    []string{"^1.6.0", "^1.7.0"},
			Dependencies: []Dependency{
				{"anymatch", "^1.3.0"},
				{"async-each", "^1.0.0"},
				{"glob-parent", "^2.0.0"},
				{"inherits", "^2.0.1"},
				{"is-binary-path", "^1.0.0"},
				{"is-glob", "^2.0.0"},
				{"path-is-absolute", "^1.0.0"},
				{"readdirp", "^2.0.0"},
			},
			OptionalDependencies: []Dependency{
				{"fsevents", "^1.0.0"},
			},
			Version: "1.7.0",
		},
		{
			Name:     "cssom",
			Resolved: "https://registry.yarnpkg.com/cssom/-/cssom-0.3.2.tgz#b8036170c79f07a90ff2f16e22284027a243848b",
			Specs:    []string{"0.3.x", ">= 0.3.2 < 0.4.0"},
			Version:  "0.3.2",
		},
		{
			Name:     "semver",
			Resolved: "https://registry.yarnpkg.com/semver/-/semver-5.4.1.tgz#e059c09d8571f0540823733433505d3a2f00b18e",
			Specs:    []string{"2 || 3 || 4 || 5", "^5.0.3", "^5.1.0", "^5.3.0"},
			Version:  "5.4.1",
		},
	}

	actualEntries, actualError := Parse("testdata/yarn.lock")

	require.Nil(t, actualError)
	assert.Equal(t, expectedEntries, actualEntries)
}

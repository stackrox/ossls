package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectsFromNpmLockfileV3(t *testing.T) {
	actualProjects, err := ProjectsFromNpmLockfileV3("testdata/package-lock.json")

	expectedProjects := make(map[string]NpmProject, 4)
	expectedProjects["@aashutoshrathi/word-wrap"] = NpmProject{
		name:     "@aashutoshrathi/word-wrap",
		optional: false,
		version:  "1.2.6",
	}
	expectedProjects["@adobe/css-tools"] = NpmProject{
		name:     "@adobe/css-tools",
		optional: false,
		version:  "4.3.2",
	}
	expectedProjects["yup/node_modules/type-fest"] = NpmProject{
		name:     "yup/node_modules/type-fest",
		optional: false,
		version:  "2.19.0",
	}
	expectedProjects["@apollo/client"] = NpmProject{
		name:     "@apollo/client",
		optional: false,
		version:  "3.8.7",
	}

	require.Nil(t, err)

	for _, actualProject := range actualProjects {
		expectedProject, ok := expectedProjects[actualProject.Name()]
		require.True(t, ok)
		assert.Equal(t, expectedProject, actualProject)
	}
}

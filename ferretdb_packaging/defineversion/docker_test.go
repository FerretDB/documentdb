package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageURL(t *testing.T) {
	// expected URLs should work
	assert.Equal(
		t,
		"https://ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
		dockerImageURL("ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://quay.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
		dockerImageURL("quay.io/ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://hub.docker.com/r/ferretdb/postgres-documentdb-dev/tags",
		dockerImageURL("ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
}

func TestDockerSummary(t *testing.T) {
	dir := t.TempDir()

	summaryF, err := os.CreateTemp(dir, "summary")
	require.NoError(t, err)
	defer summaryF.Close()

	outputF, err := os.CreateTemp(dir, "output")
	require.NoError(t, err)
	defer outputF.Close()

	var stdout bytes.Buffer
	getenv := getEnvFunc(t, map[string]string{
		"GITHUB_STEP_SUMMARY": summaryF.Name(),
		"GITHUB_OUTPUT":       outputF.Name(),
	})
	action := githubactions.New(githubactions.WithGetenv(getenv), githubactions.WithWriter(&stdout))

	result := &images{
		developmentImages: []string{
			"ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb",
			"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
		},
		productionImages: []string{
			"quay.io/ferretdb/postgres-documentdb:latest",
		},
	}

	dockerSummary(action, result)

	expectedStdout := strings.ReplaceAll(`
 |Type        |Image                                                                                                                                  |
 |----        |-----                                                                                                                                  |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb'](https://ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:latest'](https://ghcr.io/ferretdb/postgres-documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/postgres-documentdb:latest'](https://quay.io/ferretdb/postgres-documentdb:latest)                                   |

`[1:], "'", "`",
	)
	assert.Equal(t, expectedStdout, stdout.String(), "stdout does not match")

	expectedSummary := strings.ReplaceAll(`
 |Type        |Image                                                                                                                                  |
 |----        |-----                                                                                                                                  |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb'](https://ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:latest'](https://ghcr.io/ferretdb/postgres-documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/postgres-documentdb:latest'](https://quay.io/ferretdb/postgres-documentdb:latest)                                   |

`[1:], "'", "`",
	)
	b, err := io.ReadAll(summaryF)
	require.NoError(t, err)
	assert.Equal(t, expectedSummary, string(b), "summary does not match")
}

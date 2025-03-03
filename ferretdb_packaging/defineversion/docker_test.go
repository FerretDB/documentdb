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

func TestDefineDockerTags(t *testing.T) {
	for name, tc := range map[string]struct {
		env            map[string]string
		pgVersion      string
		expectedDocker *images
	}{
		"pull_request": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
				},
			},
		},
		"pull_request-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:pr-docker-tag",
				},
			},
		},

		"pull_request_target": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "docker-tag",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
				},
			},
		},
		"pull_request_target-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "docker-tag",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:pr-docker-tag",
				},
			},
		},

		"push/ferretdb": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:ferretdb",
					//"quay.io/ferretdb/postgres-documentdb-dev:ferretdb",
				},
			},
		},
		"push/ferretdb-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:ferretdb",
				},
			},
		},

		"push/main": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
		},
		"push/main-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
		},

		"push/tag/release": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb",
					//"ferretdb/postgres-documentdb-dev:latest",
					"ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb-2.0.0",
					"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
					//"quay.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb",
					//"quay.io/ferretdb/postgres-documentdb-dev:latest",
				},
				productionImages: []string{
					//"ferretdb/postgres-documentdb:16-0.102.0-ferretdb",
					//"ferretdb/postgres-documentdb:latest",
					"ghcr.io/ferretdb/postgres-documentdb:16-0.102.0-ferretdb-2.0.0",
					"ghcr.io/ferretdb/postgres-documentdb:latest",
					//"quay.io/ferretdb/postgres-documentdb:16-0.102.0-ferretdb",
					//"quay.io/ferretdb/postgres-documentdb:latest",
				},
			},
		},
		"push/tag/release-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:16-0.102.0-ferretdb-2.0.0",
					"ghcr.io/otherorg/postgres-otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo:16-0.102.0-ferretdb-2.0.0",
					"ghcr.io/otherorg/postgres-otherrepo:latest",
				},
			},
		},

		"push/tag/release-rc": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					//"ferretdb/postgres-documentdb-dev:latest",
					"ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
					//"quay.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					//"quay.io/ferretdb/postgres-documentdb-dev:latest",
				},
				productionImages: []string{
					//"ferretdb/postgres-documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					//"ferretdb/postgres-documentdb:latest",
					"ghcr.io/ferretdb/postgres-documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/postgres-documentdb:latest",
					//"quay.io/ferretdb/postgres-documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					//"quay.io/ferretdb/postgres-documentdb:latest",
				},
			},
		},
		"push/tag/release-rc-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/postgres-otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/postgres-otherrepo:latest",
				},
			},
		},

		"push/tag/wrong": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "0.102.0-ferretdb-2.0.0-rc2", // no leading v
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
		},
		"push/tag/wrong-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "0.102.0-ferretdb-2.0.0-rc2", // no leading v
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
		},

		"schedule": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:ferretdb",
					//"quay.io/ferretdb/postgres-documentdb-dev:ferretdb",
				},
			},
		},
		"schedule-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:ferretdb",
				},
			},
		},

		"workflow_run": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:ferretdb",
					//"quay.io/ferretdb/postgres-documentdb-dev:ferretdb",
				},
			},
		},
		"workflow_run-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			pgVersion: "16",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:ferretdb",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := defineDockerVersion("0.100.0", tc.pgVersion, getEnvFunc(t, tc.env))
			if tc.expectedDocker == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedDocker, actual)
		})
	}
}

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

func TestDockerTagsResults(t *testing.T) {
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

	setDockerTagsResults(action, result)

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

	expectedOutput := `
development_images<<_GitHubActionsFileCommandDelimeter_
ghcr.io/ferretdb/postgres-documentdb-dev:16-0.102.0-ferretdb,ghcr.io/ferretdb/postgres-documentdb-dev:latest
_GitHubActionsFileCommandDelimeter_
production_images<<_GitHubActionsFileCommandDelimeter_
quay.io/ferretdb/postgres-documentdb:latest
_GitHubActionsFileCommandDelimeter_
`[1:]
	b, err = io.ReadAll(outputF)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b), "output parameters does not match")
}

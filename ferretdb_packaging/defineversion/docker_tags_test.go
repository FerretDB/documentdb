// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDefineDockerTags(t *testing.T) {
	for name, tc := range map[string]struct {
		env      map[string]string
		expected *images
	}{
		"pull_request": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/documentdb-dev:pr-docker-tag",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:pr-docker-tag",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/documentdb-dev:pr-docker-tag",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:pr-docker-tag",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/documentdb-dev:ferretdb",
					"quay.io/ferretdb/documentdb-dev:ferretdb",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:ferretdb",
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
				"INPUT_PG_VERSION":  "16",
			},
		},
		"push/main-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16",
			},
		},

		"push/tag/release": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:16-0.102.0-ferretdb",
					"ferretdb/documentdb-dev:latest",
					"ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb",
					"ghcr.io/ferretdb/documentdb-dev:latest",
					"quay.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb",
					"quay.io/ferretdb/documentdb-dev:latest",
				},
				productionImages: []string{
					"ferretdb/documentdb:16-0.102.0-ferretdb",
					"ferretdb/documentdb:latest",
					"ghcr.io/ferretdb/documentdb:16-0.102.0-ferretdb",
					"ghcr.io/ferretdb/documentdb:latest",
					"quay.io/ferretdb/documentdb:16-0.102.0-ferretdb",
					"quay.io/ferretdb/documentdb:latest",
				},
			},
		},
		"push/tag/release-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:16-0.102.0-ferretdb",
					"ghcr.io/otherorg/otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:16-0.102.0-ferretdb",
					"ghcr.io/otherorg/otherrepo:latest",
				},
			},
		},

		"push/tag/release-rc-major-minor": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16.7", // set major and minor version
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb-dev:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb-dev:latest",
					"ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb-dev:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb-dev:latest",
					"quay.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb-dev:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb-dev:latest",
				},
				productionImages: []string{
					"ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb:latest",
					"ghcr.io/ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb:latest",
					"quay.io/ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb:latest",
				},
			},
		},
		"push/tag/release-rc-major-minor-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
				"INPUT_PG_VERSION":  "16.7", // set major and minor version
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo-dev:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo:16.7-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo:latest",
				},
			},
		},

		"push/tag/release-major": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16", // set major version only
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb-dev:latest",
					"ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb-dev:latest",
					"quay.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb-dev:latest",
				},
				productionImages: []string{
					"ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"ferretdb/documentdb:latest",
					"ghcr.io/ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/ferretdb/documentdb:latest",
					"quay.io/ferretdb/documentdb:16-0.102.0-ferretdb-2.0.0-rc2",
					"quay.io/ferretdb/documentdb:latest",
				},
			},
		},
		"push/tag/release-major-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.102.0-ferretdb-2.0.0-rc2",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
				"INPUT_PG_VERSION":  "16", // set major version only
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:16-0.102.0-ferretdb-2.0.0-rc2",
					"ghcr.io/otherorg/otherrepo:latest",
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
				"INPUT_PG_VERSION":  "16",
			},
		},
		"push/tag/wrong-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "0.102.0-ferretdb-2.0.0-rc2", // no leading v
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
				"INPUT_PG_VERSION":  "16",
			},
		},

		"schedule": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/documentdb-dev:ferretdb",
					"quay.io/ferretdb/documentdb-dev:ferretdb",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:ferretdb",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ferretdb/documentdb-dev:ferretdb",
					"ghcr.io/ferretdb/documentdb-dev:ferretdb",
					"quay.io/ferretdb/documentdb-dev:ferretdb",
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
				"INPUT_PG_VERSION":  "16",
			},
			expected: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:ferretdb",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := defineDockerTags(getEnvFunc(t, tc.env))
			if tc.expected == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestImageURL(t *testing.T) {
	// expected URLs should work
	assert.Equal(
		t,
		"https://ghcr.io/ferretdb/documentdb-dev:pr-docker-tag",
		imageURL("ghcr.io/ferretdb/documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://quay.io/ferretdb/documentdb-dev:pr-docker-tag",
		imageURL("quay.io/ferretdb/documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://hub.docker.com/r/ferretdb/documentdb-dev/tags",
		imageURL("ferretdb/documentdb-dev:pr-docker-tag"),
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
			"ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb",
			"ghcr.io/ferretdb/documentdb-dev:latest",
		},
		productionImages: []string{
			"quay.io/ferretdb/documentdb:latest",
		},
	}

	setDockerTagsResults(action, result)

	expectedStdout := strings.ReplaceAll(`
 |Type        |Image                                                                                                                |
 |----        |-----                                                                                                                |
 |Development |['ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb'](https://ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/documentdb-dev:latest'](https://ghcr.io/ferretdb/documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/documentdb:latest'](https://quay.io/ferretdb/documentdb:latest)                                   |

`[1:], "'", "`",
	)
	assert.Equal(t, expectedStdout, stdout.String(), "stdout does not match")

	expectedSummary := strings.ReplaceAll(`
 |Type        |Image                                                                                                                |
 |----        |-----                                                                                                                |
 |Development |['ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb'](https://ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/documentdb-dev:latest'](https://ghcr.io/ferretdb/documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/documentdb:latest'](https://quay.io/ferretdb/documentdb:latest)                                   |

`[1:], "'", "`",
	)
	b, err := io.ReadAll(summaryF)
	require.NoError(t, err)
	assert.Equal(t, expectedSummary, string(b), "summary does not match")

	expectedOutput := `
development_images<<_GitHubActionsFileCommandDelimeter_
--tag=ghcr.io/ferretdb/documentdb-dev:16-0.102.0-ferretdb --tag=ghcr.io/ferretdb/documentdb-dev:latest
_GitHubActionsFileCommandDelimeter_
production_images<<_GitHubActionsFileCommandDelimeter_
--tag=quay.io/ferretdb/documentdb:latest
_GitHubActionsFileCommandDelimeter_
`[1:]
	b, err = io.ReadAll(outputF)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b), "output parameters does not match")
}

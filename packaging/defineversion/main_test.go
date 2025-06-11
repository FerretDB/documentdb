package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getEnvFunc implements [os.Getenv] for testing.
func getEnvFunc(t *testing.T, env map[string]string) func(string) string {
	t.Helper()

	return func(key string) string {
		val, ok := env[key]
		require.True(t, ok, "missing key %q", key)

		return val
	}
}

func TestParseGitTag(t *testing.T) {
	tests := map[string]struct {
		major      int
		minor      int
		patch      int
		prerelease string
		err        string
	}{
		"v0.103.0-ferretdb-2.2.0-beta.1": {
			major:      0,
			minor:      103,
			patch:      0,
			prerelease: "ferretdb-2.2.0-beta.1",
		},
		"0.103.0-ferretdb-2.2.0-beta.1": {
			err: `unexpected git tag format "0.103.0-ferretdb-2.2.0-beta.1"`,
		},
		"v0.103.0-ferretdb": {
			err: `prerelease "ferretdb" should start with 'ferretdb-'`,
		},
		"v0.103.0": {
			err: `prerelease "" should start with 'ferretdb-'`,
		},
	}

	for tag, tc := range tests {
		t.Run(tag, func(t *testing.T) {
			major, minor, patch, prerelease, err := parseGitTag(tag)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.major, major)
			assert.Equal(t, tc.minor, minor)
			assert.Equal(t, tc.patch, patch)
			assert.Equal(t, tc.prerelease, prerelease)
		})
	}
}

func TestDefineVersion(t *testing.T) {
	const pgVersion = "17"

	for name, tc := range map[string]struct {
		env                   map[string]string
		controlDefaultVersion string
		expected              *versions
		expectedErr           error
	}{
		"pull_request": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-pr-define-version",
				},
				debian: "0.103.0~pr~define~version",
			},
		},
		"pull_request-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-pr-define-version",
				},
				debian: "0.103.0~pr~define~version",
			},
		},

		"pull_request_target": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-pr-define-version",
				},
				debian: "0.103.0~pr~define~version",
			},
		},
		"pull_request_target-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "ferretdb",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-pr-define-version",
				},
				debian: "0.103.0~pr~define~version",
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
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
			controlDefaultVersion: "0.103.0",
			expectedErr:           fmt.Errorf(`unhandled branch "main"`),
		},
		"push/main-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			controlDefaultVersion: "0.103.0",
			expectedErr:           fmt.Errorf(`unhandled branch "main"`),
		},

		"push/tag/release": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.103.0-ferretdb-2.2.0-beta.1",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ferretdb/postgres-documentdb-dev:17",
					"ferretdb/postgres-documentdb-dev:17-0.103.0",
					"ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ferretdb/postgres-documentdb-dev:latest",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
					"quay.io/ferretdb/postgres-documentdb-dev:17",
					"quay.io/ferretdb/postgres-documentdb-dev:17-0.103.0",
					"quay.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb-2.2.0-beta.1",
					"quay.io/ferretdb/postgres-documentdb-dev:latest",
				},
				dockerProductionImages: []string{
					"ferretdb/postgres-documentdb:17",
					"ferretdb/postgres-documentdb:17-0.103.0",
					"ferretdb/postgres-documentdb:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ferretdb/postgres-documentdb:latest",
					"ghcr.io/ferretdb/postgres-documentdb:17",
					"ghcr.io/ferretdb/postgres-documentdb:17-0.103.0",
					"ghcr.io/ferretdb/postgres-documentdb:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ghcr.io/ferretdb/postgres-documentdb:latest",
					"quay.io/ferretdb/postgres-documentdb:17",
					"quay.io/ferretdb/postgres-documentdb:17-0.103.0",
					"quay.io/ferretdb/postgres-documentdb:17-0.103.0-ferretdb-2.2.0-beta.1",
					"quay.io/ferretdb/postgres-documentdb:latest",
				},
				debian: "0.103.0~ferretdb~2.2.0~beta.1",
			},
		},
		"push/tag/release-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.103.0-ferretdb-2.2.0-beta.1",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17",
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-0.103.0",
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ghcr.io/otherorg/postgres-otherrepo-dev:latest",
				},
				dockerProductionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo:17",
					"ghcr.io/otherorg/postgres-otherrepo:17-0.103.0",
					"ghcr.io/otherorg/postgres-otherrepo:17-0.103.0-ferretdb-2.2.0-beta.1",
					"ghcr.io/otherorg/postgres-otherrepo:latest",
				},
				debian: "0.103.0~ferretdb~2.2.0~beta.1",
			},
		},

		"push/tag/release-wrong-control-default-version": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.103.0-ferretdb-2.2.0-beta.1",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			controlDefaultVersion: "0.104.0",
			expectedErr:           fmt.Errorf(`git tag version "0.103.0" does not match the control file default version "0.104.0"`),
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
					"quay.io/ferretdb/postgres-documentdb-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
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
			controlDefaultVersion: "0.103.0",
			expected: &versions{
				dockerDevelopmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb",
				},
				dockerProductionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-ferretdb-prod",
				},
				debian: "0.103.0~ferretdb",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			docker, err := defineVersion(tc.controlDefaultVersion, pgVersion, getEnvFunc(t, tc.env))
			if tc.expected == nil {
				require.Error(t, tc.expectedErr)
				require.Equal(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, tc.expectedErr)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, docker)
		})
	}
}

func TestSummary(t *testing.T) {
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

	result := &versions{
		dockerDevelopmentImages: []string{
			"ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb",
			"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
		},
		dockerProductionImages: []string{
			"quay.io/ferretdb/postgres-documentdb:latest",
		},
		debian: "0.103.0~ferretdb",
	}

	setSummary(action, result)

	expectedStdout := strings.ReplaceAll(`
Debian package version ('upstream_version' only): '0.103.0~ferretdb'

 |Type        |Docker image                                                                                                                           |
 |----        |------------                                                                                                                           |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb'](https://ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:latest'](https://ghcr.io/ferretdb/postgres-documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/postgres-documentdb:latest'](https://quay.io/ferretdb/postgres-documentdb:latest)                                   |

`[1:], "'", "`",
	)
	assert.Equal(t, expectedStdout, stdout.String(), "stdout does not match")

	expectedSummary := strings.ReplaceAll(`
Debian package version ('upstream_version' only): '0.103.0~ferretdb'

 |Type        |Docker image                                                                                                                           |
 |----        |------------                                                                                                                           |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb'](https://ghcr.io/ferretdb/postgres-documentdb-dev:17-0.103.0-ferretdb) |
 |Development |['ghcr.io/ferretdb/postgres-documentdb-dev:latest'](https://ghcr.io/ferretdb/postgres-documentdb-dev:latest)                           |
 |Production  |['quay.io/ferretdb/postgres-documentdb:latest'](https://quay.io/ferretdb/postgres-documentdb:latest)                                   |

`[1:], "'", "`",
	)
	b, err := io.ReadAll(summaryF)
	require.NoError(t, err)
	assert.Equal(t, expectedSummary, string(b), "summary does not match")
}

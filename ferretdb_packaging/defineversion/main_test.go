package main

import (
	"testing"

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
		major      string
		minor      string
		patch      string
		prerelease string
		err        string
	}{
		"v1.100.0-ferretdb": {
			major:      "1",
			minor:      "100",
			patch:      "0",
			prerelease: "ferretdb",
		},
		"v1.100.0-ferretdb-2.0.1": {
			major:      "1",
			minor:      "100",
			patch:      "0",
			prerelease: "ferretdb-2.0.1",
		},
		"0.100.0-ferretdb": {
			err: `unexpected git tag format "0.100.0-ferretdb"`,
		},
		"v0.100.0": {
			err: `prerelease "" should include 'ferretdb-'`,
		},
	}

	for tag, tc := range tests {
		t.Run(tag, func(t *testing.T) {
			major, minor, patch, prerelease, err := parseGitTag(tag)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}

			assert.Equal(t, tc.major, major)
			assert.Equal(t, tc.minor, minor)
			assert.Equal(t, tc.patch, patch)
			assert.Equal(t, tc.prerelease, prerelease)
		})
	}
}

func TestDefineVersion(t *testing.T) {
	const (
		controlDefaultVersion = "0.100.0"
		pgVersion             = "17"
	)

	for name, tc := range map[string]struct {
		env            map[string]string
		expectedDebian string
		expectedDocker *images
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
			expectedDebian: "0.100.0~pr~define~version",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:pr-define-version",
				},
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
			expectedDebian: "0.100.0~pr~define~version",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:pr-define-version",
				},
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
			expectedDebian: "0.100.0~pr~define~version",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/ferretdb/postgres-documentdb-dev:pr-define-version",
				},
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
			expectedDebian: "0.100.0~pr~define~version",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:pr-define-version",
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
			expectedDebian: "0.100.0~branch~ferretdb",
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
			expectedDebian: "0.100.0~branch~ferretdb",
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
		},

		"push/tag/release": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0-ferretdb-2.0.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			expectedDebian: "0.100.0~ferretdb~2.0.0",
			expectedDocker: &images{
				developmentImages: []string{
					//"ferretdb/postgres-documentdb-dev:17-0.100.0-ferretdb",
					//"ferretdb/postgres-documentdb-dev:latest",
					"ghcr.io/ferretdb/postgres-documentdb-dev:17-0.100.0-ferretdb-2.0.0",
					"ghcr.io/ferretdb/postgres-documentdb-dev:latest",
					//"quay.io/ferretdb/postgres-documentdb-dev:17-0.100.0-ferretdb",
					//"quay.io/ferretdb/postgres-documentdb-dev:latest",
				},
				productionImages: []string{
					//"ferretdb/postgres-documentdb:17-0.100.0-ferretdb",
					//"ferretdb/postgres-documentdb:latest",
					"ghcr.io/ferretdb/postgres-documentdb:17-0.100.0-ferretdb-2.0.0",
					"ghcr.io/ferretdb/postgres-documentdb:latest",
					//"quay.io/ferretdb/postgres-documentdb:17-0.100.0-ferretdb",
					//"quay.io/ferretdb/postgres-documentdb:latest",
				},
			},
		},
		"push/tag/release-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0-ferretdb-2.0.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expectedDebian: "0.100.0~ferretdb~2.0.0",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:17-0.100.0-ferretdb-2.0.0",
					"ghcr.io/otherorg/postgres-otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo:17-0.100.0-ferretdb-2.0.0",
					"ghcr.io/otherorg/postgres-otherrepo:latest",
				},
			},
		},

		// FIXME
		//
		// "push/tag/missing-prerelease": {
		// 	env: map[string]string{
		// 		"GITHUB_EVENT_NAME": "push",
		// 		"GITHUB_HEAD_REF":   "",
		// 		"GITHUB_REF_NAME":   "v0.100.0", // missing prerelease
		// 		"GITHUB_REF_TYPE":   "tag",
		// 	},
		// },
		// "push/tag/not-ferretdb-prerelease": {
		// 	env: map[string]string{
		// 		"GITHUB_EVENT_NAME": "push",
		// 		"GITHUB_HEAD_REF":   "",
		// 		"GITHUB_REF_NAME":   "v0.100.0-other", // missing ferretdb in prerelease
		// 		"GITHUB_REF_TYPE":   "tag",
		// 	},
		// },
		// "push/tag/missing-v": {
		// 	env: map[string]string{
		// 		"GITHUB_EVENT_NAME": "push",
		// 		"GITHUB_HEAD_REF":   "",
		// 		"GITHUB_REF_NAME":   "0.100.0-ferretdb",
		// 		"GITHUB_REF_TYPE":   "tag",
		// 	},
		// },
		// "push/tag/not-semvar": {
		// 	env: map[string]string{
		// 		"GITHUB_EVENT_NAME": "push",
		// 		"GITHUB_HEAD_REF":   "",
		// 		"GITHUB_REF_NAME":   "v0.100-0-ferretdb",
		// 		"GITHUB_REF_TYPE":   "tag",
		// 	},
		// },

		"schedule": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/documentdb",
			},
			expectedDebian: "0.100.0~branch~ferretdb",
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
			expectedDebian: "0.100.0~branch~ferretdb",
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
			expectedDebian: "0.100.0~branch~ferretdb",
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
			expectedDebian: "0.100.0~branch~ferretdb",
			expectedDocker: &images{
				developmentImages: []string{
					"ghcr.io/otherorg/postgres-otherrepo-dev:ferretdb",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			debian, docker, err := defineVersion(controlDefaultVersion, pgVersion, getEnvFunc(t, tc.env))
			if tc.expectedDebian == "" && tc.expectedDocker == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedDebian, debian)
			assert.Equal(t, tc.expectedDocker, docker)
		})
	}
}

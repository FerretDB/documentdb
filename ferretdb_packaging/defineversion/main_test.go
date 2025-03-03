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

func TestDefine(t *testing.T) {
	for name, tc := range map[string]struct {
		env                   map[string]string
		controlDefaultVersion string
		expected              string
	}{
		"pull_request": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
			},
			controlDefaultVersion: "0.100.0",
			expected:              "0.100.0~pr~define~version",
		},

		"pull_request_target": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-version",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
			},
			controlDefaultVersion: "0.100.0",
			expected:              "0.100.0~pr~define~version",
		},

		"push/ferretdb": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
			},
			controlDefaultVersion: "0.100.0",
			expected:              "0.100.0~branch~ferretdb",
		},
		"push/other": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "releases",
				"GITHUB_REF_TYPE":   "other", // not ferretdb branch
			},
		},

		"push/tag/v0.100.0-ferretdb": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0-ferretdb",
				"GITHUB_REF_TYPE":   "tag",
			},
			expected: "0.100.0~ferretdb",
		},
		"push/tag/v0.100.0-ferretdb-2.0.1": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0-ferretdb-2.0.1",
				"GITHUB_REF_TYPE":   "tag",
			},
			expected: "0.100.0~ferretdb~2.0.1",
		},

		"push/tag/missing-prerelease": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0", // missing prerelease
				"GITHUB_REF_TYPE":   "tag",
			},
		},
		"push/tag/not-ferretdb-prerelease": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100.0-other", // missing ferretdb in prerelease
				"GITHUB_REF_TYPE":   "tag",
			},
		},
		"push/tag/missing-v": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "0.100.0-ferretdb",
				"GITHUB_REF_TYPE":   "tag",
			},
		},
		"push/tag/not-semvar": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100-0-ferretdb",
				"GITHUB_REF_TYPE":   "tag",
			},
		},

		"schedule": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
			},
			controlDefaultVersion: "0.100.0",
			expected:              "0.100.0~branch~ferretdb",
		},

		"workflow_run": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
			},
			controlDefaultVersion: "0.100.0",
			expected:              "0.100.0~branch~ferretdb",
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := defineDebianVersion(tc.controlDefaultVersion, getEnvFunc(t, tc.env))
			if tc.expected == "" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestSemVar(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tag        string
		major      string
		minor      string
		patch      string
		prerelease string
		err        string
	}{
		"Valid": {
			tag:        "v1.100.0-ferretdb",
			major:      "1",
			minor:      "100",
			patch:      "0",
			prerelease: "ferretdb",
		},
		"SpecificVersion": {
			tag:        "v1.100.0-ferretdb-2.0.1",
			major:      "1",
			minor:      "100",
			patch:      "0",
			prerelease: "ferretdb-2.0.1",
		},
		"MissingV": {
			tag: "0.100.0-ferretdb",
			err: `unexpected tag syntax "0.100.0-ferretdb"`,
		},
		"MissingFerretDB": {
			tag: "v0.100.0",
			err: "prerelease is empty",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			major, minor, patch, prerelease, err := semVar(tc.tag)

			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}

			require.Equal(t, tc.major, major)
			require.Equal(t, tc.minor, minor)
			require.Equal(t, tc.patch, patch)
			require.Equal(t, tc.prerelease, prerelease)
		})
	}
}

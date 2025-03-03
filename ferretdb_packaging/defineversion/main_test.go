package main

import (
	"testing"

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
			err: `unexpected git tag format "0.100.0-ferretdb"`,
		},
		"MissingFerretDB": {
			tag: "v0.100.0",
			err: `prerelease "" should include 'ferretdb-'`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			major, minor, patch, prerelease, err := parseGitTag(tc.tag)

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

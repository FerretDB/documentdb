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
	"io"
	"os"
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
func TestDefine(t *testing.T) {
	for name, tc := range map[string]struct {
		env            map[string]string
		defaultVersion string // pg_documentdb/documentdb.control file's default_version field
		expected       string
	}{
		"pull_request": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~pr~define~docker~tag",
		},

		"pull_request/dependabot": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
				"GITHUB_REF_NAME":   "58/merge",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~pr~mongo~go~driver~29d768e",
		},

		"pull_request_target": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~pr~define~docker~tag",
		},

		"push/main": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~main",
		},
		"push/ferretdb": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "ferretdb",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~ferretdb",
		},
		"push/other": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "releases",
				"GITHUB_REF_TYPE":   "other", // not main or ferretdb branch
			},
		},

		"push/tag/release1": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100-0",
				"GITHUB_REF_TYPE":   "tag",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0",
		},

		"push/tag/mismatch": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v0.100-1", // default version and tag mismatch
				"GITHUB_REF_TYPE":   "tag",
			},
			defaultVersion: "0.100-0",
		},

		"push/tag/wrong": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "0.100-0", // no leading v
				"GITHUB_REF_TYPE":   "tag",
			},
		},

		"schedule": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~main",
		},

		"workflow_run": {
			env: map[string]string{
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
			},
			defaultVersion: "0.100-0",
			expected:       "0.100.0~main",
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := define(tc.defaultVersion, getEnvFunc(t, tc.env))
			if tc.expected == "" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestResults(t *testing.T) {
	dir := t.TempDir()

	summaryF, err := os.CreateTemp(dir, "summary")
	require.NoError(t, err)
	defer summaryF.Close() //nolint:errcheck // temporary file for testing

	outputF, err := os.CreateTemp(dir, "output")
	require.NoError(t, err)
	defer outputF.Close() //nolint:errcheck // temporary file for testing

	var stdout bytes.Buffer
	getenv := getEnvFunc(t, map[string]string{
		"GITHUB_STEP_SUMMARY": summaryF.Name(),
		"GITHUB_OUTPUT":       outputF.Name(),
	})
	action := githubactions.New(githubactions.WithGetenv(getenv), githubactions.WithWriter(&stdout))

	version := "0.100.0~main"

	setResults(action, version)

	expected := "version: 0.100.0~main\n"
	assert.Equal(t, expected, stdout.String(), "stdout does not match")

	b, err := io.ReadAll(summaryF)
	require.NoError(t, err)
	assert.Equal(t, expected, string(b), "summary does not match")

	expectedOutput := `
version<<_GitHubActionsFileCommandDelimeter_
0.100.0~main
_GitHubActionsFileCommandDelimeter_
`[1:]
	b, err = io.ReadAll(outputF)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b), "output parameters does not match")
}

func TestControlVersion(t *testing.T) {
	dir := t.TempDir()

	controlF, err := os.CreateTemp(dir, "test.control")
	defer controlF.Close() //nolint:errcheck // temporary file for testing

	buf := `comment = 'API surface for DocumentDB for PostgreSQL'
default_version = '0.100-0'
module_pathname = '$libdir/pg_documentdb'
relocatable = false
superuser = true
requires = 'documentdb_core, pg_cron, tsm_system_rows, vector, postgis, rum'`
	_, err = io.WriteString(controlF, buf)
	require.NoError(t, err)

	version, err := controlVersion(controlF.Name())
	require.NoError(t, err)

	require.Equal(t, "0.100-0", version)
}

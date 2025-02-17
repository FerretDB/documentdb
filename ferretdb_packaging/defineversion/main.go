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

// Package main defines debian version number of DocumentDB for CI builds.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	controlFileF := flag.String("control-file", "../pg_documentdb/documentdb.control", "pg_documentdb/documentdb.control file path")
	flag.Parse()

	action := githubactions.New()

	if *controlFileF == "" {
		action.Fatalf("%s", "-control-file flag is empty.")
	}

	version, err := controlVersion(*controlFileF)

	debugEnv(action)

	res, err := define(version, action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	setResults(action, res)
}

// controlDefaultVer matches the default_version field in control file,
// see pg_documentdb_core/documentdb_core.control.
var controlDefaultVer = regexp.MustCompile(`default_version\s*=\s*'([0-9]+\.[0-9]+-[0-9]+)'`)

// documentDBVer is the version syntax used by documentdb.
// For documentdb.control file, the version is in the format of `0.100-0`.
// For release tags, it has a leading `v` such as `v0.100-0`.
//
//nolint:lll // for readibility
var documentDBVer = regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)-(?P<patch>0|[1-9]\d*)$`)

// debianVer matches allowed characters for debian package version,
// see https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
var debianVer = regexp.MustCompile(`[^A-Za-z0-9~.+]`)

// debugEnv logs all environment variables that start with `GITHUB_` or `INPUT_`
// in debug level.
func debugEnv(action *githubactions.Action) {
	res := make([]string, 0, 30)

	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") || strings.HasPrefix(l, "INPUT_") {
			res = append(res, l)
		}
	}

	slices.Sort(res)

	action.Debugf("Dumping environment variables:")

	for _, l := range res {
		action.Debugf("\t%s", l)
	}
}

// controlVersion returns the default_version field from the control file,
// see pg_documentdb_core/documentdb_core.control.
func controlVersion(f string) (string, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}

	match := controlDefaultVer.FindSubmatch(b)
	if len(match) != 2 {
		return "", fmt.Errorf("control file did not find default_version match file: %s", f)
	}

	version := string(match[1])

	return version, nil
}

// define extracts debian version number from the environment variables of GitHub Actions.
// If the release tag is set, it checks the tag matches the control version
// and returns an error on mismatch.
func define(controlDefaultVersion string, getenv githubactions.GetenvFunc) (string, error) {
	version, err := parseVersion(controlDefaultVersion)
	if err != nil {
		return "", err
	}

	var res string

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		res = defineForPR(version, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineForBranch(version, refName)

		case "tag":
			res, err = defineForTag(version, refName)

		default:
			err = fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		err = fmt.Errorf("unhandled event type %q", event)
	}

	if err != nil {
		return "", err
	}

	if res == "" {
		panic("both res and err are nil")
	}

	return res, nil
}

// defineForPR defines debian version number for pull requests.
// It replaces branch name with allowed chars of debian package version.
func defineForPR(version, branch string) string {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]
	branch = debianVer.ReplaceAllString(branch, "~")

	return fmt.Sprintf("%s~pr~%s", version, branch)
}

// defineForBranch defines debian version number for branch builds.
func defineForBranch(version, branch string) (string, error) {
	switch branch {
	case "main", "ferretdb":
		return fmt.Sprintf("%s~%s", version, branch), nil
	default:
		return "", fmt.Errorf("unhandled branch %q", branch)
	}
}

// defineForTag defines debian version number for release build.
// It returns an error if tag version does not match the control version.
func defineForTag(controlVersion string, tag string) (string, error) {
	tagVersion, err := parseVersion(tag)
	if err != nil {
		return "", err
	}

	if tagVersion != controlVersion {
		return "", fmt.Errorf("version in control file and release tag mismatch control:%s tag:%s", controlVersion, tagVersion)
	}

	return tagVersion, nil
}

// parseVersion parses the version string in the format `0.100-0` or `v0.100-0` and
// returns a normalized version string in `0.100.0` format.
func parseVersion(version string) (string, error) {
	match := documentDBVer.FindStringSubmatch(version)
	if match == nil || len(match) != documentDBVer.NumSubexp()+1 {
		return "", fmt.Errorf("unexpected version syntax %q", version)
	}

	major := match[documentDBVer.SubexpIndex("major")]
	minor := match[documentDBVer.SubexpIndex("minor")]
	patch := match[documentDBVer.SubexpIndex("patch")]

	return major + "." + minor + "." + patch, nil
}

// setResults sets action output parameters, summary, etc.
func setResults(action *githubactions.Action, res string) {
	output := fmt.Sprintf("version: %s", res)

	action.AddStepSummary(output)
	action.Infof("%s", output)
	action.SetOutput("version", res)
}

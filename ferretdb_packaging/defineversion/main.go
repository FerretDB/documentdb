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
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	versionF := flag.String("version", "", "pg_documentdb/documentdb.control file's default_version field")
	flag.Parse()

	action := githubactions.New()

	if *versionF == "" {
		action.Fatalf("%s", "-version flag is empty.")
	}

	debugEnv(action)

	res, err := define(*versionF, action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	setResults(action, res)
}

// documentDBVer is the version syntax used by documentdb.
// For documentdb.control file, the version is in the format of `0.100-0`.
// For a release tag it has a leading `v` such as `v0.100-0`.
//
//nolint:lll // for readibility
var documentDBVer = regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)-(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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

// Define extracts Docker image names and tags from the environment variables defined by GitHub Actions.
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

// defineForPR defines package version for pull requests.
func defineForPR(version, branch string) string {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	return fmt.Sprintf("%s~pre.pr-%s", version, branch)
}

// defineForBranch defines package version for branch builds.
func defineForBranch(version, branch string) (string, error) {
	switch branch {
	case "main", "ferretdb":
		return fmt.Sprintf("%s~pre.%s", version, branch), nil
	default:
		return "", fmt.Errorf("unhandled branch %q", branch)
	}
}

// defineForTag defines package version for prerelease tag builds.
func defineForTag(version string, tag string) (string, error) {
	tagVersion, err := parseVersion(tag)
	if err != nil {
		return "", err
	}

	if tagVersion != version {
		return "", fmt.Errorf("version in control file and release tag mismatch control:%s tag:%s", version, tagVersion)
	}

	return tagVersion, nil
}

// parseVersion parses the version string in the format `0.100-0` and
// returns a normalized version string in `0.100.0` format.
func parseVersion(version string) (string, error) {
	match := documentDBVer.FindStringSubmatch(version)
	if match == nil || len(match) != documentDBVer.NumSubexp()+1 {
		return "", fmt.Errorf("unexpected version syntax %q", version)
	}

	major := match[documentDBVer.SubexpIndex("major")]
	minor := match[documentDBVer.SubexpIndex("minor")]
	patch := match[documentDBVer.SubexpIndex("patch")]
	prerelease := match[documentDBVer.SubexpIndex("prerelease")]

	if prerelease == "" {
		return major + "." + minor + "." + patch, nil
	}

	return major + "." + minor + "." + patch + "~" + prerelease, nil
}

// setResults sets action output parameters, summary, etc.
func setResults(action *githubactions.Action, res string) {
	output := fmt.Sprintf("version: %s", res)

	action.AddStepSummary(output)
	action.Infof("%s", output)
	action.SetOutput("version", res)
}

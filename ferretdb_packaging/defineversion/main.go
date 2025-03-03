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
	commandF := flag.String("command", "", "command to run, possible values: [deb-version, docker-tags]")

	controlFileF := flag.String("control-file", "../pg_documentdb/documentdb.control", "pg_documentdb/documentdb.control file path")

	flag.Parse()

	action := githubactions.New()

	debugEnv(action)

	if *commandF == "" {
		action.Fatalf("-command flag is empty.")
	}

	switch *commandF {
	case "deb-version":
		if *controlFileF == "" {
			action.Fatalf("%s", "-control-file flag is empty.")
		}

		controlDefaultVersion, err := getControlDefaultVersion(*controlFileF)
		if err != nil {
			action.Fatalf("%s", err)
		}

		packageVersion, err := definePackageVersion(controlDefaultVersion, action.Getenv)
		if err != nil {
			action.Fatalf("%s", err)
		}

		setDebianVersionResults(action, packageVersion)
	case "docker-tags":
		res, err := defineDockerTags(action.Getenv)
		if err != nil {
			action.Fatalf("%s", err)
		}

		setDockerTagsResults(action, res)
	default:
		action.Fatalf("unhandled command %q", *commandF)
	}
}

// controlDefaultVer matches major, minor and "patch" from default_version field in control file,
// see pg_documentdb_core/documentdb_core.control.
var controlDefaultVer = regexp.MustCompile(`(?m)^default_version = '(?P<major>[0-9]+)\.(?P<minor>[0-9]+)-(?P<patch>[0-9]+)'$`)

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// disallowedVer matches disallowed characters of Debian `upstream_version` when used without `debian_revision`.
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
var disallowedVer = regexp.MustCompile(`[^A-Za-z0-9~.+]`)

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

// getControlDefaultVersion returns the default_version field from the control file
// in SemVer format (0.100-0 -> 0.100.0).
func getControlDefaultVersion(f string) (string, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}

	match := controlDefaultVer.FindSubmatch(b)
	if match == nil || len(match) != controlDefaultVer.NumSubexp()+1 {
		return "", fmt.Errorf("control file did not find default_version: %s", f)
	}

	major := match[controlDefaultVer.SubexpIndex("major")]
	minor := match[controlDefaultVer.SubexpIndex("minor")]
	patch := match[controlDefaultVer.SubexpIndex("patch")]

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}

// definePackageVersion returns valid Debian package version,
// based on `default_version` in the control file and environment variables set by GitHub Actions.
//
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
// We use `upstream_version` only.
// For that reason, we can't use `-`, so we replace it with `~`.
func definePackageVersion(controlDefaultVersion string, getenv githubactions.GetenvFunc) (string, error) {
	var packageVersion string
	var err error

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		packageVersion = definePackageVersionForPR(controlDefaultVersion, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			packageVersion, err = definePackageVersionForBranch(controlDefaultVersion, refName)

		case "tag":
			packageVersion, err = definePackagerVersionForTag(refName)

		default:
			err = fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		err = fmt.Errorf("unhandled event type %q", event)
	}

	if err != nil {
		return "", err
	}

	if packageVersion == "" {
		return "", fmt.Errorf("both packageVersion and err are nil")
	}

	return packageVersion, nil
}

// definePackageVersionForPR returns valid Debian package version for PR.
// See [definePackageVersion].
func definePackageVersionForPR(controlDefaultVersion, branch string) string {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]
	res := fmt.Sprintf("%s-pr-%s", controlDefaultVersion, branch)

	return disallowedVer.ReplaceAllString(res, "~")
}

// definePackageVersionForBranch returns valid Debian package version for branch.
// See [definePackageVersion].
func definePackageVersionForBranch(controlDefaultVersion, branch string) (string, error) {
	switch branch {
	case "ferretdb":
		return fmt.Sprintf("%s~branch~%s", controlDefaultVersion, branch), nil
	default:
		return "", fmt.Errorf("unhandled branch %q", branch)
	}
}

// semVar parses tag and returns version components.
//
// It returns error for invalid tag syntax, prerelease is missing `ferretdb` or if it has buildmetadata.
func semVar(tag string) (major, minor, patch, prerelease string, err error) {
	match := semVerTag.FindStringSubmatch(tag)
	if match == nil || len(match) != semVerTag.NumSubexp()+1 {
		return "", "", "", "", fmt.Errorf("unexpected tag syntax %q", tag)
	}

	major = match[semVerTag.SubexpIndex("major")]
	minor = match[semVerTag.SubexpIndex("minor")]
	patch = match[semVerTag.SubexpIndex("patch")]
	prerelease = match[semVerTag.SubexpIndex("prerelease")]
	buildmetadata := match[semVerTag.SubexpIndex("buildmetadata")]

	if prerelease == "" {
		return "", "", "", "", fmt.Errorf("prerelease is empty")
	}

	if !strings.Contains(prerelease, "ferretdb") {
		return "", "", "", "", fmt.Errorf("prerelease %q should include `ferretdb`", prerelease)
	}

	if buildmetadata != "" {
		return "", "", "", "", fmt.Errorf("buildmetadata %q is present", buildmetadata)
	}

	return
}

// definePackagerVersionForTag returns valid Debian package version for tag.
// See [definePackageVersion].
func definePackagerVersionForTag(tag string) (string, error) {
	major, minor, patch, prerelease, err := semVar(tag)
	if err != nil {
		return "", err
	}

	res := fmt.Sprintf("%s.%s.%s-%s", major, minor, patch, prerelease)
	return disallowedVer.ReplaceAllString(res, "~"), nil
}

// setDebianVersionResults sets action output parameters, summary, etc.
func setDebianVersionResults(action *githubactions.Action, res string) {
	output := fmt.Sprintf("version: `%s`", res)

	action.AddStepSummary(output)
	action.Infof("%s", output)
	action.SetOutput("version", res)
}

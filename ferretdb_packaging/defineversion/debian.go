package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// controlDefaultVer matches major, minor and "patch" from `default_version` field in control file,
// see pg_documentdb_core/documentdb_core.control.
var controlDefaultVer = regexp.MustCompile(`(?m)^default_version = '(?P<major>[0-9]+)\.(?P<minor>[0-9]+)-(?P<patch>[0-9]+)'$`)

// disallowedDebian matches disallowed characters of Debian `upstream_version` when used without `debian_revision`.
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
var disallowedDebian = regexp.MustCompile(`[^A-Za-z0-9\.+~]`)

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

// defineDebianVersion returns valid Debian package version,
// based on `default_version` in the control file and environment variables set by GitHub Actions.
//
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
// We use `upstream_version` only.
// For that reason, we can't use `-`, so we replace it with `~`.
func defineDebianVersion(controlDefaultVersion, pgVersion string, getenv githubactions.GetenvFunc) (string, error) {
	var res string
	var err error

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		res = defineDebianVersionForPR(controlDefaultVersion, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineDebianVersionForBranch(controlDefaultVersion, refName)

		case "tag":
			res, err = defineDebianVersionForTag(refName)

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
		return "", fmt.Errorf("both packageVersion and err are nil")
	}

	return res, nil
}

// defineDebianVersionForPR returns valid Debian package version for PR.
// See [defineDebianVersion].
func defineDebianVersionForPR(controlDefaultVersion, branch string) string {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]
	res := fmt.Sprintf("%s-pr-%s", controlDefaultVersion, branch)

	return disallowedDebian.ReplaceAllString(res, "~")
}

// defineDebianVersionForBranch returns valid Debian package version for branch.
// See [defineDebianVersion].
func defineDebianVersionForBranch(controlDefaultVersion, branch string) (string, error) {
	if branch != "ferretdb" {
		return "", fmt.Errorf("unhandled branch %q", branch)
	}

	return fmt.Sprintf("%s~ferretdb", controlDefaultVersion), nil
}

// defineDebianVersionForTag returns valid Debian package version for tag.
// See [defineDebianVersion].
func defineDebianVersionForTag(tag string) (string, error) {
	major, minor, patch, prerelease, err := parseGitTag(tag)
	if err != nil {
		return "", err
	}

	res := fmt.Sprintf("%d.%d.%d-%s", major, minor, patch, prerelease)
	return disallowedDebian.ReplaceAllString(res, "~"), nil
}

// debianSummary sets action summary.
func debianSummary(action *githubactions.Action, version string) {
	output := fmt.Sprintf("Debian package version (`upstream_version` only): `%s`", version)

	// Only 3 summaries are shown in the GitHub Actions UI by default,
	// and Docker summaries are more important (and include Debian version anyway).
	// action.AddStepSummary(output)

	action.Infof("%s", output)
}

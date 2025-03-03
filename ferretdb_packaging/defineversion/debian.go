package main

import (
	"fmt"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

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

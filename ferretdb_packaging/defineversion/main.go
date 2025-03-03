// Package main defines version numbers for DocumentDB.
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

	debugEnv(action)

	if *controlFileF == "" {
		action.Fatalf("%s", "-control-file flag is empty.")
	}

	controlDefaultVersion, err := getControlDefaultVersion(*controlFileF)
	if err != nil {
		action.Fatalf("%s", err)
	}

	packageVersion, err := defineDebianVersion(controlDefaultVersion, action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	output := fmt.Sprintf("Debian package version (`upstream_version` only): `%s`", packageVersion)

	action.AddStepSummary(output)
	action.Infof("%s", output)
	action.SetOutput("version", packageVersion)

	dockerTags, err := defineDockerVersion(action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	setDockerTagsResults(action, dockerTags)
}

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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

// semVar parses tag and returns version components.
//
// It returns error for invalid tag syntax, prerelease is missing `ferretdb` or if it has buildmetadata.
//
// FIXME
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

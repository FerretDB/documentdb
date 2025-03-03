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
	pgVersionF := flag.String("pg-version", "17", "Major PostgreSQL version")

	flag.Parse()

	action := githubactions.New()

	debugEnv(action)

	if *controlFileF == "" {
		action.Fatalf("%s", "-control-file flag is empty.")
	}

	switch *pgVersionF {
	case "15", "16", "17":
		// nothing
	default:
		action.Fatalf("%s", fmt.Sprintf("Invalid PostgreSQL version %q.", *pgVersionF))
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

	dockerTags, err := defineDockerVersion(*pgVersionF, action.Getenv)
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

// parseGitTag parses git tag in specific format and returns SemVer components.
//
// Expected format is v0.100.0-ferretdb-2.0.0-rc.2,
// where v0.100.0 is a DocumentDB version (0.100-0 -> 0.100.0),
// and ferretdb-2.0.0-rc.2 is a compatible FerretDB version.
func parseGitTag(tag string) (major, minor, patch, prerelease string, err error) {
	match := semVerTag.FindStringSubmatch(tag)
	if match == nil || len(match) != semVerTag.NumSubexp()+1 {
		err = fmt.Errorf("unexpected git tag format %q", tag)
		return
	}

	major = match[semVerTag.SubexpIndex("major")]
	minor = match[semVerTag.SubexpIndex("minor")]
	patch = match[semVerTag.SubexpIndex("patch")]
	prerelease = match[semVerTag.SubexpIndex("prerelease")]
	buildmetadata := match[semVerTag.SubexpIndex("buildmetadata")]

	if !strings.Contains(prerelease, "ferretdb-") {
		err = fmt.Errorf("prerelease %q should include 'ferretdb-'", prerelease)
		return
	}

	if buildmetadata != "" {
		err = fmt.Errorf("buildmetadata %q is present", buildmetadata)
		return
	}

	return
}

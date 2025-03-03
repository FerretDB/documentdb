// Package main defines version numbers for DocumentDB.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// parseGitTag parses git tag in specific format and returns SemVer components.
//
// Expected format is v0.100.0-ferretdb-2.0.0-rc.2,
// where v0.100.0 is a DocumentDB version (0.100-0 -> 0.100.0),
// and ferretdb-2.0.0-rc.2 is a compatible FerretDB version.
func parseGitTag(tag string) (major, minor, patch int, prerelease string, err error) {
	match := semVerTag.FindStringSubmatch(tag)
	if match == nil || len(match) != semVerTag.NumSubexp()+1 {
		err = fmt.Errorf("unexpected git tag format %q", tag)
		return
	}

	if major, err = strconv.Atoi(match[semVerTag.SubexpIndex("major")]); err != nil {
		return
	}
	if minor, err = strconv.Atoi(match[semVerTag.SubexpIndex("minor")]); err != nil {
		return
	}
	if patch, err = strconv.Atoi(match[semVerTag.SubexpIndex("patch")]); err != nil {
		return
	}
	prerelease = match[semVerTag.SubexpIndex("prerelease")]
	buildmetadata := match[semVerTag.SubexpIndex("buildmetadata")]

	if !strings.HasPrefix(prerelease, "ferretdb-") {
		err = fmt.Errorf("prerelease %q should start with 'ferretdb-'", prerelease)
		return
	}

	if buildmetadata != "" {
		err = fmt.Errorf("buildmetadata %q is present", buildmetadata)
		return
	}

	return
}

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

// defineVersion returns Debian package version and Docker tags.
func defineVersion(controlDefaultVersion, pgVersion string, getenv githubactions.GetenvFunc) (string, *images, error) {
	debian, err := defineDebianVersion(controlDefaultVersion, pgVersion, getenv)
	if err != nil {
		return "", nil, err
	}

	docker, err := defineDockerVersion(controlDefaultVersion, pgVersion, getenv)
	if err != nil {
		return "", nil, err
	}

	return debian, docker, nil
}

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

	debian, docker, err := defineVersion(controlDefaultVersion, *pgVersionF, action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	debianSummary(action, debian)
	action.SetOutput("debian_version", debian)

	dockerSummary(action, docker)
	action.SetOutput("docker_development_images", strings.Join(docker.developmentImages, ","))
	action.SetOutput("docker_production_images", strings.Join(docker.productionImages, ","))
}

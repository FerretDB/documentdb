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
	"text/tabwriter"

	"github.com/sethvargo/go-githubactions"
)

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// versions represents Docker image names and tags, and Debian package version.
type versions struct {
	dockerDevelopmentImages []string
	dockerProductionImages  []string
	debian                  string
}

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

// defineVersion extracts Docker image names and tags, and Debian package version using the environment variables defined by GitHub Actions.
//
// The Debian package version is based on `default_version` in the control file.
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
// We use `upstream_version` only.
// For that reason, we can't use `-`, so we replace it with `~`.
func defineVersion(controlDefaultVersion, pgVersion string, getenv githubactions.GetenvFunc) (*versions, error) {
	repo := getenv("GITHUB_REPOSITORY")

	// to support GitHub forks
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to split %q into owner and name", repo)
	}
	owner := parts[0]
	repo = parts[1]

	var res *versions
	var err error

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		res = defineVersionForPR(controlDefaultVersion, pgVersion, owner, repo, branch)

	case "push", "schedule", "workflow_dispatch":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineVersionForBranch(controlDefaultVersion, pgVersion, owner, repo, refName)

		case "tag":
			res, err = defineVersionForTag(controlDefaultVersion, pgVersion, owner, repo, refName)

		default:
			err = fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		err = fmt.Errorf("unhandled event type %q", event)
	}

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, fmt.Errorf("both res and err are nil")
	}

	slices.Sort(res.dockerDevelopmentImages)
	slices.Sort(res.dockerProductionImages)

	return res, nil
}

// defineVersionForPR defines Docker image names and tags, and Debian package version for PR.
// See [defineVersion].
func defineVersionForPR(controlDefaultVersion, pgVersion, owner, repo, branch string) *versions {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	res := &versions{
		dockerDevelopmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s-pr-%s", owner, repo, pgVersion, branch),
		},
		debian: disallowedDebian.ReplaceAllString(fmt.Sprintf("%s-pr-%s", controlDefaultVersion, branch), "~"),
	}

	// PRs are only for testing; no Quay.io and Docker Hub repos

	return res
}

// defineVersionForBranch defines Docker image names and tags, and Debian package version for branch.
// See [defineVersion].
func defineVersionForBranch(controlDefaultVersion, pgVersion, owner, repo, branch string) (*versions, error) {
	if branch != "ferretdb" {
		return nil, fmt.Errorf("unhandled branch %q", branch)
	}

	res := &versions{
		dockerDevelopmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s-ferretdb", owner, repo, pgVersion),
		},
		debian: fmt.Sprintf("%s~ferretdb", controlDefaultVersion),
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res, nil
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "documentdb" {
		return res, nil
	}

	res.dockerDevelopmentImages = append(res.dockerDevelopmentImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb-dev:%s-ferretdb", pgVersion))
	res.dockerDevelopmentImages = append(res.dockerDevelopmentImages, fmt.Sprintf("ferretdb/postgres-documentdb-dev:%s-ferretdb", pgVersion))

	return res, nil
}

// defineVersionForTag defines Docker image names and tags, and Debian package version for tag.
// See [defineVersion].
func defineVersionForTag(controlDefaultVersion, pgVersion, owner, repo, tag string) (*versions, error) {
	major, minor, patch, prerelease, err := parseGitTag(tag)
	if err != nil {
		return nil, err
	}

	tagVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if tagVersion != controlDefaultVersion {
		return nil, fmt.Errorf("git tag version %q does not match the control file default version %q", tagVersion, controlDefaultVersion)
	}

	tags := []string{
		fmt.Sprintf("%s-%s-%s", pgVersion, tagVersion, prerelease),
		fmt.Sprintf("%s-%s", pgVersion, tagVersion),
		fmt.Sprintf("%s", pgVersion),
	}

	if pgVersion == "17" {
		tags = append(tags, "latest")
	}

	res := versions{
		debian: disallowedDebian.ReplaceAllString(fmt.Sprintf("%s-%s", tagVersion, prerelease), "~"),
	}

	for _, t := range tags {
		res.dockerDevelopmentImages = append(res.dockerDevelopmentImages, fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s", owner, repo, t))
		res.dockerProductionImages = append(res.dockerProductionImages, fmt.Sprintf("ghcr.io/%s/postgres-%s:%s", owner, repo, t))
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return &res, nil
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "documentdb" {
		return &res, nil
	}

	for _, t := range tags {
		res.dockerDevelopmentImages = append(res.dockerDevelopmentImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb-dev:%s", t))
		res.dockerProductionImages = append(res.dockerProductionImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb:%s", t))

		res.dockerDevelopmentImages = append(res.dockerDevelopmentImages, fmt.Sprintf("ferretdb/postgres-documentdb-dev:%s", t))
		res.dockerProductionImages = append(res.dockerProductionImages, fmt.Sprintf("ferretdb/postgres-documentdb:%s", t))
	}

	return &res, nil
}

// setSummary sets action summary.
func setSummary(action *githubactions.Action, version *versions) {
	var buf strings.Builder

	fmt.Fprintf(&buf, "Debian package version (`upstream_version` only): `%s`\n\n", version.debian)

	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tDocker image\t\n")
	fmt.Fprintf(w, "\t----\t------------\t\n")

	for _, image := range version.dockerDevelopmentImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range version.dockerProductionImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	_ = w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())
}

func main() {
	controlFileF := flag.String("control-file", "../pg_documentdb/documentdb.control", "pg_documentdb/documentdb.control file path")
	pgVersionF := flag.String("pg-version", "17", "Major PostgreSQL version")
	debianOnlyF := flag.Bool("debian-only", false, "Only set output for Debian package version")

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

	res, err := defineVersion(controlDefaultVersion, *pgVersionF, action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	action.SetOutput("debian_version", res.debian)

	if *debianOnlyF {
		// Only 3 summaries are shown in the GitHub Actions UI by default,
		// and Docker summaries are more important (and include Debian version anyway).
		output := fmt.Sprintf("Debian package version (`upstream_version` only): `%s`", res.debian)
		action.Infof("%s", output)
		return
	}

	setSummary(action, res)

	action.SetOutput("docker_development_images", strings.Join(res.dockerDevelopmentImages, ","))
	action.SetOutput("docker_production_images", strings.Join(res.dockerProductionImages, ","))

	developmentTagFlags := make([]string, len(res.dockerDevelopmentImages))
	for i, image := range res.dockerDevelopmentImages {
		developmentTagFlags[i] = fmt.Sprintf("--tag %s", image)
	}
	action.SetOutput("docker_development_tag_flags", strings.Join(developmentTagFlags, " "))

	productionTagFlags := make([]string, len(res.dockerProductionImages))
	for i, image := range res.dockerProductionImages {
		productionTagFlags[i] = fmt.Sprintf("--tag %s", image)
	}
	action.SetOutput("docker_production_tag_flags", strings.Join(productionTagFlags, " "))
}

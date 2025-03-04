package main

import (
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/sethvargo/go-githubactions"
)

// images represents Docker image names and tags extracted from the environment.
type images struct {
	developmentImages []string
	productionImages  []string
}

// defineDockerVersion extracts Docker image names and tags from the environment variables defined by GitHub Actions.
func defineDockerVersion(controlDefaultVersion, pgVersion string, getenv githubactions.GetenvFunc) (*images, error) {
	repo := getenv("GITHUB_REPOSITORY")

	// to support GitHub forks
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to split %q into owner and name", repo)
	}
	owner := parts[0]
	repo = parts[1]

	var res *images
	var err error

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		res = defineDockerVersionForPR(pgVersion, owner, repo, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineDockerVersionForBranch(pgVersion, owner, repo, refName)

		case "tag":
			res, err = defineDockerVersionForTag(pgVersion, owner, repo, refName)

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

	slices.Sort(res.developmentImages)
	slices.Sort(res.productionImages)

	return res, nil
}

// defineDockerVersionForPR defines Docker image names and tags for PR.
func defineDockerVersionForPR(pgVersion, owner, repo, branch string) *images {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s-pr-%s", owner, repo, pgVersion, branch),
		},
	}

	// PRs are only for testing; no Quay.io and Docker Hub repos

	return res
}

// defineDockerVersionForBranch defines Docker image names and tags for branch.
func defineDockerVersionForBranch(pgVersion, owner, repo, branch string) (*images, error) {
	if branch != "ferretdb" {
		return nil, fmt.Errorf("unhandled branch %q", branch)
	}

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s-ferretdb", owner, repo, pgVersion),
		},
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res, nil
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "documentdb" {
		return res, nil
	}

	res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb-dev:%s-ferretdb", pgVersion))
	res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/postgres-documentdb-dev:%s-ferretdb", pgVersion))

	return res, nil
}

// defineDockerVersionForBranch defines Docker image names and tags for tag.
func defineDockerVersionForTag(pgVersion, owner, repo, tag string) (*images, error) {
	major, minor, patch, prerelease, err := parseGitTag(tag)
	if err != nil {
		return nil, err
	}

	tags := []string{
		fmt.Sprintf("%s-%d.%d.%d-%s", pgVersion, major, minor, patch, prerelease),
	}

	if pgVersion == "17" {
		tags = append(tags, "latest")
	}

	var res images

	for _, t := range tags {
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s", owner, repo, t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/postgres-%s:%s", owner, repo, t))
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
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb-dev:%s", t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb:%s", t))

		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/postgres-documentdb-dev:%s", t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/postgres-documentdb:%s", t))
	}

	return &res, nil
}

// dockerImageURL returns HTML page URL for the given image name and tag.
func dockerImageURL(name string) string {
	switch {
	case strings.HasPrefix(name, "ghcr.io/"):
		return fmt.Sprintf("https://%s", name)
	case strings.HasPrefix(name, "quay.io/"):
		return fmt.Sprintf("https://%s", name)
	}

	name, _, _ = strings.Cut(name, ":")

	// there is no easy way to get Docker Hub URL for the given tag
	return fmt.Sprintf("https://hub.docker.com/r/%s/tags", name)
}

// dockerSummary sets action summary.
func dockerSummary(action *githubactions.Action, version *images) {
	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tDocker image\t\n")
	fmt.Fprintf(w, "\t----\t------------\t\n")

	for _, image := range version.developmentImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range version.productionImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	_ = w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())
}

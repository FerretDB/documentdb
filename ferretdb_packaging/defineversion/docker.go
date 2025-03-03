package main

import (
	"fmt"
	"regexp"
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

// pgVer is the version of PostgreSQL.
var pgVer = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)$`)

// defineDockerVersion extracts Docker image names and tags from the environment variables defined by GitHub Actions.
func defineDockerVersion(pgVersion string, getenv githubactions.GetenvFunc) (*images, error) {
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
		res = defineDockerVersionForPR(owner, repo, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineDockerVersionForBranch(owner, repo, refName)

		case "tag":
			var major, minor, patch, prerelease string
			if major, minor, patch, prerelease, err = parseGitTag(refName); err != nil {
				return nil, err
			}

			tags := []string{
				fmt.Sprintf("%s-%s.%s.%s-%s", pgVersion, major, minor, patch, prerelease),
				"latest",
			}

			res = defineDockerVersionForTag(owner, repo, tags)

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

// defineDockerVersionForPR defines Docker image names and tags for pull requests.
func defineDockerVersionForPR(owner, repo, branch string) *images {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:pr-%s", owner, repo, branch),
		},
	}

	// PRs are only for testing; no Quay.io and Docker Hub repos

	return res
}

// defineDockerVersionForBranch defines Docker image names and tags for branch builds.
func defineDockerVersionForBranch(owner, repo, branch string) (*images, error) {
	if branch != "ferretdb" {
		return nil, fmt.Errorf("unhandled branch %q", branch)
	}

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:ferretdb", owner, repo),
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

	// res.developmentImages = append(res.developmentImages, "quay.io/ferretdb/postgres-documentdb-dev:ferretdb")
	// res.developmentImages = append(res.developmentImages, "ferretdb/postgres-documentdb-dev:ferretdb")

	return res, nil
}

// defineDockerVersionForTag defines Docker image names and tags for prerelease tag builds.
func defineDockerVersionForTag(owner, repo string, tags []string) *images {
	res := new(images)

	for _, t := range tags {
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/postgres-%s-dev:%s", owner, repo, t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/postgres-%s:%s", owner, repo, t))
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "documentdb" {
		return res
	}

	//for _, t := range tags {
	//	res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb-dev:%s", t))
	//	res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/postgres-documentdb:%s", t))
	//
	//	res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/postgres-documentdb-dev:%s", t))
	//	res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/postgres-documentdb:%s", t))
	//}

	return res
}

// setDockerTagsResults sets action output parameters, summary, etc.
func setDockerTagsResults(action *githubactions.Action, res *images) {
	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tImage\t\n")
	fmt.Fprintf(w, "\t----\t-----\t\n")

	for _, image := range res.developmentImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range res.productionImages {
		u := dockerImageURL(image)
		_, _ = fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	_ = w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	action.SetOutput("development_images", strings.Join(res.developmentImages, ","))
	action.SetOutput("production_images", strings.Join(res.productionImages, ","))
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

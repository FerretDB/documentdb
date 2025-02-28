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

// pgVer is the version of PostgreSQL with or without minor.
var pgVer = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)(?:\.(?P<minor>0|[1-9]\d*))?$`)

// defineDockerTags extracts Docker image names and tags from the environment variables defined by GitHub Actions.
func defineDockerTags(getenv githubactions.GetenvFunc) (*images, error) {
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
		res = defineForPR(owner, repo, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineForBranch(owner, repo, refName)

		case "tag":
			var major, minor, patch, prerelease string
			if major, minor, patch, prerelease, err = semVar(refName); err != nil {
				return nil, err
			}

			pgVersion := getenv("INPUT_PG_VERSION")
			pgMatch := pgVer.FindStringSubmatch(pgVersion)
			if pgMatch == nil || len(pgMatch) != pgVer.NumSubexp()+1 {
				return nil, fmt.Errorf("unexpected PostgreSQL version %q", pgVersion)
			}

			pgMajor := pgMatch[pgVer.SubexpIndex("major")]
			pgMinor := pgMatch[pgVer.SubexpIndex("minor")]

			tags := []string{
				fmt.Sprintf("%s-%s.%s.%s-%s", pgMajor, major, minor, patch, prerelease),
				"latest",
			}

			if pgMinor != "" {
				tags = append(tags, fmt.Sprintf("%s.%s-%s.%s.%s-%s", pgMajor, pgMinor, major, minor, patch, prerelease))
			}

			res = defineForTag(owner, repo, tags)

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

// defineForPR defines Docker image names and tags for pull requests.
func defineForPR(owner, repo, branch string) *images {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-dev:pr-%s", owner, repo, branch),
		},
	}

	// PRs are only for testing; no Quay.io and Docker Hub repos

	return res
}

// defineForBranch defines Docker image names and tags for branch builds.
func defineForBranch(owner, repo, branch string) (*images, error) {
	if branch != "ferretdb" {
		return nil, fmt.Errorf("unhandled branch %q", branch)
	}

	res := &images{
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-dev:ferretdb", owner, repo),
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

	res.developmentImages = append(res.developmentImages, "quay.io/ferretdb/documentdb-dev:ferretdb")
	res.developmentImages = append(res.developmentImages, "ferretdb/documentdb-dev:ferretdb")

	return res, nil
}

// defineForTag defines Docker image names and tags for prerelease tag builds.
func defineForTag(owner, repo string, tags []string) *images {
	res := new(images)

	for _, t := range tags {
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, repo, t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/%s:%s", owner, repo, t))
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "documentdb" {
		return res
	}

	for _, t := range tags {
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/documentdb-dev:%s", t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/documentdb:%s", t))

		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/documentdb-dev:%s", t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/documentdb:%s", t))
	}

	return res
}

// setDockerTagsResults sets action output parameters, summary, etc.
func setDockerTagsResults(action *githubactions.Action, res *images) {
	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tImage\t\n")
	fmt.Fprintf(w, "\t----\t-----\t\n")

	for _, image := range res.developmentImages {
		u := imageURL(image)
		_, _ = fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range res.productionImages {
		u := imageURL(image)
		_, _ = fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	_ = w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	action.SetOutput("development_images", strings.Join(res.developmentImages, ","))
	action.SetOutput("production_images", strings.Join(res.productionImages, ","))
}

// imageURL returns HTML page URL for the given image name and tag.
func imageURL(name string) string {
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

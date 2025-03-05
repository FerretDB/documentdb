package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/sethvargo/go-githubactions"
)

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

// setSummary sets action summary.
func setSummary(action *githubactions.Action, version *versions) {
	var buf strings.Builder

	fmt.Fprintf(&buf, "Debian package version (`upstream_version` only): `%s`\n\n", version.debianVersion)

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

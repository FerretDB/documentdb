package main

import (
	"fmt"
	"strings"
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

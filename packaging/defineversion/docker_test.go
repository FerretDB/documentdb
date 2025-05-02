package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageURL(t *testing.T) {
	// expected URLs should work
	assert.Equal(
		t,
		"https://ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
		dockerImageURL("ghcr.io/ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://quay.io/ferretdb/postgres-documentdb-dev:pr-docker-tag",
		dockerImageURL("quay.io/ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
	assert.Equal(
		t,
		"https://hub.docker.com/r/ferretdb/postgres-documentdb-dev/tags",
		dockerImageURL("ferretdb/postgres-documentdb-dev:pr-docker-tag"),
	)
}

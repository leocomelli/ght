package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubClientEnvVarNotFound(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")

	_, err := NewRepoTemplate()
	assert.NotNil(t, err)
	assert.Equal(t, "GITHUB_TOKEN is not set", err.Error())
}

func TestGitHubClient(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "1234567890")

	res, err := NewRepoTemplate()
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

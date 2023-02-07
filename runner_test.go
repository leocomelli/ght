package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

var mocks = map[string]func() mock.MockBackendOption{
	"GetRepo_500": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.GetReposByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusInternalServerError,
					"500 Internal Server Error",
				)
			}),
		)
	},
	"GetRepo_404": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.GetReposByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusNotFound,
					"404 Not Found",
				)
			}),
		)
	},
	"GetRepo": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			github.Repository{
				Owner:       &github.User{Login: github.String("leocomelli")},
				Name:        github.String("ght"),
				Description: github.String("A simple CLI to create GitHub repositories"),
			},
		)
	},
	"CreateRepo": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.PostOrgsReposByOrg,
			github.Repository{
				Owner:       &github.User{Login: github.String("leocomelli")},
				Name:        github.String("ght"),
				Description: github.String("A simple CLI to create GitHub repositories"),
			},
		)
	},
	"CreateRepo_400": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.PostOrgsReposByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusBadRequest,
					"400 Bad Request",
				)
			}),
		)
	},
	"CreateRepoTemplate": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.PostReposGenerateByTemplateOwnerByTemplateRepo,
			github.Repository{
				Owner:       &github.User{Login: github.String("leocomelli")},
				Name:        github.String("ght"),
				Description: github.String("A simple CLI to create GitHub repositories"),
			},
		)
	},
	"CreateRepoTemplate_400": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.PostReposGenerateByTemplateOwnerByTemplateRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusBadRequest,
					"400 Bad Request",
				)
			}),
		)
	},
	"ReplaceTopics": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.PutReposTopicsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusOK,
					"200 OK",
				)
			}),
		)
	},
	"GetOrg": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.GetOrgsByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusOK,
					"200 OK",
				)
			}),
		)
	},
	"GetOrg_404": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.GetOrgsByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusNotFound,
					"404 Not Found",
				)
			}),
		)
	},
	"GetBranch": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.GetReposBranchesByOwnerByRepoByBranch,
			github.Branch{
				Name: github.String("main"),
			},
		)
	},
	"UpdateBranchProtection": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.PutReposBranchesProtectionByOwnerByRepoByBranch,
			github.BranchProtectionRule{},
		)
	},
	"UpdateBranchProtection_400": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.PutReposBranchesProtectionByOwnerByRepoByBranch,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusBadRequest,
					"400 Bad Request",
				)
			}),
		)
	},
	"UpdateBranchProtectionSignCommit": func() mock.MockBackendOption {
		return mock.WithRequestMatch(
			mock.PostReposBranchesProtectionRequiredSignaturesByOwnerByRepoByBranch,
			github.SignaturesProtectedBranch{},
		)
	},
	"UpdateBranchProtectionSignCommit_400": func() mock.MockBackendOption {
		return mock.WithRequestMatchHandler(
			mock.PostReposBranchesProtectionRequiredSignaturesByOwnerByRepoByBranch,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusBadRequest,
					"400 Bad Request",
				)
			}),
		)
	},
}

func TestTemplateLocalFileNotFound(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/nonexistent.json",
	}

	_, err := LoadRepoConfig(opts)
	assert.NotNil(t, err)
	assert.IsType(t, &os.PathError{}, errors.Unwrap(err))
}

func TestTemplateLocalFile(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/existing-repo.json",
	}

	cfg, err := LoadRepoConfig(opts)
	assert.Nil(t, err)
	assert.True(t, cfg.BranchProtection.EnforceAdmins)
}

func TestTemplateInvalidLocalFile(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/invalid-syntax.json",
	}

	_, err := LoadRepoConfig(opts)
	assert.NotNil(t, err)
	assert.IsType(t, &json.SyntaxError{}, errors.Unwrap(err))
}

func TestTemplateRemoteFileNotFound(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "https://nonexistent.com/nonexistent.json",
	}

	_, err := LoadRepoConfig(opts)
	assert.NotNil(t, err)
	assert.IsType(t, &url.Error{}, errors.Unwrap(err))
}

func TestTemplateRemoteFile(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "https://raw.githubusercontent.com/leocomelli/ght/main/testing/existing-repo.json",
	}

	cfg, err := LoadRepoConfig(opts)
	assert.Nil(t, err)
	assert.True(t, cfg.BranchProtection.EnforceAdmins)
}

func TestErrorLoadingRepoConfig(t *testing.T) {
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/nonexistent.json",
	}

	_, err := Run(nil, opts)

	assert.NotNil(t, err)
	assert.IsType(t, &os.PathError{}, errors.Unwrap(err))
}

func TestInternalServerErrorWhenGetRepo(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_500"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/empty.json",
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	assert.True(t, err.(*github.ErrorResponse).Response.StatusCode == http.StatusInternalServerError)
}

func TestGetRepo(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo"](),
		mocks["ReplaceTopics"](),
		mocks["GetOrg_404"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/empty.json",
	}

	res, err := Run(rt, opts)

	assert.Nil(t, err)
	assert.Equal(t, "leocomelli/ght", res.Fullname)
	assert.Equal(t, false, res.Created)
}

func TestCreateWithNoRepoConfig(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/empty.json",
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	assert.Equal(t, "no repository section in template file", err.Error())
}

func TestCreateSimpleRepo(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["GetOrg"](),
		mocks["CreateRepo"](),
		mocks["ReplaceTopics"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/simple-repo.json",
	}

	res, err := Run(rt, opts)

	assert.Nil(t, err)
	assert.Equal(t, "leocomelli/ght", res.Fullname)
	assert.Equal(t, true, res.Created)
}

func TestErrorCreatingSimpleRepo(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["GetOrg"](),
		mocks["CreateRepo_400"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/simple-repo.json",
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	err = errors.Unwrap(err)
	assert.IsType(t, &github.ErrorResponse{}, err)
	assert.Equal(t, http.StatusBadRequest, err.(*github.ErrorResponse).Response.StatusCode)
}

func TestCreateSimpleRepoUsingTemplate(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["CreateRepoTemplate"](),
		mocks["ReplaceTopics"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/simple-repo-template.json",
	}

	res, err := Run(rt, opts)

	assert.Nil(t, err)
	assert.Equal(t, "leocomelli/ght", res.Fullname)
	assert.Equal(t, true, res.Created)
}

func TestErrorCreatingSimpleRepoUsingTemplate(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["CreateRepoTemplate_400"](),
		mocks["ReplaceTopics"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/simple-repo-template.json",
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	err = errors.Unwrap(err)
	assert.IsType(t, &github.ErrorResponse{}, err)
	assert.Equal(t, http.StatusBadRequest, err.(*github.ErrorResponse).Response.StatusCode)
}

func TestCreateRepoWithBranchProtection(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["GetOrg"](),
		mocks["CreateRepo"](),
		mocks["ReplaceTopics"](),
		mocks["GetBranch"](),
		mocks["UpdateBranchProtection"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/repo-branch-protection.json",
		Branches: []string{"main"},
	}

	res, err := Run(rt, opts)

	assert.Nil(t, err)
	assert.Equal(t, "leocomelli/ght", res.Fullname)
	assert.Equal(t, true, res.Created)
}

func TestErrorBranchNotFoundCreatingRepoWithBranchProtection(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["CreateRepo"](),
		mocks["GetOrg"](),
		mocks["ReplaceTopics"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/repo-branch-protection.json",
		Branches: []string{"main"},
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	assert.Equal(t, "failed to get branch main. check if the branch exists; if you are creating a new repository use the auto_init option |→ unexpected status code: 404 Not Found", err.Error())
}

func TestErrorCreatingRepoWithBranchProtection(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["GetOrg"](),
		mocks["CreateRepo"](),
		mocks["GetBranch"](),
		mocks["UpdateBranchProtection_400"](),
		mocks["ReplaceTopics"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/repo-branch-protection.json",
		Branches: []string{"main"},
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	err = errors.Unwrap(err)
	assert.IsType(t, &github.ErrorResponse{}, err)
	assert.Equal(t, http.StatusBadRequest, err.(*github.ErrorResponse).Response.StatusCode)
}

func TestCreateRepoWithBranchProtectionAndSignedCommit(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["CreateRepo"](),
		mocks["GetOrg"](),
		mocks["ReplaceTopics"](),
		mocks["GetBranch"](),
		mocks["UpdateBranchProtection"](),
		mocks["UpdateBranchProtectionSignCommit"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/repo-branch-protection-complete.json",
		Branches: []string{"main"},
	}

	res, err := Run(rt, opts)

	assert.Nil(t, err)
	assert.Equal(t, "leocomelli/ght", res.Fullname)
	assert.Equal(t, true, res.Created)
}

func TestErrorCreatingRepoWithBranchProtectionAndSignedCommit(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mocks["GetRepo_404"](),
		mocks["CreateRepo"](),
		mocks["GetOrg"](),
		mocks["ReplaceTopics"](),
		mocks["GetBranch"](),
		mocks["UpdateBranchProtection"](),
		mocks["UpdateBranchProtectionSignCommit_400"](),
	)

	rt := &RepoTemplate{client: github.NewClient(mockedHTTPClient)}
	opts := &RepoOptions{
		Owner:    "leocomelli",
		Name:     "ght",
		Template: "./testing/repo-branch-protection-complete.json",
		Branches: []string{"main"},
	}

	_, err := Run(rt, opts)

	assert.NotNil(t, err)
	err = errors.Unwrap(err)
	assert.IsType(t, &github.ErrorResponse{}, err)
	assert.Equal(t, http.StatusBadRequest, err.(*github.ErrorResponse).Response.StatusCode)
}

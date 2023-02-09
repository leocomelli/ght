package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
)

const (
	PullRequestTemplate = ".github/pull_request_template.md"
	IssueTemplate       = ".github/issue_template.md"
)

var (
	// ErrRepoConfigNotFound is returned when no repository section is found in the template file
	ErrRepoConfigNotFound = errors.New("no repository section in template file")
)

// RepoTemplate represents the action output
type RepoResponse struct {
	Fullname string
	Created  bool
}

// Run performs the actions according to the repo config
func Run(rt *RepoTemplate, opts *RepoOptions) (*RepoResponse, error) {
	res := &RepoResponse{
		Fullname: fmt.Sprintf("%s/%s", opts.Owner, opts.Name),
	}

	cfg, err := LoadRepoConfig(opts)
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("Loading repo config from %s", opts.Template)

	// Check if repo exists
	_, err = rt.GetRepo(opts.Owner, opts.Name)
	err = errors.Unwrap(err)
	if err != nil && err.(*github.ErrorResponse).Response.StatusCode != http.StatusNotFound {
		return nil, err
	}

	// Create repo if it doesn't exist
	if err != nil && err.(*github.ErrorResponse).Response.StatusCode == http.StatusNotFound {
		if cfg.Repository == nil && cfg.TemplateRepo == nil {
			return nil, ErrRepoConfigNotFound
		}

		if _, err := rt.CreateRepo(opts, cfg); err != nil {
			return nil, err
		}
		res.Created = true
	}

	// Replace topics
	if opts.Topics != nil && len(opts.Topics) > 0 {
		if err := rt.ReplaceTopics(opts.Owner, opts.Name, opts.Topics); err != nil {
			return nil, err
		}
	}

	// Configure issue template
	if cfg.PullRequestTemplate != "" {
		if err := CreateOrUpdateContent(rt, opts.Owner, opts.Name, PullRequestTemplate, cfg.PullRequestTemplate); err != nil {
			return nil, err
		}
	}

	// Issue issue template
	if cfg.IssueTemplate != "" {
		if err := CreateOrUpdateContent(rt, opts.Owner, opts.Name, IssueTemplate, cfg.IssueTemplate); err != nil {
			return nil, err
		}
	}

	// Update branch protection rules
	if cfg.BranchProtection != nil {
		if err := rt.BranchProtectionRules(opts.Owner, opts.Name, opts.Branches, cfg.BranchProtection, cfg.RequiredSignedCommits); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func CreateOrUpdateContent(rt *RepoTemplate, owner, repo, ghPath, path string) error {
	prTmplData, err := Data(path)
	if err != nil {
		return err
	}

	if err := rt.CreateUpdateContent(owner, repo, ghPath, prTmplData); err != nil {
		return err
	}

	return nil
}

// LoadRepoConfig loads the repository config from a file or url
func LoadRepoConfig(opts *RepoOptions) (*Config, error) {
	data, err := Data(opts.Template)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json |→ %w", err)
	}

	return cfg, nil
}

// Data returns the data from a file or url
func Data(path string) ([]byte, error) {
	if strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get url %s |→ %w", path, err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from url %s |→ %w", path, err)
		}

		return body, nil
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s |→ %w", path, err)
	}

	return file, nil
}

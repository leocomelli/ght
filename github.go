package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v50/github"
)

// RepoOptions is a wrapper around the github.Client
type RepoTemplate struct {
	client *github.Client
}

// NewRepoTemplate creates a new RepoTemplate.
func NewRepoTemplate() (*RepoTemplate, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	ctx := context.Background()
	cli := github.NewTokenClient(ctx, token)

	return &RepoTemplate{
		client: cli,
	}, nil
}

// GetOrg fetches an organization.
//
// GitHub API docs: https://docs.github.com/en/rest/reference/orgs#get-an-organization
func (r *RepoTemplate) GetOrg(org string) (*github.Organization, error) {
	ctx := context.Background()

	logger.Debug().Msgf("fetching org %s", org)

	res, _, err := r.client.Organizations.Get(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch org %s |→ %w", org, err)
	}

	return res, nil
}

// GetRepo fetches a repository.
//
// GitHub API docs: https://docs.github.com/en/rest/repos/repos#get-a-repository
func (r *RepoTemplate) GetRepo(owner, repo string) (*github.Repository, error) {
	ctx := context.Background()

	logger.Debug().Msgf("fetching repo %s/%s", owner, repo)

	res, _, err := r.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s |→ %w", owner, repo, err)
	}

	return res, nil
}

// CreateRepo creates a repository from scratch or using a template.
//
// GitHub API docs: https://docs.github.com/en/rest/reference/repos#create-a-repository-for-the-authenticated-user
// GitHub API docs: https://docs.github.com/en/rest/reference/repos#create-a-repository-using-a-template
func (r *RepoTemplate) CreateRepo(opts *RepoOptions, cfg *Config) (*github.Repository, error) {
	ctx := context.Background()

	repo := cfg.Repository

	logger.Debug().Msgf("creating repo %s/%s", opts.Owner, opts.Name)

	// Create a repo using a template.
	if cfg.TemplateRepo != nil {
		logger.Debug().Msg("using template repo")

		tmpl := &github.TemplateRepoRequest{
			Name:               github.String(opts.Name),
			Owner:              github.String(opts.Owner),
			Description:        github.String(opts.Description),
			IncludeAllBranches: cfg.TemplateRepo.IncludeAllBranches,
			Private:            cfg.TemplateRepo.Private,
		}

		res, _, err := r.client.Repositories.CreateFromTemplate(ctx, opts.Owner, opts.Name, tmpl)
		if err != nil {
			return nil, fmt.Errorf("failed to create repo using template %s/%s |→ %w", opts.Owner, opts.Name, err)
		}

		return res, nil
	}

	// Check if the owner is an organization.
	owner := opts.Owner
	_, err := r.GetOrg(opts.Owner)
	if err != nil {
		err = errors.Unwrap(err)
		if err.(*github.ErrorResponse).Response.StatusCode == http.StatusNotFound {
			owner = ""
		}
	}

	// Create a repo from scratch.
	repo.Name = github.String(opts.Name)
	repo.Description = github.String(opts.Description)
	repo.Topics = opts.Topics

	res, _, err := r.client.Repositories.Create(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo %s/%s |→ %w", opts.Owner, opts.Name, err)
	}

	return res, nil
}

// GetBranch fetches a branch.
//
// GitHub API docs: https://docs.github.com/en/rest/reference/repos#get-a-branch
func (r *RepoTemplate) GetBranch(org, name, branch string) (*github.Branch, error) {
	ctx := context.Background()
	b, _, err := r.client.Repositories.GetBranch(ctx, org, name, branch, true)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// BranchProtectionRules sets branches protection rules.
//
// Github API docs: https://docs.github.com/en/rest/reference/repos#update-branch-protection
func (r *RepoTemplate) BranchProtectionRules(owner, repo string, branches []string, protection *github.ProtectionRequest, signedCommits bool) error {
	ctx := context.Background()

	for _, branch := range branches {
		logger.Debug().Msgf("setting branch protection rules on %s", branch)

		_, err := r.GetBranch(owner, repo, branch)
		if err != nil {
			return fmt.Errorf("failed to get branch %s. check if the branch exists; if you are creating a new repository use the auto_init option |→ %w", branch, err)
		}

		_, _, err = r.client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, protection)
		if err != nil {
			return fmt.Errorf("failed to set branch protection rules on %s |→ %w", branch, err)
		}

		if signedCommits {
			if err := r.BranchCommitSignProtection(owner, repo, branch); err != nil {
				return fmt.Errorf("failed to set branch protection rules for signed commits on %s |→ %w", branch, err)
			}
		}
	}

	return nil
}

// BranchCommitSignProtection sets branch protection rules for signed commits.
//
// Github API docs: https://docs.github.com/en/rest/reference/repos#require-commit-signature-protection
func (r *RepoTemplate) BranchCommitSignProtection(owner, repo string, branch string) error {
	ctx := context.Background()

	logger.Debug().Msgf("setting branch protection rules for signed commits on %s", branch)

	_, _, err := r.client.Repositories.RequireSignaturesOnProtectedBranch(ctx, owner, repo, branch)
	if err != nil {
		return err
	}

	return nil
}

// ReplaceTopics replaces the topics of a repository.
//
// Github API docs: https://docs.github.com/en/rest/reference/repos#replace-all-topics-for-a-repository
func (r *RepoTemplate) ReplaceTopics(owner, repo string, topics []string) error {
	ctx := context.Background()

	_, _, err := r.client.Repositories.ReplaceAllTopics(ctx, owner, repo, topics)
	if err != nil {
		return fmt.Errorf("failed to replace topics on %s/%s |→ %w", owner, repo, err)
	}

	return nil
}

// CreateUpdateContent creates or updates a file in a repository.
//
// Github API docs: https://docs.github.com/en/rest/reference/repos#create-or-update-file-contents
// Github API docs: https://docs.github.com/en/rest/reference/repos#update-a-file
func (r *RepoTemplate) CreateUpdateContent(owner, repo, path string, content []byte) error {
	ctx := context.Background()

	getOpts := &github.RepositoryContentGetOptions{
		Ref: "main",
	}
	res, _, _, err := r.client.Repositories.GetContents(ctx, owner, repo, ".github/pull_request_template.md", getOpts)
	if err != nil && err.(*github.ErrorResponse).Response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to get file %s/%s/%s |→ %w", owner, repo, path, err)
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("Add/Update Pull Request Template"),
		Content: content,
	}

	if res != nil && res.SHA != nil {
		opts.SHA = res.SHA

		_, _, err := r.client.Repositories.UpdateFile(ctx, owner, repo, path, opts)

		if err != nil {
			return fmt.Errorf("failed to update file %s/%s/%s |→ %w", owner, repo, path, err)
		}

		return nil
	}

	_, _, err = r.client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to create file %s/%s/%s |→ %w", owner, repo, path, err)
	}

	return nil
}

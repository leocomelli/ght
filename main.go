package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// RepoOptions is the options for creating a new repository
type RepoOptions struct {
	Name        string
	Owner       string
	Description string
	Topics      []string
	Branches    []string
	Template    string
	Debug       bool
}

// Config is the configuration for the repository
type Config struct {
	Repository            *github.Repository          `json:"repository"`
	BranchProtection      *github.ProtectionRequest   `json:"branch_protection"`
	TemplateRepo          *github.TemplateRepoRequest `json:"template_repo"`
	RequiredSignedCommits bool                        `json:"required_signed_commits"`
	PullRequestTemplate   string                      `json:"pull_request_template"`
	IssueTemplate         string                      `json:"issue_template"`
}

const tmpl = `
Version: %s
BuildDate: %s
GitCommit: %s
`

var (
	// Version contains the current version of the app.
	Version = ""
	// BuildDate contains the date and time of build process.
	BuildDate = ""
	// GitHash contains the hash of last commit in the repository.
	GitHash = ""

	logger zerolog.Logger
	opts   *RepoOptions
)

func main() {
	cmd := command()

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("error executing command")
		os.Exit(1)
	}
}

func command() *cobra.Command {
	opts = &RepoOptions{}

	root := &cobra.Command{
		Use:   "ght",
		Short: "ght is a CLI tool for creating a new repository based on the template",
	}

	repo := &cobra.Command{
		Use:     "repo",
		Aliases: []string{"r", "repository"},
		Short:   "Create a new repository based on the template",
		RunE: func(cmd *cobra.Command, args []string) error {
			debugMode(opts)

			var (
				rt  *RepoTemplate
				err error
			)

			if rt, err = NewRepoTemplate(); err != nil {
				return err
			}

			if _, err := Run(rt, opts); err != nil {
				logger.Error().Err(err).Msg("")
			}

			return nil
		},
	}

	repo.Flags().StringVarP(&opts.Name, "name", "n", "", "the name of the repository")
	repo.Flags().StringVarP(&opts.Owner, "owner", "o", "", "the name of the owner, can be an organization or an authenticated user")
	repo.Flags().StringVarP(&opts.Description, "description", "d", "", "a short description of the repository")
	repo.Flags().StringSliceVarP(&opts.Topics, "topics", "l", []string{}, "an array of topics to add to the repository")
	repo.Flags().StringSliceVarP(&opts.Branches, "branches", "b", []string{}, "the names of the branches to which the protection rules will be applied")
	repo.Flags().StringVarP(&opts.Template, "template", "t", "", "the name of the JSON file contains the template, can be a local or remote file")
	repo.Flags().BoolVarP(&opts.Debug, "debug", "v", false, "enable debug mode")

	_ = repo.MarkFlagRequired("owner")
	_ = repo.MarkFlagRequired("name")
	_ = repo.MarkFlagRequired("template")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of ght",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(tmpl, Version, BuildDate, GitHash)
			return nil
		},
	}

	root.AddCommand(repo)
	root.AddCommand(version)

	return root
}

func debugMode(opts *RepoOptions) {
	level := zerolog.InfoLevel
	if opts.Debug {
		level = zerolog.DebugLevel
	}

	logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(level).
		With().
		Timestamp().
		Logger()

	logger.Debug().Msg("debug mode enabled")
}

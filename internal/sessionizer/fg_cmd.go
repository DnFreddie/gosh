/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"errors"
	"fmt"
	"os"

	"github.com/DnFreddie/gosh/pkg/github"
	"github.com/spf13/cobra"
)

var (
	_git_url  string // GitHub repository URL to clone
	_reposDir string // Directory where repositories are stored
)

var FgCmd = &cobra.Command{
	Use:   "fg",
	Short: "Select a GitHub repo, clone it if absent, and create a session",
	Long: `Uses the GitHub API to list available repositories. If the repository does not exist in
~/HOME/github.com/DnFreddie, it will be cloned and a tmux session created. Otherwise, it switches to an existing session.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if _reposDir == "" {
			_reposDir = os.ExpandEnv("${HOME}/github.com")
		}
		if _git_url != "" {
			return cloneAndTmux(_git_url, _reposDir)
		}

		if err := Fg(_reposDir); err != nil {
			return fmt.Errorf("error in Fg operation: %w", err)
		}
		return nil
	},
}

func init() {
	FgCmd.Flags().StringVarP(&_git_url, "url", "u", "", "URL to clone and create session for")
	FgCmd.Flags().StringVarP(&_reposDir, "path", "p", "", "Directory where the repository is stored")
}

func cloneAndTmux(url, reposDir string) error {
	repo, err := github.NewRepo(url)

	if err != nil {
		return fmt.Errorf("error creating repo: %w", err)
	}

	if err := repo.Clone(reposDir); err != nil && !errors.Is(err, github.RepoExistErr{}) {
		return fmt.Errorf("error cloning repo: %w", err)
	}

	return handleTmuxSession(repo.Name, repo.Path)
}

func handleTmuxSession(repoName, repoPath string) error {
	tmux, err := NewTmux(repoName, repoPath)
	if err != nil {
		return err
	}
	if err := tmux.CreateSession(); err != nil {
		return fmt.Errorf("error creating tmux session: %w", err)
	}
	return nil
}

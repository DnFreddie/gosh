/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
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

var _url string
var _reposDir string

var FgCmd = &cobra.Command{
	Use:   "fg",
	Short: "Chose or a github repo if not exitent download it and create a session",
	Long: `Uses github api to list available repos then if not exist in ~/HOME/github.com/DnFreddie
it will clone it and switch session  else it will create a tmux session
	`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _reposDir == "" {
			_reposDir = os.ExpandEnv("${HOME}/github.com")
		}
		if _url != "" {
			repo, err := github.NewRepo(_url)
			if err != nil {
				return fmt.Errorf("error creating repo: %w", err)
			}
			if err := repo.Clone(_reposDir); !errors.Is(err, github.RepoExistErr{}) {
				return fmt.Errorf("error cloning repo: %w", err)
			}

			tmux, err := NewTmux(repo.Name, repo.Path)
			if err != nil {
				return fmt.Errorf("error creating tmux instance: %w", err)
			}
			if err := tmux.CreateSession(); err != nil {
				return fmt.Errorf("error creating tmux session: %w", err)
			}
			return nil
		}
		if err := Fg(_reposDir); err != nil {
			return fmt.Errorf("error in Fg operation: %w", err)
		}
		return nil
	},
}

func init() {
	FgCmd.Flags().StringVarP(&_url, "url", "u", "", "URL to clone and change session")
	FgCmd.Flags().StringVarP(&_reposDir, "path", "p", "", "Path to were repo is stored")
}

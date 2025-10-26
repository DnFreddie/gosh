package sessionizer

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/DnFreddie/gosh/pkg/busybox"
	"github.com/spf13/cobra"
)

var TnCmd = &cobra.Command{
	Use:   "tn [command]",
	Short: "Manage and switch between tmux sessions",
	Long: `tn is a lightweight tmux helper.

Available commands:
  create   Create a new tmux session in the current or specified directory
  window   Switch between windows across sessions
  (run without arguments to switch sessions)
`,

	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		tmux, err := NewTmux()
		if err != nil {
			return err
		}
		return tmux.Tn()
	},
}

var createCmd = &cobra.Command{
	Use:          "create [path]",
	Aliases:      []string{"c"},
	Short:        "Create a new session in current or specified directory",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetDir string
		var err error

		if len(args) > 0 {
			targetDir, err = filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if _, err := os.Stat(targetDir); err != nil {
				if os.IsNotExist(err) {
					return errors.New("The specified path doesn't exist")
				}
				return err
			}
		} else {
			targetDir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		tmux, err := NewTmux()
		if err != nil {
			return err
		}

		sessionName := path.Base(targetDir)
		return tmux.CreateSession(sessionName, targetDir)
	},
}

// Windows subcommand
var windowsCmd = &cobra.Command{
	Use:          "windows",
	Aliases:      []string{"w"},
	Short:        "Switch between windows across all sessions",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		tmux, err := NewTmux()
		if err != nil {
			return err
		}

		windows, err := tmux.GetWindows("*")
		if err != nil {
			return err
		}

		choice, err := busybox.RunTerm(windows, func(w TmuxWindow) string {
			return fmt.Sprintf("%s  %s  %s",
				busybox.InColors(busybox.Yellow, fmt.Sprintf("%d", w.Index)),
				busybox.InColors(busybox.BrightGreen, w.SessionName),
				busybox.InColors(busybox.BrightWhite, w.Name),
			)
		})
		if err != nil {
			return err
		}

		return tmux.SwitchWindow(choice.SessionName, choice.Index)
	},
}

func init() {
	TnCmd.AddCommand(createCmd)
	TnCmd.AddCommand(windowsCmd)
}

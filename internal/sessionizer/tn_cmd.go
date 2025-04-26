package sessionizer

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

var _create bool
var _path string

var TnCmd = &cobra.Command{
	Use:          "tn",
	Short:        "Switch to session or create a new one",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _create {
			var targetDir string
			var err error

			if _path != "" {
				targetDir, err = filepath.Abs(_path)
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

			tmux, err := NewTmux("", targetDir)

			if err != nil {
				return err
			}
			tmux.SetName(path.Base(targetDir))

			return tmux.CreateSession()
		}

		tmux, err := NewTmux(_path, "")

		if err != nil {
			return err
		}
		return tmux.Tn()
	},
}

func init() {
	TnCmd.Flags().BoolVarP(&_create, "create", "c", false, "Create a new session")
	TnCmd.Flags().StringVarP(&_path, "path", "p", "", "Path for new session (optional)")
}

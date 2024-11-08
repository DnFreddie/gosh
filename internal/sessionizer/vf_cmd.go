/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"github.com/spf13/cobra"
)

// vfCmd represents the tn command
var VfCmd = &cobra.Command{
	Use:          "vf",
	Short:        "Look on directories in home and open one in nvim in tmux pane",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := Vf(); err != nil {
			return err
		}
		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tnCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tnCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

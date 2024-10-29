/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/DnFreddie/gosh/internal/sessionizer"
	"github.com/spf13/cobra"
)

// sessionizerCmd represents the sessionizer command
var sessionizerCmd = &cobra.Command{
	Use:     "sessionizer [fd/fs/fg/vf]",
	Aliases: []string{"ss"},
	Short:   "Maniputlate tmux sessions",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(sessionizerCmd)
	sessionizerCmd.AddCommand(sessionizer.FsCmd)
	sessionizerCmd.AddCommand(sessionizer.VfCmd)
	sessionizerCmd.AddCommand(sessionizer.FgCmd)
	sessionizerCmd.AddCommand(sessionizer.FdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sessionizerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sessionizerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

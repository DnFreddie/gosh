/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/DnFreddie/gosh/internal/editor"
	"github.com/spf13/cobra"
)

// fmCmd represents the fm command
var editCmd = &cobra.Command{
	Use:     "edit [br/]",
	Short:   "Edit with vim",
	Long:    `Edit with vim or vi if not available `,
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		editor.Edit(strings.Join(args, " "))

	},
}

func init() {

	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(editCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

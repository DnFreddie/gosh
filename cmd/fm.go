/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/DnFreddie/gosh/internal/fm"
	"github.com/spf13/cobra"
)

// fmCmd represents the fm command
var fmCmd = &cobra.Command{
	Use:   "fm [br/]",
	Short: "Stirngs manpitulations on the command line",
	Long:  `Thsi contains the scripts that i use for text manipulation `,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()

	},
}

func init() {

	// Here you will define your flags and configuration settings.
	fmCmd.AddCommand(fm.Br)
	rootCmd.AddCommand(fmCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

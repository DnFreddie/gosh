/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"github.com/spf13/cobra"
)

// fsCmd represents the fs command
var FsCmd = &cobra.Command{
	Use:          "fs",
	Short:        "Read .ssh/config and create or switch the tmux session ",
	SilenceUsage: true,
	Long:         `Chose the Host and ssh into it wiht new tmux session created`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := Fs(); err != nil {
			return err
		}
		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

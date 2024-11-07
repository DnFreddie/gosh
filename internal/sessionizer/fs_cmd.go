/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"fmt"

	"github.com/spf13/cobra"
)

// fsCmd represents the fs command
var FsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Read .ssh/config and create or switch the tmux session ",
	Long:  `Have to be inside  tmux and requiers fzf  `,
	Run: func(cmd *cobra.Command, args []string) {
		if err := Fs(); err != nil {
			fmt.Println(err)
			return
		}
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

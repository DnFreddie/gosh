/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"fmt"

	"github.com/DnFreddie/gosh/internal"
	"github.com/DnFreddie/gosh/scripts"
	"github.com/spf13/cobra"
)

// vfCmd represents the tn command
var VfCmd = &cobra.Command{
	Use:   "vf",
	Short: "Look on directories in home and open one in nvim in tmux pane",
	Run: func(cmd *cobra.Command, args []string) {

		if err := internal.RunScript(scripts.Vf); err != nil {
			fmt.Println(err)
			return
		}
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

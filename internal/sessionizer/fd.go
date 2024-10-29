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

// fdCmd represents the fd command
var FdCmd = &cobra.Command{
	Use:   "fd",
	Short: "List all dirs in home and lest u create the session else switches to one",
	Run: func(cmd *cobra.Command, args []string) {
		if err := internal.RunScript(scripts.Fd); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

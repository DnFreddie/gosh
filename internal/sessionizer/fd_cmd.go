/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"github.com/spf13/cobra"
)

// fdCmd represents the fd command
var FdCmd = &cobra.Command{
	Use:          "fd",
	Short:        "List all dirs in home and lest u create the session else switches to one",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := Fd(); err != nil {

			return err
		}
		return nil
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

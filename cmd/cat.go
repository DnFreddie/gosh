/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/DnFreddie/gosh/pkg/busybox"
	"github.com/spf13/cobra"
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:     "cat",
	Aliases: []string{"c"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		for _, v := range args {
			filePath := v
			abs, err := filepath.Abs(filePath)
			if err != nil {
				fmt.Println(err)
				continue
			}
			f, err := os.Open(abs)
			defer func() {
				f.Close()
			}()
			if err != nil {
				fmt.Println(err)
				continue
			}
			h := busybox.NewHighlightedPager(filePath, io.Reader(f))
			if err := h.Run(); err != nil {
				fmt.Println(err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(catCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// catCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// catCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

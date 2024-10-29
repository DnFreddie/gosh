/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package sessionizer

import (
	"fmt"
	"net/url"

	"github.com/DnFreddie/gosh/internal"
	"github.com/DnFreddie/gosh/scripts"
	"github.com/spf13/cobra"
)

var url_string string

var FgCmd = &cobra.Command{
	Use:   "fg",
	Short: "Chose or a github repo if not exitent download it and create a session ",
	Long: `Uses gh to list available repos then if not exist in ~/HOME/github.com/DnFreddie
it will clone it and switch session  else it will create a tmux session

	`,
	Run: func(cmd *cobra.Command, args []string) {
		if url_string != "" {
			if _, err := url.ParseRequestURI(url_string); err != nil {
				fmt.Println("This is not valid git repository")
				return
			}
		}

		if err := internal.RunScript(scripts.Fg, url_string); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {

	FgCmd.Flags().StringVarP(&url_string, "url", "u", "", "URL to clone and change session")
}

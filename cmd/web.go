package cmd

import (
	"fmt"

	"github.com/dsseng/wiso/pkg/web"
	"github.com/spf13/cobra"
)

var (
	webUrl string
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start a web interface to perform user auth and admin access",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting web server as %v\n", webUrl)
		web.Start(webUrl, db)
	},
}

func init() {
	webCmd.PersistentFlags().StringVarP(&webUrl,
		"url",
		"u",
		"http://localhost:8989/",
		"Specifies port and base address",
	)
	rootCmd.AddCommand(webCmd)
}

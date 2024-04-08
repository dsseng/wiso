package cmd

import (
	"fmt"

	"github.com/dsseng/wiso/pkg/web"
	"github.com/spf13/cobra"
)

var (
	webPort uint16
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start a web interface to perform user auth and admin access",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting web server on port %d\n", webPort)
		web.Start(webPort, db)
	},
}

func init() {
	webCmd.PersistentFlags().Uint16VarP(&webPort,
		"port",
		"p",
		8989,
		"Port to listen at",
	)
	rootCmd.AddCommand(webCmd)
}

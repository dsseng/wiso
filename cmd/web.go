package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/dsseng/wiso/pkg/oidc"
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
		url, err := url.Parse(webUrl)
		if err != nil {
			fmt.Println("Error parsing URL: ", err.Error())
			return
		}
		fmt.Printf("Starting web server as %v\n", webUrl)
		app := web.App{
			DB:      db,
			BaseURL: url,
			OIDC: &oidc.OIDCProvider{
				ClientID:     os.Getenv("CLIENT_ID"),
				ClientSecret: os.Getenv("CLIENT_SECRET"),
				Issuer:       os.Getenv("ISSUER"),
				ID:           "gitea",
				Name:         "Gitea",
			},
			PasswordAuth: false,
			LogoLogin:    "/static/logo.png",
			LogoWelcome:  "/static/logo-welcome.png",
			LogoError:    "/static/logo-error.png",
			SupportURL:   "https://github.com/dsseng",
		}
		app.Start()
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

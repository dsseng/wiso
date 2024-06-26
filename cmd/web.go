package cmd

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/dsseng/wiso/pkg/users"
	"github.com/dsseng/wiso/pkg/web"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	configPath string
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start a web interface to perform user auth and admin access",
	Run: func(cmd *cobra.Command, args []string) {
		app := web.App{
			Database:        "",
			DB:              db,
			Base:            "http://localhost:8989/",
			OIDC:            nil,
			LDAP:            nil,
			LogoLogin:       "/static/logo.png",
			LogoWelcome:     "/static/logo-welcome.png",
			LogoError:       "/static/logo-error.png",
			SupportURL:      "https://github.com/dsseng",
			CleanupInterval: time.Hour,
		}
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Error reading config file: %v\n", err)
			return
		}

		if err := yaml.Unmarshal(data, &app); err != nil {
			fmt.Printf("Error unmarshaling config: %v\n", err)
			return
		}

		app.BaseURL, err = url.Parse(app.Base)
		if err != nil {
			fmt.Printf("Error parsing URL: %v\n", err)
			return
		}

		if app.DB == nil {
			var err error
			app.DB, err = gorm.Open(postgres.Open(app.Database))
			if err != nil {
				fmt.Println("Failed to connect database", err)
			}
		}

		app.DB.AutoMigrate(&users.DeviceSession{})
		app.DB.AutoMigrate(&users.User{})

		go (func() {
			users.CleanupOutdatedSessions(app.DB)
			for range time.Tick(app.CleanupInterval) {
				users.CleanupOutdatedSessions(app.DB)
			}
		})()

		fmt.Printf("Starting web server as %v\n", app.BaseURL.String())
		err = app.Start()
		if err != nil {
			fmt.Println("Failed to start web server", err)
		}
	},
}

func init() {
	webCmd.PersistentFlags().StringVarP(&configPath,
		"config",
		"c",
		"config.yaml",
		"Specifies path to the config",
	)
	rootCmd.AddCommand(webCmd)
}

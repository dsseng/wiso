package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbAddress string
	db        *gorm.DB
)

var rootCmd = &cobra.Command{
	Use:   "wiso",
	Short: "Wiso is a modern network users manager using FreeRADIUS as the MAC auth backend",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		db, err = gorm.Open(postgres.Open(dbAddress))
		if err != nil {
			panic("failed to connect database")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dbAddress, "database", "d", "", "DSN to connect to DB, e.g. 'host=10.0.0.1 user=a password=b dbname=radius port=5432 sslmode=disable'")
}

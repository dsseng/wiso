package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/users"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "List and manage users",
	Run: func(cmd *cobra.Command, args []string) {
		var results []users.User
		res := db.Model(&users.User{}).Preload("DeviceSessions").Find(&results)
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Full name", "Sessions", "Picture URL")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range results {
			tbl.AddRow(r.ID, r.Username, r.FullName, len(r.DeviceSessions), r.Picture)
		}

		tbl.Print()
	},
}

var userFindCmd = &cobra.Command{
	Use:   "find [username]",
	Short: "Find the user",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []users.User
		res := db.Model(&users.User{}).Preload("DeviceSessions").Find(&entries, "username = ?", args[0])
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Full name", "Sessions", "Picture URL")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range entries {
			tbl.AddRow(r.ID, r.Username, r.FullName, len(r.DeviceSessions), r.Picture)
		}

		tbl.Print()
	},
}

var userDelCmd = &cobra.Command{
	Use:   "del [username]",
	Short: "Delete a user (deleting their sessions)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []users.User
		res := db.Model(&users.User{}).Preload("DeviceSessions").Find(&entries, "username = ?", args[0])
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}
		if res.RowsAffected == 0 {
			fmt.Println("Not found")
			return
		}
		if len(entries[0].DeviceSessions) > 0 {
			res = db.Delete(&entries[0].DeviceSessions)
			if res.Error != nil {
				fmt.Println("A DB error occured", res.Error)
				return
			}
		}

		res = db.Delete(&entries)
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		if res.RowsAffected > 0 {
			fmt.Println("Deleted, rows affected:", res.RowsAffected)
		} else {
			fmt.Println("Not found")
		}
	},
}

func init() {
	userCmd.AddCommand(findCmd)
	userCmd.AddCommand(userDelCmd)
	rootCmd.AddCommand(userCmd)
}

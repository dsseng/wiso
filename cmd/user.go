package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
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

var userSessCmd = &cobra.Command{
	Use:   "sess [username]",
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

		if len(entries) > 0 {
			tbl.Print()

			headerFmt = color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt = color.New(color.FgYellow).SprintfFunc()

			tbl = table.New("Sess ID", "Dev ID", "MAC", "Expiry")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, r := range entries[0].DeviceSessions {
				dev := radius.RadCheck{}
				db.First(&dev, "id = ?", r.RadcheckID)
				tbl.AddRow(r.ID, dev.ID, dev.Username, r.DueDate)
			}

			tbl.Print()
		}
	},
}

var userSessDelCmd = &cobra.Command{
	Use:   "del [username]",
	Short: "Delete all user's sessions",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []users.User
		res := db.Model(&users.User{}).Preload("DeviceSessions").Find(&entries, "username = ?", args[0])
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		if len(entries) > 0 {
			if len(entries[0].DeviceSessions) > 0 {
				res = db.Delete(&entries[0].DeviceSessions)
				if res.Error != nil {
					fmt.Println("A DB error occured", res.Error)
					return
				}
			}
		} else {
			fmt.Println("Not found")
		}
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
	userSessCmd.AddCommand(userSessDelCmd)
	userCmd.AddCommand(userSessCmd)
	userCmd.AddCommand(userDelCmd)
	rootCmd.AddCommand(userCmd)
}

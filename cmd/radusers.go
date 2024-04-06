package cmd

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/spf13/cobra"
)

var macAddressRe = regexp.MustCompile(
	`(?m)^[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}$`,
)

var radusersCmd = &cobra.Command{
	Use:   "radusers",
	Short: "Manage users in the RADIUS DB (typically MAC addresses)",
	Run: func(cmd *cobra.Command, args []string) {
		var results []radius.RadCheck
		res := db.Find(&results)
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Auth")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range results {
			authData := fmt.Sprintf("%s %s %s", r.Attribute, r.Op, r.Value)
			if authData == "Cleartext-Password := macauth" {
				authData = "Authorized by MAC"
			}
			tbl.AddRow(r.ID, r.Username, authData)
		}

		tbl.Print()
	},
}

var findCmd = &cobra.Command{
	Use:   "find [MAC]",
	Short: "Find a MAC address in the RADIUS DB",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []radius.RadCheck
		res := db.Find(&entries, "username = ?", args[0])
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			return
		}

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Auth")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range entries {
			authData := fmt.Sprintf("%s %s %s", r.Attribute, r.Op, r.Value)
			if authData == "Cleartext-Password := macauth" {
				authData = "Authorized by MAC"
			}
			tbl.AddRow(r.ID, r.Username, authData)
		}

		tbl.Print()
	},
}

var addCmd = &cobra.Command{
	Use:   "add [MAC]",
	Short: "Add a MAC address to the RADIUS DB",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !macAddressRe.MatchString(args[0]) {
			fmt.Println("Not a valid MAC address, must be uppercase", args[0])
			return
		}

		var entries []radius.RadCheck
		res := db.Find(&entries, "username = ?", args[0])
		if res.Error != nil {
			fmt.Println("Unable to add row, DB error", res.Error)
			return
		} else if res.RowsAffected > 0 {
			fmt.Println("Unable to add row, already exists")
			return
		}

		entry := radius.RadCheck{
			Username:  args[0],
			Attribute: "Cleartext-Password",
			Op:        ":=",
			Value:     "macauth",
		}
		res = db.Create(&entry)
		if res.Error != nil {
			fmt.Println(res.Error)
			return
		} else {
			fmt.Println("Added successfully, modified rows:", res.RowsAffected)
		}
	},
}

var delCmd = &cobra.Command{
	Use:   "del [MAC]",
	Short: "Delete a MAC address from the RADIUS DB (without disconnecting the client)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var r radius.RadCheck
		res := db.Where("username = ?", args[0]).Delete(&r)
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
	radusersCmd.AddCommand(findCmd)
	radusersCmd.AddCommand(addCmd)
	radusersCmd.AddCommand(delCmd)
	rootCmd.AddCommand(radusersCmd)
}

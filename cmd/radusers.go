package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/spf13/cobra"
)

var radusersCmd = &cobra.Command{
	Use:   "radusers",
	Short: "Manage users in the RADIUS DB (typically MAC addresses)",
	Run: func(cmd *cobra.Command, args []string) {
		var results []radius.RadCheck
		db.Find(&results)

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Auth")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range results {
			tbl.AddRow(r.ID, r.Username, fmt.Sprintf("%s %s %s", r.Attribute, r.Op, r.Value))
		}

		tbl.Print()
	},
	TraverseChildren: true,
}

var addCmd = &cobra.Command{
	Use:   "add [MAC]",
	Short: "Add a MAC address to the RADIUS DB",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
	TraverseChildren: true,
}

func init() {
	radusersCmd.AddCommand(addCmd)
	rootCmd.AddCommand(radusersCmd)
}

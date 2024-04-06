package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(radusersCmd)
}

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

package cmd

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Read authentication logs from the RADIUS database",
	Run: func(cmd *cobra.Command, args []string) {
		var results []radius.RadPostAuth
		db.Find(&results)

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "Reply", "Date/Time")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range results {
			tbl.AddRow(r.ID, r.Username, r.Reply, r.AuthDate)
		}

		tbl.Print()
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}

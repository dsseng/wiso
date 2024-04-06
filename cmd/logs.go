package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/spf13/cobra"
)

var (
	logsSince string
	logLines  int
)

var logsCmd = &cobra.Command{
	Use:   "logs [MAC - optional]",
	Short: "Read authentication logs from the RADIUS database",
	Run: func(cmd *cobra.Command, args []string) {
		var results []radius.RadPostAuth
		duration, err := time.ParseDuration(logsSince)
		if err != nil {
			fmt.Println("Failed to parse log timeout:", logsSince)
			return
		}

		query := db
		if len(args) > 0 {
			query = query.Where("username = ?", args[0])
		}
		query.Order("authdate DESC").Limit(logLines).Find(&results, "authdate >= ?", time.Now().Add(-duration))

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
	logsCmd.PersistentFlags().StringVar(&logsSince, "since", "168h", "Time to show recent logs for. Valid time units are “ns”, “us”, “ms”, “s”, “m”, “h”")
	logsCmd.PersistentFlags().IntVarP(&logLines, "lines", "n", 10, "How many lines to show. -1 means displaying all found")
	rootCmd.AddCommand(logsCmd)
}

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
	sessSince  string
	sessLines  int
	sessActive bool
)

var sessCmd = &cobra.Command{
	Use:   "sess [MAC - optional]",
	Short: "Get active and past sessions from the RADIUS database",
	Run: func(cmd *cobra.Command, args []string) {
		var results []radius.RadAcct
		duration, err := time.ParseDuration(sessSince)
		if err != nil {
			fmt.Println("Failed to parse session timeout:", sessSince)
			return
		}

		query := db
		if sessActive {
			query = query.Where("acctstoptime is null")
		}
		if len(args) > 0 {
			query = query.Where("username = ?", args[0])
		}
		query.Order("acctupdatetime DESC").Limit(logLines).Find(&results, "acctupdatetime >= ?", time.Now().Add(-duration))

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("ID", "Username", "AP", "Start", "Updated", "Stop", "Duration", "IN", "OUT", "IP")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, r := range results {
			tbl.AddRow(r.RadAcctId, r.Username, r.NASIPAddress, r.AcctStartTime, r.AcctUpdateTime, r.AcctStopTime, time.Second*time.Duration(r.AcctSessionTime), r.AcctInputOctets, r.AcctOutputOctets, r.FramedIPAddress)
		}

		tbl.Print()
	},
}

func init() {
	sessCmd.PersistentFlags().StringVar(&sessSince, "since", "168h", "Time to show recent logs for. Valid time units are “ns”, “us”, “ms”, “s”, “m”, “h”")
	sessCmd.PersistentFlags().IntVarP(&sessLines, "lines", "n", 10, "How many lines to show. -1 means displaying all found")
	sessCmd.PersistentFlags().BoolVarP(&sessActive, "active", "a", false, "Whether to only show current sessions")
	rootCmd.AddCommand(sessCmd)
}

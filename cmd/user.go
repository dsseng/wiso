package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"go.withmatt.com/size"

	"github.com/dsseng/wiso/pkg/radius"
	"github.com/dsseng/wiso/pkg/users"
	"github.com/spf13/cobra"
)

var (
	userSessLines  int
	userSessActive bool
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
			sessHeaderFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			sessColumnFmt := color.New(color.FgYellow).SprintfFunc()

			tbl = table.New("Sess ID", "Dev ID", "MAC", "Expiry")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			sessTable := table.New("ID", "Username", "AP", "Start", "Updated",
				"Stop", "Duration", "Uploaded", "Downloaded", "IP")
			sessTable.WithHeaderFormatter(sessHeaderFmt).WithFirstColumnFormatter(sessColumnFmt)

			for _, r := range entries[0].DeviceSessions {
				dev := radius.RadCheck{}
				res = db.First(&dev, "id = ?", r.RadcheckID)
				if res.Error != nil {
					fmt.Println("A DB error occured", res.Error)
					return
				}
				if res.RowsAffected == 0 {
					continue
				}
				tbl.AddRow(r.ID, dev.ID, dev.Username, r.DueDate)

				query := db.Limit(userSessLines).Model(&[]radius.RadAcct{})
				if userSessActive {
					query = query.
						Where("acctupdatetime >= ?", time.Now().Add(-time.Hour)).
						Where("acctstoptime is null")
				}
				acct := []radius.RadAcct{}
				res = db.Find(&acct, "username = ?", dev.Username)
				if res.Error != nil {
					fmt.Println("A DB error occured", res.Error)
					return
				}
				query.Order("acctupdatetime DESC")

				for _, r := range acct {
					sessTable.AddRow(
						r.RadAcctId,
						r.Username,
						r.NASIPAddress,
						r.AcctStartTime,
						r.AcctUpdateTime,
						r.AcctStopTime,
						time.Second*time.Duration(r.AcctSessionTime),
						size.Capacity(r.AcctInputOctets)*size.Byte,
						size.Capacity(r.AcctOutputOctets)*size.Byte,
						r.FramedIPAddress,
					)
				}
			}
			tbl.Print()
			sessTable.Print()

		}
	},
}

var userLogout = &cobra.Command{
	Use:   "logout [username]",
	Short: "Delete all user's sessions disallowing reconnect without auth",
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
	userSessCmd.PersistentFlags().IntVarP(
		&sessLines,
		"lines",
		"n",
		10,
		"How many lines to show. -1 means displaying all found",
	)
	userSessCmd.PersistentFlags().BoolVarP(
		&sessActive,
		"active",
		"a",
		false,
		"Whether to only show current sessions",
	)
	userCmd.AddCommand(userSessCmd)
	userCmd.AddCommand(userLogout)
	userCmd.AddCommand(userDelCmd)
	rootCmd.AddCommand(userCmd)
}

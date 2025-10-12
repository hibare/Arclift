package backup

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

const (
	backupKeyColumnWidthMin = 20
	backupKeyColumnWidthMax = 64
)

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		backups, err := bm.ListBackups(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error listing backups", "error", err)
			return err
		}
		if len(backups) == 0 {
			slog.InfoContext(ctx, "No backups found")
			return nil
		} else {
			fmt.Printf("\nTotal backups %d\n", len(backups)) //nolint:forbidigo // CLI output requires fmt.Printf
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetColumnConfigs([]table.ColumnConfig{
				{
					Name:     "Backup Key",
					WidthMin: backupKeyColumnWidthMin,
					WidthMax: backupKeyColumnWidthMax,
				},
			})
			t.AppendHeader(table.Row{"#", "Backup Key"})

			for i, backup := range backups {

				t.AppendRow([]interface{}{i + 1, backup})
				t.AppendSeparator()
			}

			t.Render()
		}
		return nil
	},
}

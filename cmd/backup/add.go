package backup

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Perform a backup",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := bm.Backup(ctx); err != nil {
			slog.ErrorContext(ctx, "error backing up", "error", err)
			return err
		}
		return nil
	},
}

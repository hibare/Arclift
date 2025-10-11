package backup

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// purgeCmd represents the purge command.
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge old backups",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := bm.PurgeOldBackups(ctx); err != nil {
			slog.ErrorContext(ctx, "error purging old backups", "error", err)
			return err
		}
		return nil
	},
}

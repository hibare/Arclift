package backup

import (
	"github.com/hibare/GoS3Backup/cmd/common"
	"github.com/hibare/GoS3Backup/internal/backup"
	"github.com/spf13/cobra"
)

var bm backup.BackupManagerIface

// BackupCmd represents the backup command.
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Perform backups & related operations",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		configPath := cmd.Root().PersistentFlags().Lookup("config").Value.String()
		bm, err = common.NewBackupManager(cmd.Context(), configPath)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return bm.Backup(cmd.Context())
	},
}

func init() {
	BackupCmd.AddCommand(addCmd)
	BackupCmd.AddCommand(purgeCmd)
	BackupCmd.AddCommand(listCmd)
}

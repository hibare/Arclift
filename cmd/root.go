package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	cmdBackup "github.com/hibare/arclift/cmd/backup"
	"github.com/hibare/arclift/cmd/common"
	cmdConfig "github.com/hibare/arclift/cmd/config"
	"github.com/hibare/arclift/internal/config"
	"github.com/hibare/arclift/internal/constants"
	"github.com/hibare/arclift/internal/version"
	"github.com/spf13/cobra"
)

var (
	ConfigPath string
)

var RootCmd = &cobra.Command{
	Use:     "arclift",
	Short:   "Application to backup directories to S3",
	Long:    "",
	Version: version.CurrentVersion,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		s := gocron.NewScheduler(time.UTC)

		bm, err := common.NewBackupManager(ctx, ConfigPath)
		if err != nil {
			return err
		}

		// Schedule backup job
		if _, bcErr := s.Cron(config.Current.Backup.Cron).Do(func() {
			if baErr := bm.Backup(ctx); baErr != nil {
				slog.ErrorContext(ctx, "Error backing up", "error", baErr)
			}
			if bpErr := bm.PurgeOldBackups(ctx); bpErr != nil {
				slog.ErrorContext(ctx, "Error purging old backups", "error", bpErr)
			}
		}); bcErr != nil {
			slog.ErrorContext(ctx, "Error setting up cron", "error", bcErr)
			return bcErr
		}
		slog.InfoContext(ctx, "Scheduled backup job", "cron", config.Current.Backup.Cron)

		// Schedule version check job
		if _, vcErr := s.Cron(constants.VersionCheckCron).Do(func() {
			if vErr := version.V.CheckUpdate(); vErr != nil {
				slog.ErrorContext(ctx, "Error checking for updates", "error", vErr)
			}
		}); vcErr != nil {
			slog.WarnContext(ctx, "Failed to schedule version check job", "error", vcErr)
		}

		s.StartBlocking()
		return nil
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Add global flags
	RootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "", "Path to config file")

	// Add commands
	RootCmd.AddCommand(cmdConfig.ConfigCmd)
	RootCmd.AddCommand(cmdBackup.BackupCmd)

	// Perform initial version check
	go func() {
		_ = version.V.CheckUpdate()
		if version.V.IsUpdateAvailable() {
			slog.Info(version.V.GetUpdateNotification())
		}
	}()
}

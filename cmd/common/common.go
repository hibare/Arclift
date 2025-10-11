package common

import (
	"context"

	"github.com/hibare/GoS3Backup/internal/backup"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/notifiers"
	"github.com/hibare/GoS3Backup/internal/storage/s3"
)

func NewBackupManager(ctx context.Context, configPath string) (backup.BackupManagerIface, error) {
	cfg, err := config.GetConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}

	store := s3.NewS3Storage(cfg)
	if err := store.Init(ctx); err != nil {
		return nil, err
	}

	notifierStore := notifiers.NewNotifier(cfg)
	if err := notifierStore.InitStore(); err != nil {
		return nil, err
	}

	return backup.NewBackupManager(cfg, store, notifierStore), nil
}

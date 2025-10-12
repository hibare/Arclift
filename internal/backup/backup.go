package backup

import (
	"context"
	"errors"
	"log/slog"
	"os"

	commonGPG "github.com/hibare/GoCommon/v2/pkg/crypto/gpg"
	"github.com/hibare/GoCommon/v2/pkg/datetime"
	commonFiles "github.com/hibare/GoCommon/v2/pkg/file"
	"github.com/hibare/arclift/internal/config"
	"github.com/hibare/arclift/internal/notifiers"
	"github.com/hibare/arclift/internal/storage"
)

var (
	// ErrNoProcessableFiles is returned when no processable files are found.
	ErrNoProcessableFiles = errors.New("no processable files")
)

// BackupManagerIface defines the interface for the backup manager.
type BackupManagerIface interface {
	Backup(ctx context.Context) error
	PurgeOldBackups(ctx context.Context) error
	ListBackups(ctx context.Context) ([]string, error)
}

// BackupManager implements the BackupManagerIface.
type BackupManager struct {
	cfg           *config.Config
	store         storage.StorageIface
	gpg           commonGPG.GPGIface
	notifierStore notifiers.NotifierStoreIface
}

func (b *BackupManager) unArchivedBackup(ctx context.Context, dir string) (storage.UploadDirResponse, error) {
	slog.InfoContext(ctx, "uploading directory", "dir", dir)
	resp, err := b.store.UploadDir(ctx, dir)
	if err != nil {
		slog.ErrorContext(ctx, "Error uploading directory", "dir", dir, "error", err)
		return storage.UploadDirResponse{}, err
	}
	return resp, nil
}

func (b *BackupManager) archivedBackup(ctx context.Context, dir string) (storage.UploadDirResponse, error) {
	var uploadPath string

	slog.InfoContext(ctx, "Archiving dir", "dir", dir)

	archiveResp, err := commonFiles.ArchiveDir(dir, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error archiving dir", "dir", dir, "error", err)
		return storage.UploadDirResponse{}, err
	}

	if archiveResp.SuccessFiles <= 0 {
		slog.ErrorContext(ctx, "No processable files", "dir", dir)
		return storage.UploadDirResponse{}, ErrNoProcessableFiles
	}

	uploadPath = archiveResp.ArchivePath

	slog.InfoContext(ctx, "Archived dir", "dir", dir, "archiveResp", archiveResp)

	if b.cfg.Backup.Encryption.Enabled {
		slog.InfoContext(ctx, "Fetching GPG key")
		if _, gErr := b.gpg.FetchGPGPubKeyFromKeyServer(b.cfg.Backup.Encryption.GPG.KeyID, b.cfg.Backup.Encryption.GPG.KeyServer); gErr != nil {
			slog.ErrorContext(ctx, "Error fetching GPG key", "error", gErr)
			return storage.UploadDirResponse{}, gErr
		}

		slog.InfoContext(ctx, "Encrypting archive")
		encryptedFilePath, eErr := b.gpg.EncryptFile(archiveResp.ArchivePath)
		if eErr != nil {
			slog.ErrorContext(ctx, "Error encrypting archive", "error", eErr)
			return storage.UploadDirResponse{}, eErr
		}

		uploadPath = encryptedFilePath
		slog.InfoContext(ctx, "Encrypted archive", "uploadPath", uploadPath)
		_ = os.Remove(archiveResp.ArchivePath)
	}

	slog.InfoContext(ctx, "uploading file", "uploadPath", uploadPath, "storage", b.store.Name())
	resp, err := b.store.UploadFile(ctx, uploadPath)
	if err != nil {
		slog.ErrorContext(ctx, "Error uploading file", "error", err)
		return storage.UploadDirResponse{}, err
	}

	slog.InfoContext(ctx, "Uploaded file", "uploadPath", uploadPath)
	_ = os.Remove(uploadPath)
	return storage.UploadDirResponse{
		BaseKey:      resp,
		TotalFiles:   archiveResp.TotalFiles,
		TotalDirs:    archiveResp.TotalDirs,
		SuccessFiles: archiveResp.SuccessFiles,
		FailedFiles:  archiveResp.FailedFiles,
	}, nil
}

// Backup performs a backup & sends notifications.
func (b *BackupManager) Backup(ctx context.Context) error {
	for _, dir := range b.cfg.Backup.Dirs {
		slog.InfoContext(ctx, "Processing path", "path", dir)

		if b.cfg.Backup.ArchiveDirs {
			backupResp, err := b.archivedBackup(ctx, dir)
			if err != nil {
				slog.ErrorContext(ctx, "Error backing up dir", "dir", dir, "error", err)
				b.notifierStore.NotifyBackupFailure(ctx, dir, backupResp.TotalDirs, backupResp.TotalFiles, err)
				continue
			}

			slog.InfoContext(ctx, "Backed up dir", "dir", dir, "backupResp", backupResp)
			b.notifierStore.NotifyBackupSuccess(ctx, dir, backupResp.TotalDirs, backupResp.TotalFiles, backupResp.SuccessFiles, backupResp.BaseKey)
			continue
		}

		backupResp, err := b.unArchivedBackup(ctx, dir)
		if err != nil {
			slog.ErrorContext(ctx, "Error backing up dir", "dir", dir, "error", err)
			b.notifierStore.NotifyBackupFailure(ctx, dir, backupResp.TotalDirs, backupResp.TotalFiles, err)
			continue
		}

		slog.InfoContext(ctx, "Backed up dir", "dir", dir, "backupResp", backupResp)
		b.notifierStore.NotifyBackupSuccess(ctx, dir, backupResp.TotalDirs, backupResp.TotalFiles, backupResp.SuccessFiles, backupResp.BaseKey)
		continue
	}
	return nil
}

// ListBackups lists the backups.
func (b *BackupManager) ListBackups(ctx context.Context) ([]string, error) {
	keys, err := b.store.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error listing backups", "error", err)
		return nil, err
	}

	if len(keys) == 0 {
		slog.InfoContext(ctx, "No backups found")
		return []string{}, nil
	}

	keys = b.store.TrimPrefix(keys)
	keys = datetime.SortDateTimes(keys)
	slog.DebugContext(ctx, "Found backups", "keys", keys)
	return keys, nil
}

// PurgeOldBackups purges old backups.
func (b *BackupManager) PurgeOldBackups(ctx context.Context) error {
	keys, err := b.ListBackups(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error listing backups", "error", err)
		return err
	}

	if len(keys) <= b.cfg.Backup.RetentionCount {
		slog.InfoContext(ctx, "No backups to purge")
		return nil
	}

	keysToDelete := keys[b.cfg.Backup.RetentionCount:]
	slog.InfoContext(ctx, "Found backups to delete", "keys", keysToDelete, "retention", b.cfg.Backup.RetentionCount)

	for _, key := range keysToDelete {
		slog.InfoContext(ctx, "Deleting backup", "key", key)
		err := b.store.Delete(ctx, key)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting backup", "key", key, "error", err)
			b.notifierStore.NotifyBackupDeleteFailure(ctx, key, err)
			continue
		}
	}

	slog.InfoContext(ctx, "Deletion completed successfully")
	return nil
}

func newBackupManager(cfg *config.Config, store storage.StorageIface, notifierStore notifiers.NotifierStoreIface) *BackupManager {
	return &BackupManager{
		cfg:           cfg,
		store:         store,
		gpg:           commonGPG.NewGPG(commonGPG.Options{}),
		notifierStore: notifierStore,
	}
}

// NewBackupManager creates a new backup manager.
var NewBackupManager = newBackupManager

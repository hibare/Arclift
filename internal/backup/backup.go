package backup

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/notifiers"
	"github.com/hibare/GoS3Backup/internal/s3"
	"github.com/hibare/GoS3Backup/internal/utils"
)

func Backup() {
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetTimeStampedPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)
	log.Infof("prefix: %s", prefix)

	// Loop through individual backup dir & perform backup
	for _, dir := range config.Current.Backup.Dirs {
		log.Infof("Processing path %s", dir)

		// Estimate number of files & dirs in backup directory
		files, dirs := utils.EstimateFilesAndDirs(config.Current.Backup.Dirs[0])

		log.Infof("Estimated files: %d, dirs: %d", files, dirs)

		if err := s3.Upload(sess, config.Current.S3.Bucket, prefix, dir); err != nil {
			log.Errorf("Error uploading files: %v", err)
			notifiers.BackupFailedNotification(err.Error(), dir, dirs, files)
			continue
		}

		notifiers.BackupSuccessfulNotification(dir, dirs, files, prefix)
	}

}

func ListBackups() []string {
	var keys []string
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)
	log.Infof("prefix: %s", prefix)

	// Retrieve objects by prefix
	keys, err := s3.ListObjectsAtPrefixRoot(sess, config.Current.S3.Bucket, prefix)
	if err != nil {
		log.Errorf("Error listing objects: %v", err)
		notifiers.BackupDeletionFailureNotification(err.Error(), constants.NotAvailable)
		return keys
	}

	if len(keys) == 0 {
		log.Info("No backups found")
		return keys
	}

	log.Infof("Found %d backups", len(keys))

	// Remove prefix from key to get datetime string
	keys = utils.TrimPrefix(keys, prefix)

	// Sort datetime strings by descending order
	sortedKeys := utils.SortDateTimes(keys)

	return sortedKeys
}

func PurgeOldBackups() {
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)

	backups := ListBackups()

	if len(backups) <= int(config.Current.Backup.RetentionCount) {
		log.Info("No backups to delete")
		return
	}

	keysToDelete := backups[config.Current.Backup.RetentionCount:]
	log.Infof("Found %d backups to delete (backup rentention %d) [%s]", len(keysToDelete), config.Current.Backup.RetentionCount, keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		log.Infof("Deleting backup %s", key)
		key = filepath.Join(prefix, key)

		if err := s3.DeleteObjects(sess, config.Current.S3.Bucket, key, true); err != nil {
			log.Errorf("Error deleting backup %s: %v", key, err)
			notifiers.BackupDeletionFailureNotification(err.Error(), key)
			continue
		}
	}

	log.Info("Deletion completed successfully")
}

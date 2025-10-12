// Package storage defines the interface for various storage backends.
package storage

import "context"

type UploadDirResponse struct {
	BaseKey      string
	TotalFiles   int
	TotalDirs    int
	SuccessFiles int
	FailedFiles  map[string]error
}

// StorageIface defines a generic storage backend used to upload and manage backups.
// revive:disable-next-line exported
type StorageIface interface {
	// Init prepares the storage (e.g., establishes session)
	Init(context.Context) error

	// UploadFile uploads a local file and returns the remote key/path
	UploadFile(context.Context, string) (string, error)

	// UploadDir uploads a local directory and returns the remote key/path
	UploadDir(context.Context, string) (UploadDirResponse, error)

	// List returns keys/identifiers under configured prefix
	List(context.Context) ([]string, error)

	// Delete deletes the provided key/path from storage
	Delete(context.Context, string) error

	// TrimPrefix trims the configured prefix from a given key, if present
	TrimPrefix(keys []string) []string

	// Name returns the name of the storage backend (e.g., "s3", "gcs")
	Name() string
}

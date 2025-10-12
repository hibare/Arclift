// Package notifiers implements various notification mechanisms for backup events.
package notifiers

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/hibare/arclift/internal/config"
	"github.com/hibare/arclift/internal/notifiers/discord"
)

var (
	// ErrNotifiersDisabled is returned when notifiers are globally disabled.
	ErrNotifiersDisabled = errors.New("notifiers are disabled")

	// ErrNotifierDisabled is returned when a specific notifier is disabled.
	ErrNotifierDisabled = errors.New("notifier is disabled")
)

// NotifiersIface defines the interface that all notifier implementations must satisfy.
// revive:disable-next-line exported
type NotifiersIface interface {
	Enabled() bool
	NotifyBackupSuccess(ctx context.Context, directory string, totalDirs, totalFiles, successFiles int, key string) error
	NotifyBackupFailure(ctx context.Context, directory string, totalDirs, totalFiles int, err error) error
	NotifyBackupDeleteFailure(ctx context.Context, key string, err error) error
}

// NotifierStoreIface defines the interface for managing multiple notifiers.
type NotifierStoreIface interface {
	Enabled() bool
	NotifyBackupSuccess(ctx context.Context, directory string, totalDirs, totalFiles, successFiles int, key string)
	NotifyBackupFailure(ctx context.Context, directory string, totalDirs, totalFiles int, err error)
	NotifyBackupDeleteFailure(ctx context.Context, key string, err error)
	InitStore() error
}

// Notifier manages multiple notifier implementations.
type Notifier struct {
	cfg   *config.Config
	mu    sync.RWMutex
	store []NotifiersIface
}

func (n *Notifier) register(nf NotifiersIface) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.store = append(n.store, nf)
}

// Enabled checks if notifiers are globally enabled in the configuration.
func (n *Notifier) Enabled() bool {
	return n.cfg.Notifiers.Enabled
}

// NotifyBackupSuccess sends a backup success notification using all enabled notifiers.
func (n *Notifier) NotifyBackupSuccess(ctx context.Context, directory string, totalDirs, totalFiles, successFiles int, key string) {
	if !n.Enabled() {
		slog.ErrorContext(ctx, "Notifiers are disabled; skipping NotifyBackupSuccess")
	}

	for _, notifier := range n.store {
		if !notifier.Enabled() {
			slog.DebugContext(ctx, "Notifier disabled; skipping NotifyBackupSuccess")
			continue
		}
		if err := notifier.NotifyBackupSuccess(ctx, directory, totalDirs, totalFiles, successFiles, key); err != nil {
			slog.ErrorContext(ctx, "Failed to send NotifyBackupSuccess", "error", err)
		}
	}
}

// NotifyBackupFailure sends a backup failure notification using all enabled notifiers.
func (n *Notifier) NotifyBackupFailure(ctx context.Context, directory string, totalDirs, totalFiles int, bErr error) {
	if !n.Enabled() {
		slog.ErrorContext(ctx, "Notifiers are disabled; skipping NotifyBackupFailure")
	}

	for _, notifier := range n.store {
		if !notifier.Enabled() {
			slog.DebugContext(ctx, "Notifier disabled; skipping NotifyBackupFailure")
			continue
		}
		if err := notifier.NotifyBackupFailure(ctx, directory, totalDirs, totalFiles, bErr); err != nil {
			slog.ErrorContext(ctx, "Failed to send NotifyBackupFailure", "error", err)
		}
	}
}

// NotifyBackupDeleteFailure sends a backup deletion failure notification using all enabled notifiers.
func (n *Notifier) NotifyBackupDeleteFailure(ctx context.Context, key string, bErr error) {
	if !n.Enabled() {
		slog.ErrorContext(ctx, "Notifiers are disabled; skipping NotifyBackupDeleteFailure")
	}

	for _, notifier := range n.store {
		if !notifier.Enabled() {
			slog.DebugContext(ctx, "Notifier disabled; skipping NotifyBackupDeleteFailure")
			continue
		}
		if err := notifier.NotifyBackupDeleteFailure(ctx, key, bErr); err != nil {
			slog.ErrorContext(ctx, "Failed to send NotifyBackupDeleteFailure", "error", err)
		}
	}
}

// InitStore initializes and registers all available notifiers.
func (n *Notifier) InitStore() error {
	if n.cfg.Notifiers.Discord.Enabled {
		d, err := discord.NewDiscordNotifier(n.cfg)
		if err != nil {
			return err
		}

		n.register(d)
	}
	return nil
}

// NewNotifier creates a new Notifier instance with the provided configuration.
func NewNotifier(cfg *config.Config) NotifierStoreIface {
	return &Notifier{cfg: cfg}
}

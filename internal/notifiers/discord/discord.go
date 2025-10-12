// Package discord provides implementations for various notification services.
package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/hibare/GoCommon/v2/pkg/notifiers/discord"
	"github.com/hibare/arclift/internal/config"
	"github.com/hibare/arclift/internal/constants"
	"github.com/hibare/arclift/internal/version"
)

const (
	successColor         = 1498748
	failureColor         = 14554702
	deletionFailureColor = 14590998
)

// Discord sends notifications to a Discord channel via webhook.
type Discord struct {
	Cfg    *config.Config
	client discord.ClientIface
}

// Enabled checks if the Discord notifier is enabled in the configuration.
func (d *Discord) Enabled() bool {
	return d.Cfg.Notifiers.Discord.Enabled
}

// NotifyBackupSuccess sends a success notification to the Discord channel.
func (d *Discord) NotifyBackupSuccess(ctx context.Context, directory string, totalDirs, totalFiles, successFiles int, key string) error {
	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Directory",
				Description: directory,
				Color:       successColor,
				Fields: []discord.EmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
					{
						Name:   "Dirs",
						Value:  strconv.Itoa(totalDirs),
						Inline: true,
					},
					{
						Name:   "Files",
						Value:  fmt.Sprintf("%d/%d", successFiles, totalFiles),
						Inline: true,
					},
				},
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramPrettyIdentifier,
		Content:    fmt.Sprintf("**Backup Successful** - *%s*", d.Cfg.Backup.Hostname),
	}

	if version.V.IsUpdateAvailable() {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			slog.Error("error adding footer to message", "error", err)
		}
	}

	return d.client.Send(ctx, &message)
}

// NotifyBackupFailure sends a failure notification to the Discord channel.
func (d *Discord) NotifyBackupFailure(ctx context.Context, directory string, totalDirs, totalFiles int, err error) error {
	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       failureColor,
				Fields: []discord.EmbedField{
					{
						Name:   "Directory",
						Value:  directory,
						Inline: false,
					},
					{
						Name:   "Dirs",
						Value:  strconv.Itoa(totalDirs),
						Inline: true,
					},
					{
						Name:   "Files",
						Value:  strconv.Itoa(totalFiles),
						Inline: true,
					},
				},
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramPrettyIdentifier,
		Content:    fmt.Sprintf("**Backup Failed** - *%s*", d.Cfg.Backup.Hostname),
	}

	if version.V.IsUpdateAvailable() {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			slog.Error("error adding footer to message", "error", err)
		}
	}

	return d.client.Send(ctx, &message)
}

// NotifyBackupDeleteFailure sends a deletion failure notification to the Discord channel.
func (d *Discord) NotifyBackupDeleteFailure(ctx context.Context, key string, err error) error {
	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       deletionFailureColor,
				Fields: []discord.EmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
				},
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramPrettyIdentifier,
		Content:    fmt.Sprintf("**Backup Deletion Failed** - *%s*", d.Cfg.Backup.Hostname),
	}

	if version.V.IsUpdateAvailable() {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			slog.Error("error adding footer to message", "error", err)
		}
	}

	return d.client.Send(ctx, &message)
}

// NewDiscordNotifier creates a new Discord notifier instance.
func NewDiscordNotifier(cfg *config.Config) (*Discord, error) {
	client, err := discord.NewClient(discord.Options{
		WebhookURL: cfg.Notifiers.Discord.Webhook,
	})
	if err != nil {
		return nil, err
	}

	return &Discord{
		Cfg:    cfg,
		client: client,
	}, nil
}

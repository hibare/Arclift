package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	commonRuntime "github.com/hibare/GoCommon/v2/pkg/os/runtime"
	commonUtils "github.com/hibare/GoCommon/v2/pkg/utils"
	"github.com/hibare/arclift/internal/constants"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// S3Config is the configuration for the S3 client.
type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"   yaml:"endpoint"`
	Region    string `mapstructure:"region"     yaml:"region"`
	AccessKey string `mapstructure:"access-key" yaml:"access-key"`
	SecretKey string `mapstructure:"secret-key" yaml:"secret-key"`
	Bucket    string `mapstructure:"bucket"     yaml:"bucket"`
	Prefix    string `mapstructure:"prefix"     yaml:"prefix"`
}

// GPGConfig is the configuration for the GPG client.
type GPGConfig struct {
	KeyServer string `mapstructure:"key-server" yaml:"key-server"`
	KeyID     string `mapstructure:"key-id"     yaml:"key-id"`
}

// Encryption is the configuration for the encryption.
type Encryption struct {
	Enabled bool      `mapstructure:"enabled" yaml:"enabled"`
	GPG     GPGConfig `mapstructure:"gpg"     yaml:"gpg"`
}

// BackupConfig is the configuration for the backup.
type BackupConfig struct {
	Dirs           []string   `mapstructure:"dirs"             yaml:"dirs"`
	Hostname       string     `mapstructure:"hostname"         yaml:"hostname"`
	RetentionCount int        `mapstructure:"retention-count"  yaml:"retention-count"`
	DateTimeLayout string     `mapstructure:"date-time-layout" yaml:"date-time-layout"`
	Cron           string     `mapstructure:"cron"             yaml:"cron"`
	ArchiveDirs    bool       `mapstructure:"archive-dirs"     yaml:"archive-dirs"`
	Encryption     Encryption `mapstructure:"encryption"       yaml:"encryption"`
}

func (b *BackupConfig) validate() error {
	if len(b.Dirs) == 0 {
		return errors.New("dirs is required")
	}

	if b.RetentionCount <= 0 {
		return errors.New("retention-count must be greater than 0")
	}

	if b.Cron == "" {
		return errors.New("cron is required")
	}

	// ToDo: Add cron validation

	// Check if encryption is enabled & encryption config is enabled.
	if b.Encryption.Enabled && !b.ArchiveDirs {
		slog.Warn("Backup encryption is only available when archive dirs are enabled. Disabling encryption")
		b.Encryption.Enabled = false
	} else if b.Encryption.Enabled {
		if b.Encryption.GPG.KeyServer == "" || b.Encryption.GPG.KeyID == "" {
			slog.Error("Encryption is enabled but GPG key server or key ID is missing")
			b.Encryption.Enabled = false
		}
	}

	return nil
}

// DiscordNotifierConfig is the configuration for the Discord notifier.
type DiscordNotifierConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Webhook string `mapstructure:"webhook" yaml:"webhook"`
}

func (d *DiscordNotifierConfig) validate() error {
	if d.Enabled && d.Webhook == "" {
		slog.Warn("Discord notifier is enabled but webhook is not set. Disabling Discord notifier")
		d.Enabled = false
	}
	return nil
}

// NotifiersConfig is the configuration for the notifiers.
type NotifiersConfig struct {
	Enabled bool                  `mapstructure:"enabled" yaml:"enabled"`
	Discord DiscordNotifierConfig `mapstructure:"discord" yaml:"discord"`
}

func (n *NotifiersConfig) validate() error {
	if err := n.Discord.validate(); err != nil {
		return err
	}
	return nil
}

// LoggerConfig is the configuration for the logger.
type LoggerConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
	Mode  string `mapstructure:"mode"  yaml:"mode"`
}

func (l *LoggerConfig) validate() error {
	if !commonLogger.IsValidLogLevel(l.Level) {
		return fmt.Errorf("invalid logger level: %s", l.Level)
	}

	if !commonLogger.IsValidLogMode(l.Mode) {
		return fmt.Errorf("invalid logger mode: %s", l.Mode)
	}

	return nil
}

// Config is the configuration for the program.
type Config struct {
	S3        S3Config        `mapstructure:"s3"        yaml:"s3"`
	Backup    BackupConfig    `mapstructure:"backup"    yaml:"backup"`
	Notifiers NotifiersConfig `mapstructure:"notifiers" yaml:"notifiers"`
	Logger    LoggerConfig    `mapstructure:"logger"    yaml:"logger"`
}

func (c *Config) validate() error {
	validators := []func() error{
		c.Logger.validate,
		c.Backup.validate,
		c.Notifiers.validate,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) getViper(ctx context.Context, path string) *viper.Viper {
	v := viper.New()
	v.SetConfigName(commonRuntime.ConfigFileName)
	v.SetConfigType(commonRuntime.ConfigFileExtension)

	runtime := commonRuntime.New()

	// Config search paths.
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(runtime.GetConfigDir(), constants.ProgramIdentifier))
	}

	// Environment variable binding.
	v.SetEnvPrefix("ARCLIFT")
	v.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))
	v.AutomaticEnv()

	envBindings := map[string]string{
		"s3.endpoint":                      "s3.endpoint",
		"s3.region":                        "s3.region",
		"s3.access-key":                    "s3.access-key",
		"s3.secret-key":                    "s3.secret-key",
		"s3.bucket":                        "s3.bucket",
		"s3.prefix":                        "s3.prefix",
		"backup.retention-count":           "backup.retention-count",
		"backup.date-time-layout":          "backup.date-time-layout",
		"backup.cron":                      "backup.cron",
		"backup.archive-dirs":              "backup.archive-dirs",
		"Backup.Encryption.Enabled":        "backup.encryption.enabled",
		"backup.encryption.gpg.key-server": "backup.encryption.gpg.key-server",
		"backup.encryption.gpg.key-id":     "backup.encryption.gpg.key-id",
		"notifiers.discord.enabled":        "notifiers.discord.enabled",
		"notifiers.discord.webhook":        "notifiers.discord.webhook",
		"logger.level":                     "logger.level",
		"logger.mode":                      "logger.mode",
	}

	for configKey, envVar := range envBindings {
		if err := v.BindEnv(configKey, envVar); err != nil {
			slog.WarnContext(ctx, "Failed to bind environment variable",
				slog.String("config", configKey),
				slog.String("env", envVar),
				slog.String("error", err.Error()))
		}
	}

	// Add default values.
	v.SetDefault("s3.endpoint", "")
	v.SetDefault("s3.region", "")
	v.SetDefault("s3.access-key", "")
	v.SetDefault("s3.secret-key", "")
	v.SetDefault("s3.bucket", "")
	v.SetDefault("s3.prefix", "")
	v.SetDefault("backup.dirs", []string{})
	v.SetDefault("backup.retention-count", constants.DefaultRetentionCount)
	v.SetDefault("backup.date-time-layout", constants.DefaultDateTimeLayout)
	v.SetDefault("backup.cron", constants.DefaultCron)
	v.SetDefault("backup.hostname", commonUtils.GetHostname())
	v.SetDefault("backup.archive-dirs", false)
	v.SetDefault("backup.encryption.enabled", false)
	v.SetDefault("backup.encryption.gpg.key-server", "")
	v.SetDefault("backup.encryption.gpg.key-id", "")
	v.SetDefault("notifiers.enabled", false)
	v.SetDefault("notifiers.discord.enabled", false)
	v.SetDefault("notifiers.discord.webhook", "")
	v.SetDefault("logger.level", commonLogger.DefaultLoggerLevel)
	v.SetDefault("logger.mode", commonLogger.DefaultLoggerMode)

	return v
}

// LoadConfig loads the configuration from the config file.
func LoadConfig(ctx context.Context, configPath string) (*Config, error) {
	cfg := &Config{}
	v := cfg.getViper(ctx, configPath)

	// Try read config.
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
			slog.WarnContext(ctx, "No config file found, relying on env vars/defaults")
		} else {
			return nil, err
		}
	} else {
		slog.InfoContext(ctx, "Using config file", slog.String("file", v.ConfigFileUsed()))
	}

	// Unmarshal into Current.
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Initialize logger.
	commonLogger.InitLogger(&cfg.Logger.Level, &cfg.Logger.Mode)

	return cfg, nil
}

// Current is the current configuration.
var Current *Config

// GetConfig gets the current configuration.
func GetConfig(ctx context.Context, configPath string) (*Config, error) {
	if Current == nil {
		var err error
		Current, err = LoadConfig(ctx, configPath)
		if err != nil {
			return nil, err
		}
	}
	return Current, nil
}

// GenerateConfigFile generates a new config file.
func GenerateConfigFile(ctx context.Context, configPath string) (string, error) {
	cfg := &Config{}
	v := cfg.getViper(ctx, configPath)

	// Unmarshal viper's defaults into the config struct
	if err := v.Unmarshal(cfg); err != nil {
		return "", fmt.Errorf("failed to unmarshal defaults: %w", err)
	}

	// Marshal the config struct to YAML
	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Create a reader from the YAML bytes
	reader := bytes.NewReader(yamlBytes)

	// Read the YAML into viper
	if err := v.ReadConfig(reader); err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	return v.ConfigFileUsed(), v.WriteConfig()
}

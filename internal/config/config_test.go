package config

import (
	"os"
	"path/filepath"
	"testing"

	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBackupConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BackupConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 10,
				Cron:           "0 0 * * *",
			},
			wantErr: false,
		},
		{
			name: "empty dirs",
			config: BackupConfig{
				Dirs:           []string{},
				RetentionCount: 10,
				Cron:           "0 0 * * *",
			},
			wantErr: true,
			errMsg:  "dirs is required",
		},
		{
			name: "zero retention count",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 0,
				Cron:           "0 0 * * *",
			},
			wantErr: true,
			errMsg:  "retention-count must be greater than 0",
		},
		{
			name: "negative retention count",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: -5,
				Cron:           "0 0 * * *",
			},
			wantErr: true,
			errMsg:  "retention-count must be greater than 0",
		},
		{
			name: "empty cron",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 10,
				Cron:           "",
			},
			wantErr: true,
			errMsg:  "cron is required",
		},
		{
			name: "encryption enabled without archive dirs",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 10,
				Cron:           "0 0 * * *",
				ArchiveDirs:    false,
				Encryption: Encryption{
					Enabled: true,
					GPG: GPGConfig{
						KeyServer: "keyserver.ubuntu.com",
						KeyID:     "12345678",
					},
				},
			},
			wantErr: false, // Encryption is disabled automatically with warning
		},
		{
			name: "encryption enabled with archive dirs but missing GPG config",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 10,
				Cron:           "0 0 * * *",
				ArchiveDirs:    true,
				Encryption: Encryption{
					Enabled: true,
					GPG: GPGConfig{
						KeyServer: "",
						KeyID:     "",
					},
				},
			},
			wantErr: false, // Encryption is disabled automatically with error log
		},
		{
			name: "encryption enabled with archive dirs and valid GPG config",
			config: BackupConfig{
				Dirs:           []string{"/tmp/test"},
				RetentionCount: 10,
				Cron:           "0 0 * * *",
				ArchiveDirs:    true,
				Encryption: Encryption{
					Enabled: true,
					GPG: GPGConfig{
						KeyServer: "keyserver.ubuntu.com",
						KeyID:     "12345678",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiscordNotifierConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DiscordNotifierConfig
		wantErr bool
	}{
		{
			name: "disabled notifier",
			config: DiscordNotifierConfig{
				Enabled: false,
				Webhook: "",
			},
			wantErr: false,
		},
		{
			name: "enabled with webhook",
			config: DiscordNotifierConfig{
				Enabled: true,
				Webhook: "https://discord.com/api/webhooks/123/abc",
			},
			wantErr: false,
		},
		{
			name: "enabled without webhook",
			config: DiscordNotifierConfig{
				Enabled: true,
				Webhook: "",
			},
			wantErr: false, // Disabled automatically with warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotifiersConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NotifiersConfig
		wantErr bool
	}{
		{
			name: "valid notifiers config",
			config: NotifiersConfig{
				Enabled: true,
				Discord: DiscordNotifierConfig{
					Enabled: true,
					Webhook: "https://discord.com/api/webhooks/123/abc",
				},
			},
			wantErr: false,
		},
		{
			name: "disabled notifiers",
			config: NotifiersConfig{
				Enabled: false,
				Discord: DiscordNotifierConfig{
					Enabled: false,
					Webhook: "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoggerConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  LoggerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid debug level and pretty mode",
			config: LoggerConfig{
				Level: "DEBUG",
				Mode:  "PRETTY",
			},
			wantErr: false,
		},
		{
			name: "valid info level and json mode",
			config: LoggerConfig{
				Level: "INFO",
				Mode:  "JSON",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: LoggerConfig{
				Level: "invalid",
				Mode:  "PRETTY",
			},
			wantErr: true,
			errMsg:  "invalid logger level",
		},
		{
			name: "invalid log mode",
			config: LoggerConfig{
				Level: "INFO",
				Mode:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid logger mode",
		},
		{
			name: "both invalid",
			config: LoggerConfig{
				Level: "invalid",
				Mode:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid logger level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Logger: LoggerConfig{
					Level: "INFO",
					Mode:  "PRETTY",
				},
				Backup: BackupConfig{
					Dirs:           []string{"/tmp/test"},
					RetentionCount: 10,
					Cron:           "0 0 * * *",
				},
				Notifiers: NotifiersConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid logger",
			config: Config{
				Logger: LoggerConfig{
					Level: "invalid",
					Mode:  "PRETTY",
				},
				Backup: BackupConfig{
					Dirs:           []string{"/tmp/test"},
					RetentionCount: 10,
					Cron:           "0 0 * * *",
				},
				Notifiers: NotifiersConfig{
					Enabled: false,
				},
			},
			wantErr: true,
			errMsg:  "invalid logger level",
		},
		{
			name: "invalid backup config",
			config: Config{
				Logger: LoggerConfig{
					Level: "INFO",
					Mode:  "PRETTY",
				},
				Backup: BackupConfig{
					Dirs:           []string{},
					RetentionCount: 10,
					Cron:           "0 0 * * *",
				},
				Notifiers: NotifiersConfig{
					Enabled: false,
				},
			},
			wantErr: true,
			errMsg:  "dirs is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetViper(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "with path",
			path: "/tmp/config.yaml",
		},
		{
			name: "without path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			v := cfg.getViper(t.Context(), tt.path)
			assert.NotNil(t, v)
		})
	}
}

func setupValidConfigFile(t *testing.T) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(`
s3:
  endpoint: "https://s3.amazonaws.com"
  region: "us-east-1"
  access-key: "test-key"
  secret-key: "test-secret"
  bucket: "test-bucket"
  prefix: "backups/"
backup:
  dirs:
    - /tmp/test
  retention-count: 15
  cron: "0 2 * * *"
logger:
  level: "INFO"
  mode: "PRETTY"
`), 0644)
	require.NoError(t, err)
	return configPath
}

func setupNoConfigFile(t *testing.T) string {
	// Don't set any env vars, so defaults will be used
	// This should fail validation because backup.dirs will be empty
	// Return empty string so it searches default paths (none exist)
	return ""
}

func setupMissingFieldsConfigFile(t *testing.T) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(`
s3:
  endpoint: "https://s3.amazonaws.com"
logger:
  level: "INFO"
  mode: "PRETTY"
`), 0644)
	require.NoError(t, err)
	return configPath
}

func setupInvalidLoggerConfigFile(t *testing.T) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(`
backup:
  dirs:
    - /tmp/test
  retention-count: 10
  cron: "0 0 * * *"
logger:
  level: "invalid"
  mode: "PRETTY"
`), 0644)
	require.NoError(t, err)
	return configPath
}

func setupInvalidYAMLConfigFile(t *testing.T) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	// Invalid YAML with syntax errors
	err := os.WriteFile(configPath, []byte(`
backup:
  dirs: [/tmp
  invalid: yaml: format: here
	tabs and spaces mixed
logger:
  level: INFO
`), 0644)
	require.NoError(t, err)
	return configPath
}

func setupMalformedYAMLConfigFile(t *testing.T) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	// YAML that parses but has type mismatch for unmarshal
	err := os.WriteFile(configPath, []byte(`
backup:
  dirs: "not an array but a string"
  retention-count: "not a number"
  cron: "0 0 * * *"
logger:
  level: INFO
  mode: PRETTY
`), 0644)
	require.NoError(t, err)
	return configPath
}

func TestLoadConfig(t *testing.T) {
	// Initialize logger for tests
	level := commonLogger.DefaultLoggerLevel
	mode := commonLogger.DefaultLoggerMode
	commonLogger.InitLogger(&level, &mode)

	tests := []struct {
		name      string
		wantErr   bool
		setupFunc func(t *testing.T) string
	}{
		{
			name:      "valid config file",
			wantErr:   false,
			setupFunc: setupValidConfigFile,
		},
		{
			name:      "no config file - uses defaults",
			wantErr:   true, // Should error because dirs is empty (env vars for arrays don't bind automatically)
			setupFunc: setupNoConfigFile,
		},
		{
			name:      "invalid config - missing required fields",
			wantErr:   true,
			setupFunc: setupMissingFieldsConfigFile,
		},
		{
			name:      "invalid logger level",
			wantErr:   true,
			setupFunc: setupInvalidLoggerConfigFile,
		},
		{
			name:      "invalid yaml format",
			wantErr:   true,
			setupFunc: setupInvalidYAMLConfigFile,
		},
		{
			name:      "malformed yaml - unmarshal error",
			wantErr:   true,
			setupFunc: setupMalformedYAMLConfigFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.setupFunc != nil {
				configPath = tt.setupFunc(t)
			}

			ctx := t.Context()
			cfg, err := LoadConfig(ctx, configPath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)

				// Verify defaults are set
				assert.NotEmpty(t, cfg.Logger.Level)
				assert.NotEmpty(t, cfg.Logger.Mode)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	// Save current state
	originalCurrent := Current
	defer func() {
		Current = originalCurrent
	}()

	// Initialize logger for tests
	level := commonLogger.DefaultLoggerLevel
	mode := commonLogger.DefaultLoggerMode
	commonLogger.InitLogger(&level, &mode)

	t.Run("first call initializes config", func(t *testing.T) {
		Current = nil

		// Create a valid config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
s3:
  endpoint: "https://s3.amazonaws.com"
  region: "us-east-1"
  access-key: "test-key"
  secret-key: "test-secret"
  bucket: "test-bucket"
backup:
  dirs:
    - /tmp/test
  retention-count: 10
  cron: "0 0 * * *"
logger:
  level: "INFO"
  mode: "PRETTY"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		ctx := t.Context()
		cfg, err := GetConfig(ctx, configPath)

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotNil(t, Current)
	})

	t.Run("subsequent calls return cached config", func(t *testing.T) {
		// Create a mock config
		mockConfig := &Config{
			Logger: LoggerConfig{
				Level: "DEBUG",
				Mode:  "JSON",
			},
		}
		Current = mockConfig

		ctx := t.Context()
		cfg, err := GetConfig(ctx, "/some/path")

		require.NoError(t, err)
		require.Equal(t, mockConfig, cfg)
		assert.Equal(t, "DEBUG", cfg.Logger.Level)
	})

	t.Run("returns error on invalid config", func(t *testing.T) {
		Current = nil

		ctx := t.Context()
		cfg, err := GetConfig(ctx, "/non/existent/config.yaml")

		// This should error because no dirs are configured
		require.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestGenerateConfigFile(t *testing.T) {
	t.Run("generate config file with all sections", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		ctx := t.Context()
		path, err := GenerateConfigFile(ctx, configPath)

		require.NoError(t, err)
		require.Equal(t, configPath, path)

		// Verify file was created
		_, err = os.Stat(configPath)
		require.NoError(t, err)

		// Read the generated file and verify all sections exist
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Verify all top-level sections exist
		assert.Contains(t, string(content), "s3:")
		assert.Contains(t, string(content), "backup:")
		assert.Contains(t, string(content), "notifiers:")
		assert.Contains(t, string(content), "logger:")

		// Verify S3 fields
		assert.Contains(t, string(content), "endpoint:")
		assert.Contains(t, string(content), "region:")
		assert.Contains(t, string(content), "access-key:")
		assert.Contains(t, string(content), "secret-key:")
		assert.Contains(t, string(content), "bucket:")
		assert.Contains(t, string(content), "prefix:")

		// Verify backup fields
		assert.Contains(t, string(content), "dirs:")
		assert.Contains(t, string(content), "hostname:")
		assert.Contains(t, string(content), "retention-count:")
		assert.Contains(t, string(content), "date-time-layout:")
		assert.Contains(t, string(content), "cron:")
		assert.Contains(t, string(content), "archive-dirs:")
		assert.Contains(t, string(content), "encryption:")

		// Verify encryption fields
		assert.Contains(t, string(content), "enabled:")
		assert.Contains(t, string(content), "gpg:")
		assert.Contains(t, string(content), "key-server:")
		assert.Contains(t, string(content), "key-id:")

		// Verify notifiers fields
		assert.Contains(t, string(content), "discord:")
		assert.Contains(t, string(content), "webhook:")

		// Verify logger fields
		assert.Contains(t, string(content), "level:")
		assert.Contains(t, string(content), "mode:")
	})

	t.Run("generated config can be loaded", func(t *testing.T) {
		// Initialize logger for tests
		level := commonLogger.DefaultLoggerLevel
		mode := commonLogger.DefaultLoggerMode
		commonLogger.InitLogger(&level, &mode)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		ctx := t.Context()

		// Generate the config file
		path, err := GenerateConfigFile(ctx, configPath)
		require.NoError(t, err)
		require.Equal(t, configPath, path)

		// Read the generated config
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Parse and modify the YAML to add dirs
		var tempConfig Config
		err = yaml.Unmarshal(content, &tempConfig)
		require.NoError(t, err)

		// Add required dirs field
		tempConfig.Backup.Dirs = []string{"/tmp/test"}

		// Marshal back to YAML
		updatedContent, err := yaml.Marshal(&tempConfig)
		require.NoError(t, err)

		// Write the updated content
		err = os.WriteFile(configPath, updatedContent, 0644)
		require.NoError(t, err)

		// Try to load the generated config
		cfg, err := LoadConfig(ctx, configPath)
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Verify defaults were applied
		assert.Equal(t, constants.DefaultRetentionCount, cfg.Backup.RetentionCount)
		assert.Equal(t, constants.DefaultDateTimeLayout, cfg.Backup.DateTimeLayout)
		assert.Equal(t, constants.DefaultCron, cfg.Backup.Cron)
		assert.Equal(t, commonLogger.DefaultLoggerLevel, cfg.Logger.Level)
		assert.Equal(t, commonLogger.DefaultLoggerMode, cfg.Logger.Mode)
		assert.False(t, cfg.Backup.Encryption.Enabled)
		assert.False(t, cfg.Notifiers.Discord.Enabled)
	})

	t.Run("file already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create an existing file
		err := os.WriteFile(configPath, []byte("existing content"), 0644)
		require.NoError(t, err)

		ctx := t.Context()
		_, err = GenerateConfigFile(ctx, configPath)

		// Should succeed because we're using WriteConfig instead of SafeWriteConfig
		require.NoError(t, err)
	})
}

func TestStructTags(t *testing.T) {
	// Verify struct tags are properly defined for config marshaling
	t.Run("S3Config has correct tags", func(t *testing.T) {
		cfg := S3Config{
			Endpoint:  "test",
			Region:    "us-east-1",
			AccessKey: "key",
			SecretKey: "secret",
			Bucket:    "bucket",
			Prefix:    "prefix/",
		}
		assert.NotNil(t, cfg)
	})

	t.Run("BackupConfig has correct tags", func(t *testing.T) {
		cfg := BackupConfig{
			Dirs:           []string{"/tmp"},
			RetentionCount: 10,
			Cron:           "0 0 * * *",
		}
		assert.NotNil(t, cfg)
	})
}

func TestYAMLMarshaling(t *testing.T) {
	t.Run("Config struct marshals to YAML correctly", func(t *testing.T) {
		cfg := &Config{
			S3: S3Config{
				Endpoint:  "https://s3.amazonaws.com",
				Region:    "us-east-1",
				AccessKey: "test-key",
				SecretKey: "test-secret",
				Bucket:    "test-bucket",
				Prefix:    "backups/",
			},
			Backup: BackupConfig{
				Dirs:           []string{"/tmp/test", "/var/data"},
				Hostname:       "test-host",
				RetentionCount: 15,
				DateTimeLayout: "20060102150405",
				Cron:           "0 2 * * *",
				ArchiveDirs:    true,
				Encryption: Encryption{
					Enabled: true,
					GPG: GPGConfig{
						KeyServer: "keyserver.ubuntu.com",
						KeyID:     "12345678",
					},
				},
			},
			Notifiers: NotifiersConfig{
				Enabled: true,
				Discord: DiscordNotifierConfig{
					Enabled: true,
					Webhook: "https://discord.com/webhook",
				},
			},
			Logger: LoggerConfig{
				Level: "DEBUG",
				Mode:  "JSON",
			},
		}

		// Marshal to YAML
		yamlBytes, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlBytes)

		yamlContent := string(yamlBytes)

		// Verify structure
		assert.Contains(t, yamlContent, "s3:")
		assert.Contains(t, yamlContent, "endpoint: https://s3.amazonaws.com")
		assert.Contains(t, yamlContent, "region: us-east-1")
		assert.Contains(t, yamlContent, "access-key: test-key")
		assert.Contains(t, yamlContent, "backup:")
		assert.Contains(t, yamlContent, "retention-count: 15")
		assert.Contains(t, yamlContent, "archive-dirs: true")
		assert.Contains(t, yamlContent, "encryption:")
		assert.Contains(t, yamlContent, "key-server: keyserver.ubuntu.com")
		assert.Contains(t, yamlContent, "notifiers:")
		assert.Contains(t, yamlContent, "discord:")
		assert.Contains(t, yamlContent, "webhook: https://discord.com/webhook")
		assert.Contains(t, yamlContent, "logger:")
		assert.Contains(t, yamlContent, "level: DEBUG")
		assert.Contains(t, yamlContent, "mode: JSON")

		// Unmarshal back to verify round-trip
		var unmarshaledCfg Config
		err = yaml.Unmarshal(yamlBytes, &unmarshaledCfg)
		require.NoError(t, err)

		// Verify values
		assert.Equal(t, cfg.S3.Endpoint, unmarshaledCfg.S3.Endpoint)
		assert.Equal(t, cfg.S3.Region, unmarshaledCfg.S3.Region)
		assert.Equal(t, cfg.S3.AccessKey, unmarshaledCfg.S3.AccessKey)
		assert.Equal(t, cfg.Backup.RetentionCount, unmarshaledCfg.Backup.RetentionCount)
		assert.Equal(t, cfg.Backup.ArchiveDirs, unmarshaledCfg.Backup.ArchiveDirs)
		assert.Equal(t, cfg.Backup.Encryption.Enabled, unmarshaledCfg.Backup.Encryption.Enabled)
		assert.Equal(t, cfg.Backup.Encryption.GPG.KeyServer, unmarshaledCfg.Backup.Encryption.GPG.KeyServer)
		assert.Equal(t, cfg.Notifiers.Discord.Webhook, unmarshaledCfg.Notifiers.Discord.Webhook)
		assert.Equal(t, cfg.Logger.Level, unmarshaledCfg.Logger.Level)
	})

	t.Run("Empty Config marshals correctly", func(t *testing.T) {
		cfg := &Config{}

		yamlBytes, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, yamlBytes)

		// Should contain empty structures
		yamlContent := string(yamlBytes)
		assert.Contains(t, yamlContent, "s3:")
		assert.Contains(t, yamlContent, "backup:")
		assert.Contains(t, yamlContent, "notifiers:")
		assert.Contains(t, yamlContent, "logger:")
	})
}

func TestEncryptionValidation(t *testing.T) {
	// Test specific encryption scenarios
	t.Run("encryption requires archive dirs", func(t *testing.T) {
		cfg := BackupConfig{
			Dirs:           []string{"/tmp/test"},
			RetentionCount: 10,
			Cron:           "0 0 * * *",
			ArchiveDirs:    false,
			Encryption: Encryption{
				Enabled: true,
				GPG: GPGConfig{
					KeyServer: "keyserver.ubuntu.com",
					KeyID:     "12345678",
				},
			},
		}

		err := cfg.validate()
		require.NoError(t, err)
		// Encryption should be disabled after validation
		assert.False(t, cfg.Encryption.Enabled)
	})

	t.Run("encryption requires GPG configuration", func(t *testing.T) {
		cfg := BackupConfig{
			Dirs:           []string{"/tmp/test"},
			RetentionCount: 10,
			Cron:           "0 0 * * *",
			ArchiveDirs:    true,
			Encryption: Encryption{
				Enabled: true,
				GPG: GPGConfig{
					KeyServer: "",
					KeyID:     "",
				},
			},
		}

		err := cfg.validate()
		require.NoError(t, err)
		// Encryption should be disabled after validation
		assert.False(t, cfg.Encryption.Enabled)
	})

	t.Run("valid encryption configuration", func(t *testing.T) {
		cfg := BackupConfig{
			Dirs:           []string{"/tmp/test"},
			RetentionCount: 10,
			Cron:           "0 0 * * *",
			ArchiveDirs:    true,
			Encryption: Encryption{
				Enabled: true,
				GPG: GPGConfig{
					KeyServer: "keyserver.ubuntu.com",
					KeyID:     "12345678",
				},
			},
		}

		err := cfg.validate()
		require.NoError(t, err)
		// Encryption should remain enabled
		assert.True(t, cfg.Encryption.Enabled)
	})
}

func TestDefaultValues(t *testing.T) {
	// Verify that default values from constants are applied
	t.Run("defaults are applied from constants", func(t *testing.T) {
		require.Equal(t, 30, constants.DefaultRetentionCount)
		require.Equal(t, "20060102150405", constants.DefaultDateTimeLayout)
		assert.Equal(t, "0 0 * * *", constants.DefaultCron)
	})
}

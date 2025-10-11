package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/spf13/cobra"
)

var InitConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize application config",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		cPath := cmd.Root().PersistentFlags().Lookup("config").Value.String()

		if configPath, err := config.GenerateConfigFile(cmd.Context(), cPath); err != nil {
			slog.ErrorContext(ctx, "error generating config file", "error", err)
			os.Exit(1)
		} else {
			fmt.Printf("\n\nConfig file path: %s\n", configPath)                                            //nolint:forbidigo // CLI output requires fmt.Printf
			fmt.Printf("Empty config file is loaded at above location. Edit config as per your needs.\n\n") //nolint:forbidigo // CLI output requires fmt.Printf
		}
	},
}

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package init

import (
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize application (create necessary config and other things)",
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
	},
}

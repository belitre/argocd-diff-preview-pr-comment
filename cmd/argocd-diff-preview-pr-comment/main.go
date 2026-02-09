package main

import (
	"os"
	"strings"

	"github.com/belitre/argocd-diff-preview-pr-comment/cmd/argocd-diff-preview-pr-comment/add"
	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/version"
	"github.com/spf13/cobra"
)

var (
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "argocd-diff-preview-pr-comment",
	Short: version.GetDescription(),
	Long: `argocd-diff-preview-pr-comment processes ArgoCD application diffs from argocd-diff-preview,
splits them by application, and posts organized comments on GitHub pull requests.

This tool helps teams review ArgoCD changes more effectively by providing clear,
application-specific diff summaries directly in PR comments.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger with the specified log level
		level, err := logger.ParseLogLevel(logLevel)
		if err != nil {
			return err
		}
		return logger.Initialize(level)
	},
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()
		log.Info("Welcome to argocd-diff-preview-pr-comment!")
		log.Infof("Version: %s", version.GetFullVersion())
		log.Info("Use --help to see available commands and flags.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()
		log.Infof("Version: %s", version.GetVersion())
		log.Infof("Commit: %s", version.GetCommit())
		log.Infof("Description: %s", version.GetDescription())
	},
}

func init() {
	rootCmd.Version = version.GetVersion()
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(add.NewAddCommand())

	// Add global log-level flag
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Set the logging level ("+strings.Join(logger.ValidLogLevels(), ", ")+")")
}

func main() {
	defer logger.Sync()

	if err := rootCmd.Execute(); err != nil {
		log := logger.GetLogger()
		log.Errorf("Error: %v", err)
		os.Exit(1)
	}
}

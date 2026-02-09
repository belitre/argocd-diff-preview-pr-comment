package add

import (
	"fmt"
	"os"
	"time"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/github"
	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/splitter"
	"github.com/spf13/cobra"
)

var (
	diffFile  string
	maxLength int

	githubToken string
	prRef       string

	maxRetries     int
	retryDelay     time.Duration
	backoffFactor  float64
	requestTimeout time.Duration

	dryRun bool
)

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Post ArgoCD diff to GitHub PR as comments",
		Long: `Process ArgoCD diff files and post them as comments to GitHub Pull Requests.

If the diff file exceeds the specified max length, it will be split into
multiple comments. The tool automatically handles GitHub rate limiting with
configurable retry logic.

GitHub Token:
The GitHub token can be provided via:
  - --github-token flag
  - GH_TOKEN environment variable
  - GITHUB_TOKEN environment variable

PR Reference:
Accepts the following formats:
  - owner/repo#123
  - https://github.com/owner/repo/pull/123`,
		RunE: runAdd,
	}

	cmd.Flags().StringVarP(&diffFile, "file", "f", "", "Path to the diff markdown file (required)")
	cmd.Flags().IntVarP(&maxLength, "max-length", "m", 65536, "Maximum length in bytes for a single comment (default: 65536, GitHub's limit)")

	cmd.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub personal access token (can also use GH_TOKEN or GITHUB_TOKEN env vars)")
	cmd.Flags().StringVarP(&prRef, "pr", "p", "", "Pull request reference (e.g., owner/repo#123 or PR URL) (required)")

	cmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum number of retry attempts for failed requests")
	cmd.Flags().DurationVar(&retryDelay, "retry-delay", 2*time.Second, "Initial delay between retries")
	cmd.Flags().Float64Var(&backoffFactor, "backoff-factor", 2.0, "Exponential backoff multiplier for retries")
	cmd.Flags().DurationVar(&requestTimeout, "request-timeout", 30*time.Second, "HTTP request timeout")

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without actually posting comments")

	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("pr")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()

	token := githubToken
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitHub token is required. Provide it via --github-token flag, GH_TOKEN, or GITHUB_TOKEN environment variable")
	}

	owner, repo, prNumber, err := github.ValidatePRReference(prRef)
	if err != nil {
		return fmt.Errorf("invalid PR reference: %w", err)
	}

	log.Infof("Target PR: %s/%s#%d", owner, repo, prNumber)
	log.Infof("Processing diff file: %s", diffFile)
	log.Infof("Max comment length: %d bytes", maxLength)

	if dryRun {
		log.Info("DRY RUN MODE - No comments will be posted")
	}

	size, err := splitter.CountFileSize(diffFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	log.Infof("Input file size: %d bytes", size)

	results, err := splitter.SplitDiffFile(diffFile, maxLength)
	if err != nil {
		return fmt.Errorf("failed to split diff file: %w", err)
	}

	if len(results) == 1 && results[0].TotalParts == 1 {
		log.Info("No splitting needed - file is within size limit")
		log.Infof("Content size: %d bytes", results[0].Size)
	} else {
		log.Infof("Split file into %d parts", len(results))
		for _, result := range results {
			log.Debugf("Part %d of %d - Length: %d bytes", result.PartNumber, result.TotalParts, result.Size)
		}
	}

	ghConfig := github.Config{
		Token:          token,
		MaxRetries:     maxRetries,
		RetryDelay:     retryDelay,
		BackoffFactor:  backoffFactor,
		RequestTimeout: requestTimeout,
	}
	client := github.NewClient(ghConfig)

	log.Infof("Posting %d comment(s) to PR...", len(results))

	for _, result := range results {
		log.Infof("Posting part %d of %d...", result.PartNumber, result.TotalParts)

		err := client.PostPRComment(owner, repo, prNumber, result.Content, ghConfig, dryRun)
		if err != nil {
			return fmt.Errorf("failed to post comment part %d: %w", result.PartNumber, err)
		}

		if result.PartNumber < result.TotalParts && !dryRun {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if dryRun {
		log.Info("DRY RUN completed - No comments were posted")
	} else {
		log.Info("Successfully posted all comments to PR")
	}

	return nil
}

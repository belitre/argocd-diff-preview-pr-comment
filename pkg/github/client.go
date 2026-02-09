package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
	"github.com/google/go-github/v69/github"
)

// Client represents a GitHub API client
type Client struct {
	client *github.Client
}

// Config holds configuration for GitHub client
type Config struct {
	Token          string
	MaxRetries     int
	RetryDelay     time.Duration
	BackoffFactor  float64
	RequestTimeout time.Duration
}

// NewClient creates a new GitHub client
func NewClient(config Config) *Client {
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	return &Client{
		client: github.NewClient(httpClient).WithAuthToken(config.Token),
	}
}

// PostPRComment posts a comment to a GitHub PR with retry logic
func (c *Client) PostPRComment(owner, repo string, prNumber int, comment string, config Config, dryRun bool) error {
	log := logger.GetLogger()
	ctx := context.Background()

	if dryRun {
		log.Infof("[DRY RUN] Would post comment to PR #%d in %s/%s", prNumber, owner, repo)
		log.Infof("[DRY RUN] Comment length: %d bytes", len(comment))
		log.Debugf("[DRY RUN] Comment content:\n%s", comment)
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(float64(config.RetryDelay) * float64(attempt) * config.BackoffFactor)
			log.Warnf("Retry attempt %d/%d after %v", attempt, config.MaxRetries, delay)
			time.Sleep(delay)
		}

		issueComment := &github.IssueComment{
			Body: github.String(comment),
		}

		_, resp, err := c.client.Issues.CreateComment(ctx, owner, repo, prNumber, issueComment)
		if err != nil {
			lastErr = err

			// Check if it's a rate limit error
			if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) {
				if resp.Rate.Remaining == 0 {
					waitTime := time.Until(resp.Rate.Reset.Time)
					if waitTime > 0 {
						log.Warnf("Rate limited. Waiting %v until reset at %v", waitTime, resp.Rate.Reset.Time)
						time.Sleep(waitTime + time.Second) // Add 1 second buffer
						continue
					}
				}
			}

			// For other errors, only retry if we haven't exhausted attempts
			if attempt < config.MaxRetries {
				log.Warnf("Failed to post comment: %v", err)
			}
			continue
		}

		// Success - log rate limit info for monitoring
		if resp != nil && resp.Rate.Remaining >= 0 {
			log.Debugf("Rate limit remaining: %d, resets at: %v", resp.Rate.Remaining, resp.Rate.Reset.Time)
		}

		log.Infof("Successfully posted comment to PR #%d", prNumber)
		return nil
	}

	return fmt.Errorf("failed to post comment after %d retries: %w", config.MaxRetries, lastErr)
}

// ValidatePRReference validates and parses a GitHub PR reference
// Accepts formats: owner/repo#123, https://github.com/owner/repo/pull/123
func ValidatePRReference(ref string) (owner, repo string, prNumber int, err error) {
	// Handle URL format: https://github.com/owner/repo/pull/123
	if strings.HasPrefix(ref, "https://github.com/") || strings.HasPrefix(ref, "http://github.com/") {
		ref = strings.TrimPrefix(ref, "https://github.com/")
		ref = strings.TrimPrefix(ref, "http://github.com/")

		parts := strings.Split(ref, "/")
		if len(parts) >= 4 && parts[2] == "pull" {
			owner = parts[0]
			repo = parts[1]
			prNumber, err = strconv.Atoi(parts[3])
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid PR number in URL: %s", parts[3])
			}
			return owner, repo, prNumber, nil
		}
		return "", "", 0, fmt.Errorf("invalid GitHub PR URL format")
	}

	// Handle short format: owner/repo#123
	if strings.Contains(ref, "#") {
		parts := strings.Split(ref, "#")
		if len(parts) != 2 {
			return "", "", 0, fmt.Errorf("invalid PR reference format, expected owner/repo#123")
		}

		repoParts := strings.Split(parts[0], "/")
		if len(repoParts) != 2 {
			return "", "", 0, fmt.Errorf("invalid repository format, expected owner/repo")
		}

		owner = repoParts[0]
		repo = repoParts[1]
		prNumber, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid PR number: %s", parts[1])
		}

		return owner, repo, prNumber, nil
	}

	return "", "", 0, fmt.Errorf("invalid PR reference format, expected owner/repo#123 or https://github.com/owner/repo/pull/123")
}

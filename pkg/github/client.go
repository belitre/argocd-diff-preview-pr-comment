package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
)

// Client represents a GitHub API client
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
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
	return &Client{
		token:   config.Token,
		baseURL: "https://api.github.com",
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// CommentRequest represents a comment to be posted
type CommentRequest struct {
	Body string `json:"body"`
}

// RateLimitInfo contains GitHub rate limit information
type RateLimitInfo struct {
	Remaining int
	Reset     time.Time
}

// PostPRComment posts a comment to a GitHub PR with retry logic
func (c *Client) PostPRComment(owner, repo string, prNumber int, comment string, config Config, dryRun bool) error {
	log := logger.GetLogger()

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.baseURL, owner, repo, prNumber)

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

		err := c.doPostComment(url, comment)
		if err == nil {
			log.Infof("Successfully posted comment to PR #%d", prNumber)
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		if rateLimitErr, ok := err.(*RateLimitError); ok {
			waitTime := time.Until(rateLimitErr.ResetTime)
			if waitTime > 0 {
				log.Warnf("Rate limited. Waiting %v until reset at %v", waitTime, rateLimitErr.ResetTime)
				time.Sleep(waitTime + time.Second) // Add 1 second buffer
				continue
			}
		}

		// For other errors, only retry if we haven't exhausted attempts
		if attempt < config.MaxRetries {
			log.Warnf("Failed to post comment: %v", err)
		}
	}

	return fmt.Errorf("failed to post comment after %d retries: %w", config.MaxRetries, lastErr)
}

// doPostComment performs the actual HTTP request to post a comment
func (c *Client) doPostComment(url, comment string) error {
	log := logger.GetLogger()

	commentReq := CommentRequest{Body: comment}
	jsonData, err := json.Marshal(commentReq)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		rateLimitInfo := c.parseRateLimitHeaders(resp)
		if rateLimitInfo.Remaining == 0 {
			return &RateLimitError{
				ResetTime: rateLimitInfo.Reset,
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Log rate limit info for monitoring
	rateLimitInfo := c.parseRateLimitHeaders(resp)
	log.Debugf("Rate limit remaining: %d, resets at: %v", rateLimitInfo.Remaining, rateLimitInfo.Reset)

	return nil
}

// parseRateLimitHeaders extracts rate limit information from response headers
func (c *Client) parseRateLimitHeaders(resp *http.Response) RateLimitInfo {
	info := RateLimitInfo{}

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		info.Remaining, _ = strconv.Atoi(remaining)
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = time.Unix(timestamp, 0)
		}
	}

	return info
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	ResetTime time.Time
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, resets at %v", e.ResetTime)
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

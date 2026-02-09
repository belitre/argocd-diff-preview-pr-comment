package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
)

func TestNewClient(t *testing.T) {
	config := Config{
		Token:          "test-token",
		RequestTimeout: 30 * time.Second,
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.client == nil {
		t.Error("GitHub client not initialized")
	}
}

func TestPostPRComment_DryRun(t *testing.T) {
	config := Config{
		Token:          "test-token",
		RequestTimeout: 30 * time.Second,
	}

	client := NewClient(config)

	err := client.PostPRComment("owner", "repo", 123, "Test comment", config, true)
	if err != nil {
		t.Errorf("DryRun should not return error, got: %v", err)
	}
}

func TestPostPRComment_Success(t *testing.T) {
	// Create a test server that mimics GitHub API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/repos/owner/repo/issues/123/comments") {
			t.Errorf("Expected path to contain '/repos/owner/repo/issues/123/comments', got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Authorization header to start with 'Bearer ', got %q", authHeader)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "4999")
		w.Header().Set("X-RateLimit-Reset", "1234567890")
		w.WriteHeader(http.StatusCreated)

		response := map[string]interface{}{
			"id":   123,
			"body": "Test comment",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom base URL pointing to test server
	config := Config{
		Token:          "test-token",
		RequestTimeout: 30 * time.Second,
		MaxRetries:     0,
	}

	httpClient := &http.Client{Timeout: config.RequestTimeout}
	testClient := &Client{
		client: github.NewClient(httpClient).WithAuthToken(config.Token),
	}

	// Override the base URL to point to our test server
	testClient.client, _ = testClient.client.WithEnterpriseURLs(server.URL, server.URL)

	err := testClient.PostPRComment("owner", "repo", 123, "Test comment", config, false)
	if err != nil {
		t.Errorf("PostPRComment failed: %v", err)
	}
}

func TestValidatePRReference(t *testing.T) {
	tests := []struct {
		name          string
		ref           string
		expectedOwner string
		expectedRepo  string
		expectedPR    int
		shouldError   bool
	}{
		{
			name:          "Valid short format",
			ref:           "belitre/my-repo#123",
			expectedOwner: "belitre",
			expectedRepo:  "my-repo",
			expectedPR:    123,
			shouldError:   false,
		},
		{
			name:          "Valid URL format",
			ref:           "https://github.com/belitre/my-repo/pull/456",
			expectedOwner: "belitre",
			expectedRepo:  "my-repo",
			expectedPR:    456,
			shouldError:   false,
		},
		{
			name:        "Invalid format - no hash or slash",
			ref:         "invalid",
			shouldError: true,
		},
		{
			name:        "Invalid format - missing PR number",
			ref:         "owner/repo#",
			shouldError: true,
		},
		{
			name:        "Invalid format - non-numeric PR",
			ref:         "owner/repo#abc",
			shouldError: true,
		},
		{
			name:        "Invalid format - missing repo",
			ref:         "owner#123",
			shouldError: true,
		},
		{
			name:        "Invalid URL - incomplete",
			ref:         "https://github.com/owner/repo",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, prNumber, err := ValidatePRReference(tt.ref)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if owner != tt.expectedOwner {
				t.Errorf("Expected owner %s, got %s", tt.expectedOwner, owner)
			}

			if repo != tt.expectedRepo {
				t.Errorf("Expected repo %s, got %s", tt.expectedRepo, repo)
			}

			if prNumber != tt.expectedPR {
				t.Errorf("Expected PR number %d, got %d", tt.expectedPR, prNumber)
			}
		})
	}
}

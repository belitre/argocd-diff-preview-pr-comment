package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", client.token)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestDoPostComment_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header 'token test-token', got %q", r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	client := &Client{
		token: "test-token",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	err := client.doPostComment(server.URL, "Test comment")
	if err != nil {
		t.Errorf("doPostComment failed: %v", err)
	}
}

func TestDoPostComment_RateLimit(t *testing.T) {
	// Create a test server that returns rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1234567890")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := &Client{
		token: "test-token",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	err := client.doPostComment(server.URL, "Test comment")
	if err == nil {
		t.Error("Expected rate limit error, got nil")
	}

	// Check if it's a RateLimitError
	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Errorf("Expected RateLimitError, got %T", err)
	} else {
		if rateLimitErr.ResetTime.Unix() != 1234567890 {
			t.Errorf("Expected reset time 1234567890, got %d", rateLimitErr.ResetTime.Unix())
		}
	}
}

func TestDoPostComment_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad request"}`))
	}))
	defer server.Close()

	client := &Client{
		token: "test-token",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	err := client.doPostComment(server.URL, "Test comment")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error message to contain '400', got: %v", err)
	}
}

func TestParseRateLimitHeaders(t *testing.T) {
	tests := []struct {
		name              string
		remaining         string
		reset             string
		expectedRemaining int
		expectedReset     int64
	}{
		{
			name:              "Valid headers",
			remaining:         "42",
			reset:             "1234567890",
			expectedRemaining: 42,
			expectedReset:     1234567890,
		},
		{
			name:              "Zero remaining",
			remaining:         "0",
			reset:             "1234567890",
			expectedRemaining: 0,
			expectedReset:     1234567890,
		},
		{
			name:              "Missing headers",
			remaining:         "",
			reset:             "",
			expectedRemaining: 0,
			expectedReset:     0,
		},
		{
			name:              "Invalid reset time",
			remaining:         "10",
			reset:             "invalid",
			expectedRemaining: 10,
			expectedReset:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{},
			}
			if tt.remaining != "" {
				resp.Header.Set("X-RateLimit-Remaining", tt.remaining)
			}
			if tt.reset != "" {
				resp.Header.Set("X-RateLimit-Reset", tt.reset)
			}

			client := &Client{}
			info := client.parseRateLimitHeaders(resp)

			if info.Remaining != tt.expectedRemaining {
				t.Errorf("Expected remaining %d, got %d", tt.expectedRemaining, info.Remaining)
			}

			if tt.expectedReset != 0 && info.Reset.Unix() != tt.expectedReset {
				t.Errorf("Expected reset time %d, got %d", tt.expectedReset, info.Reset.Unix())
			}

			if tt.expectedReset == 0 && !info.Reset.IsZero() {
				t.Errorf("Expected zero reset time, got %v", info.Reset)
			}
		})
	}
}

func TestRateLimitError_Error(t *testing.T) {
	resetTime := time.Unix(1234567890, 0)
	err := &RateLimitError{
		ResetTime: resetTime,
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "rate limit") {
		t.Errorf("Error message should contain 'rate limit', got: %s", errMsg)
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

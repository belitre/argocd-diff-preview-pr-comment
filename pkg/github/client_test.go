package github

import (
	"testing"
)

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

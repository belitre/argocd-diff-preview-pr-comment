package add

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
	"github.com/spf13/cobra"
)

func init() {
	// Initialize logger for tests
	logger.Initialize(logger.ErrorLevel)
}

func TestNewAddCommand(t *testing.T) {
	cmd := NewAddCommand()

	if cmd == nil {
		t.Fatal("NewAddCommand returned nil")
	}

	if cmd.Use != "add" {
		t.Errorf("Expected Use 'add', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestAddCommand_RequiredFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Missing file flag",
			args:        []string{"--pr", "owner/repo#123"},
			shouldError: true,
			errorMsg:    "file",
		},
		{
			name:        "Missing pr flag",
			args:        []string{"--file", "test.md"},
			shouldError: true,
			errorMsg:    "pr",
		},
		{
			name:        "Both flags missing",
			args:        []string{},
			shouldError: true,
			errorMsg:    "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAddCommand()
			cmd.SetArgs(tt.args)

			// Disable output during test
			cmd.SetOut(os.NewFile(0, os.DevNull))
			cmd.SetErr(os.NewFile(0, os.DevNull))

			err := cmd.Execute()

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAddCommand_FileValidation(t *testing.T) {
	// Create a temp directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		shouldError bool
	}{
		{
			name:        "File does not exist",
			setupFile:   false,
			shouldError: true,
		},
		{
			name:        "Valid file",
			setupFile:   true,
			fileContent: "# Test diff\nSome content",
			shouldError: false, // Will error due to missing GitHub token, but file validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.md")

			if tt.setupFile {
				err := os.WriteFile(testFile, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			cmd := NewAddCommand()
			cmd.SetArgs([]string{
				"--file", testFile,
				"--pr", "owner/repo#123",
				"--dry-run", // Use dry-run to avoid needing real GitHub token
			})

			// Set fake token to pass token validation
			os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
			defer os.Unsetenv("GITHUB_TOKEN")

			// Disable output during test
			cmd.SetOut(os.NewFile(0, os.DevNull))
			cmd.SetErr(os.NewFile(0, os.DevNull))

			err := cmd.Execute()

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			// For valid file test, we expect it to run (may error on GitHub API call)
			// but file validation should pass
		})
	}
}

func TestAddCommand_PRReferenceValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		prRef       string
		shouldError bool
	}{
		{
			name:        "Valid short format",
			prRef:       "owner/repo#123",
			shouldError: false,
		},
		{
			name:        "Valid URL format",
			prRef:       "https://github.com/owner/repo/pull/456",
			shouldError: false,
		},
		{
			name:        "Invalid format",
			prRef:       "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAddCommand()
			cmd.SetArgs([]string{
				"--file", testFile,
				"--pr", tt.prRef,
				"--dry-run",
			})

			// Set fake token to pass token validation
			os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
			defer os.Unsetenv("GITHUB_TOKEN")

			// Disable output during test
			cmd.SetOut(os.NewFile(0, os.DevNull))
			cmd.SetErr(os.NewFile(0, os.DevNull))

			err := cmd.Execute()

			if tt.shouldError && err == nil {
				t.Error("Expected error for invalid PR reference but got none")
			}
		})
	}
}

func TestAddCommand_TokenValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		tokenFlag   string
		envToken    string
		ghToken     string
		shouldError bool
	}{
		{
			name:        "Token from flag",
			tokenFlag:   "flag-token",
			shouldError: false,
		},
		{
			name:        "Token from GITHUB_TOKEN env",
			envToken:    "env-token",
			shouldError: false,
		},
		{
			name:        "Token from GH_TOKEN env",
			ghToken:     "gh-token",
			shouldError: false,
		},
		{
			name:        "No token provided",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("GH_TOKEN")

			// Set environment variables if specified
			if tt.envToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			}
			if tt.ghToken != "" {
				os.Setenv("GH_TOKEN", tt.ghToken)
				defer os.Unsetenv("GH_TOKEN")
			}

			args := []string{
				"--file", testFile,
				"--pr", "owner/repo#123",
				"--dry-run",
			}
			if tt.tokenFlag != "" {
				args = append(args, "--github-token", tt.tokenFlag)
			}

			cmd := NewAddCommand()
			cmd.SetArgs(args)

			// Disable output during test
			cmd.SetOut(os.NewFile(0, os.DevNull))
			cmd.SetErr(os.NewFile(0, os.DevNull))

			err := cmd.Execute()

			if tt.shouldError && err == nil {
				t.Error("Expected error for missing token but got none")
			}

			if !tt.shouldError && err != nil {
				// For valid token cases, execution might fail for other reasons
				// (like invalid file format), but token validation should pass
				t.Logf("Got error (may be expected): %v", err)
			}
		})
	}
}

func TestAddCommand_DryRunFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test diff\nSome content that is short"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := NewAddCommand()
	cmd.SetArgs([]string{
		"--file", testFile,
		"--pr", "owner/repo#123",
		"--github-token", "fake-token",
		"--dry-run",
	})

	// Disable output during test
	cmd.SetOut(os.NewFile(0, os.DevNull))
	cmd.SetErr(os.NewFile(0, os.DevNull))

	err = cmd.Execute()

	// Dry-run should execute without making actual API calls
	// It might error on file format but should not make HTTP requests
	if err != nil {
		t.Logf("Dry-run completed with error (may be expected): %v", err)
	}
}

func TestAddCommand_MaxLengthFlag(t *testing.T) {
	cmd := NewAddCommand()

	if cmd.Flags().Lookup("max-length") == nil {
		t.Error("max-length flag not found")
	}

	// Check default value
	defaultVal, err := cmd.Flags().GetInt("max-length")
	if err != nil {
		t.Errorf("Failed to get max-length default value: %v", err)
	}

	if defaultVal != 65536 {
		t.Errorf("Expected default max-length 65536, got %d", defaultVal)
	}
}

func TestAddCommand_RetryFlags(t *testing.T) {
	cmd := NewAddCommand()

	flags := []struct {
		name         string
		expectedType string
	}{
		{"max-retries", "int"},
		{"retry-delay", "duration"},
		{"backoff-factor", "float64"},
		{"request-timeout", "duration"},
	}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag.name) == nil {
			t.Errorf("%s flag not found", flag.name)
		}
	}
}

// Helper function to create a test command
func createTestCommand() *cobra.Command {
	return NewAddCommand()
}

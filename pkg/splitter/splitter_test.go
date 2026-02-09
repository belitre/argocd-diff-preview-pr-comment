package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
)

func init() {
	// Initialize logger for tests
	logger.Initialize(logger.ErrorLevel)
}

func TestSplitDiffFile_NoSplitNeeded(t *testing.T) {
	// Create a temp directory for test files
	tmpDir := t.TempDir()

	// Create a small test file
	content := `## Argo CD Diff Preview

Summary:
` + "```yaml" + `
Total: 1 files changed
` + "```" + `

<details>
<summary>app-name</summary>

` + "```diff" + `
Some diff content
` + "```" + `

</details>

_Stats_:
[Applications: 1]
`

	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with a large max length
	results, err := SplitDiffFile(testFile, 10000)
	if err != nil {
		t.Fatalf("SplitDiffFile failed: %v", err)
	}

	// Should return only one result
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].PartNumber != 1 || results[0].TotalParts != 1 {
		t.Errorf("Expected part 1 of 1, got part %d of %d", results[0].PartNumber, results[0].TotalParts)
	}

	if results[0].Size != len(content) {
		t.Errorf("Expected size %d, got %d", len(content), results[0].Size)
	}

	if results[0].Content != content {
		t.Error("Content mismatch in single result")
	}
}

func TestSplitDiffFile_SplitRequired(t *testing.T) {
	// Create a temp directory for test files
	tmpDir := t.TempDir()

	// Create a larger test file that will require splitting
	// Add more lines to ensure the file needs to be split
	diffLines := ""
	for i := 1; i <= 50; i++ {
		diffLines += fmt.Sprintf("+Line %d with some additional content to make it longer\n", i)
	}

	content := `## Argo CD Diff Preview

Summary:
` + "```yaml" + `
Total: 1 files changed
` + "```" + `

<details>
<summary>app-name (path/to/app)</summary>
<br>

` + "```diff" + `
` + diffLines + "```" + `

</details>

_Stats_:
[Applications: 1]
`

	testFile := filepath.Join(tmpDir, "test-split.md")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with a max length that will force splitting
	// The file is ~2600 bytes, so 1000 bytes will definitely split it
	results, err := SplitDiffFile(testFile, 1000)
	if err != nil {
		t.Fatalf("SplitDiffFile failed: %v", err)
	}

	// Should create multiple results
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	// Check that results have correct structure and are within size limit
	for i, result := range results {
		// Verify part numbers
		if result.PartNumber != i+1 {
			t.Errorf("Result %d has incorrect PartNumber: expected %d, got %d", i, i+1, result.PartNumber)
		}

		if result.TotalParts != len(results) {
			t.Errorf("Result %d has incorrect TotalParts: expected %d, got %d", i, len(results), result.TotalParts)
		}

		// Verify size doesn't exceed max length
		if result.Size > 1000 {
			t.Errorf("Result %d size %d exceeds max length 1000", i, result.Size)
		}

		// Verify content size matches Size field
		if len(result.Content) != result.Size {
			t.Errorf("Result %d: Content length %d doesn't match Size field %d", i, len(result.Content), result.Size)
		}

		// Check for header only in first part
		if i == 0 {
			if len(result.Content) < len("## Argo CD Diff Preview") ||
				result.Content[:len("## Argo CD Diff Preview")] != "## Argo CD Diff Preview" {
				t.Errorf("Result %d missing header", i)
			}
			// First part should have the footer
			if len(result.Content) < len("_Stats_:") ||
				result.Content[len(result.Content)-len("**Part 1 of")-20:len(result.Content)-len("**Part 1 of")] != "_Stats_:\n[Applications: 1]\n" {
				// Footer should appear before the part indicator
				t.Logf("Result %d content end: %q", i, result.Content[len(result.Content)-100:])
			}
		} else {
			// Subsequent parts should NOT have header
			if len(result.Content) >= len("## Argo CD Diff Preview") &&
				result.Content[:len("## Argo CD Diff Preview")] == "## Argo CD Diff Preview" {
				t.Errorf("Result %d should not have header", i)
			}
			// Subsequent parts should NOT have footer (_Stats_)
			if len(result.Content) > 0 && containsString(result.Content, "_Stats_:") {
				t.Errorf("Result %d should not have footer (_Stats_:)", i)
			}
		}

		// All parts should have the part indicator "Part X of Y"
		expectedPartIndicator := fmt.Sprintf("**Part %d of %d**", i+1, len(results))
		if !containsString(result.Content, expectedPartIndicator) {
			t.Errorf("Result %d missing part indicator %q", i, expectedPartIndicator)
		}

		// Verify content is not empty
		if len(result.Content) < 10 {
			t.Errorf("Result %d content too short", i)
		}
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCountFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Hello, World!"
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size, err := CountFileSize(testFile)
	if err != nil {
		t.Fatalf("CountFileSize failed: %v", err)
	}

	expectedSize := len(content)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestCountFileSize_NonExistent(t *testing.T) {
	_, err := CountFileSize("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestExtractAppName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard format",
			input:    "<summary>app-name (path/to/app)</summary>",
			expected: "app-name (path/to/app)",
		},
		{
			name:     "With continuation",
			input:    "<summary>app-name (path/to/app) (continuation...)</summary>",
			expected: "app-name (path/to/app) (continuation...)",
		},
		{
			name:     "Missing closing tag",
			input:    "<summary>app-name",
			expected: "",
		},
		{
			name:     "Missing opening tag",
			input:    "app-name</summary>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAppName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReadFileLines(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Line 1\nLine 2\nLine 3"
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lines, err := ReadFileLines(testFile)
	if err != nil {
		t.Fatalf("ReadFileLines failed: %v", err)
	}

	expectedLines := 3
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}

	if lines[0] != "Line 1" || lines[1] != "Line 2" || lines[2] != "Line 3" {
		t.Errorf("Lines content mismatch: got %v", lines)
	}
}

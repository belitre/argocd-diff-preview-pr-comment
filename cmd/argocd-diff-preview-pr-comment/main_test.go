package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the root command
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("Failed to execute root command: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected text
	expectedTexts := []string{
		"Welcome to argocd-diff-preview-pr-comment!",
		"Version:",
	}

	for _, expected := range expectedTexts {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q, but got: %s", expected, output)
		}
	}
}

func TestRootCommandVersion(t *testing.T) {
	// Test version flag
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the version command
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected text
	expectedTexts := []string{
		"Version:",
		"Commit:",
		"Description:",
	}

	for _, expected := range expectedTexts {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q, but got: %s", expected, output)
		}
	}
}

func TestLogLevelFlag(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "valid debug level",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "valid info level",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "valid warn level",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "valid error level",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the version command with log-level flag
			rootCmd.SetArgs([]string{"--log-level", tt.level, "version"})
			err := rootCmd.Execute()

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output (to prevent blocking)
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() with log-level=%s error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}
		})
	}
}

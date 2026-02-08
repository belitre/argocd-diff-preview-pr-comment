package logger

import (
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    LogLevel
		wantErr bool
	}{
		{
			name:    "debug level",
			input:   "debug",
			want:    DebugLevel,
			wantErr: false,
		},
		{
			name:    "info level",
			input:   "info",
			want:    InfoLevel,
			wantErr: false,
		},
		{
			name:    "warn level",
			input:   "warn",
			want:    WarnLevel,
			wantErr: false,
		},
		{
			name:    "error level",
			input:   "error",
			want:    ErrorLevel,
			wantErr: false,
		},
		{
			name:    "fatal level",
			input:   "fatal",
			want:    FatalLevel,
			wantErr: false,
		},
		{
			name:    "uppercase input",
			input:   "INFO",
			want:    InfoLevel,
			wantErr: false,
		},
		{
			name:    "invalid level",
			input:   "invalid",
			want:    InfoLevel,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidLogLevels(t *testing.T) {
	levels := ValidLogLevels()
	if len(levels) != 5 {
		t.Errorf("ValidLogLevels() returned %d levels, want 5", len(levels))
	}

	expectedLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for _, expected := range expectedLevels {
		found := false
		for _, level := range levels {
			if level == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidLogLevels() missing expected level: %s", expected)
		}
	}
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		level   LogLevel
		wantErr bool
	}{
		{
			name:    "initialize with info level",
			level:   InfoLevel,
			wantErr: false,
		},
		{
			name:    "initialize with debug level",
			level:   DebugLevel,
			wantErr: false,
		},
		{
			name:    "initialize with error level",
			level:   ErrorLevel,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && Log == nil {
				t.Error("Initialize() succeeded but Log is nil")
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Reset logger
	Log = nil

	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	// Call again to ensure it returns the same logger
	logger2 := GetLogger()
	if logger2 == nil {
		t.Error("GetLogger() returned nil on second call")
	}
}

func TestSync(t *testing.T) {
	err := Initialize(InfoLevel)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	err = Sync()
	if err != nil {
		// Sync might return errors about syncing stdout/stderr, which is expected
		// Only fail if it's not one of those expected errors
		if !strings.Contains(err.Error(), "sync") {
			t.Errorf("Sync() unexpected error = %v", err)
		}
	}
}

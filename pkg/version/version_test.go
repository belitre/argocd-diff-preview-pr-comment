package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetCommit(t *testing.T) {
	commit := GetCommit()
	if commit == "" {
		t.Error("Commit should not be empty")
	}
}

func TestGetDescription(t *testing.T) {
	description := GetDescription()
	if description == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(description, "ArgoCD") {
		t.Error("Description should mention ArgoCD")
	}
}

func TestGetFullVersion(t *testing.T) {
	fullVersion := GetFullVersion()
	if fullVersion == "" {
		t.Error("Full version should not be empty")
	}
	if !strings.Contains(fullVersion, "commit:") {
		t.Error("Full version should contain commit information")
	}
}

func TestGetInfo(t *testing.T) {
	info := GetInfo()
	if info == "" {
		t.Error("Info should not be empty")
	}

	expectedFields := []string{"Version:", "Commit:", "Description:"}
	for _, field := range expectedFields {
		if !strings.Contains(info, field) {
			t.Errorf("Info should contain %q", field)
		}
	}
}

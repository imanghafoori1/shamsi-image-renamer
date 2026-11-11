package main

import (
	"os"
	"testing"
)

func TestCleanupExifTool(t *testing.T) {
	// Save and restore global state so the test does not interfere with
	// other tests that use ensureExifTool.
	origPath := exiftoolPath
	defer func() { exiftoolPath = origPath }()

	dir, err := os.MkdirTemp("", "renamer-cleanup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	exiftoolPath = dir + string(os.PathSeparator) + "exiftool.exe"

	cleanupExifTool()

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory still exists after cleanup: %v", err)
	}
	if exiftoolPath != "" {
		t.Errorf("exiftoolPath = %q, want empty", exiftoolPath)
	}
}

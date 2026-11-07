package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

//go:embed exiftool
var exiftoolFS embed.FS

var (
	exiftoolPath string
	exiftoolOnce sync.Once
	exiftoolErr  error
)

func ensureExifTool() (string, error) {
	exiftoolOnce.Do(func() {
		exiftoolPath, exiftoolErr = extractExifTool()
	})
	return exiftoolPath, exiftoolErr
}

// extractExifTool tries a system temp directory first. If that fails
// (common on Windows due to execution policy on AppData\Local\Temp),
// it retries next to the running executable.
func extractExifTool() (string, error) {
	bases := []string{
		"",                      // system temp
		filepath.Dir(mustExe()), // next to renamer.exe
	}

	var lastErr error
	for _, base := range bases {
		dir, err := os.MkdirTemp(base, "renamer-exiftool-*")
		if err != nil {
			lastErr = err
			continue
		}
		exe, err := writeExifTool(dir)
		if err != nil {
			os.RemoveAll(dir)
			lastErr = err
			continue
		}
		return exe, nil
	}
	return "", fmt.Errorf("could not extract exiftool: %w", lastErr)
}

func writeExifTool(dstDir string) (string, error) {
	err := fs.WalkDir(exiftoolFS, "exiftool", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("exiftool", path)
		dst := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}

		data, err := exiftoolFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0755)
	})
	if err != nil {
		return "", err
	}

	exePath := filepath.Join(dstDir, "exiftool.exe")
	if _, err := os.Stat(exePath); err != nil {
		return "", fmt.Errorf("exiftool.exe not found: %w", err)
	}
	return exePath, nil
}

func mustExe() string {
	p, err := os.Executable()
	if err != nil {
		return "."
	}
	return p
}

func cleanupExifTool() {
	if exiftoolPath != "" {
		os.RemoveAll(filepath.Dir(exiftoolPath))
		exiftoolPath = ""
	}
}

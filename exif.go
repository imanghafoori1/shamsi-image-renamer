package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type exiftoolOutput struct {
	DateTimeOriginal string `json:"DateTimeOriginal"`
	Model            string `json:"Model"`
	Make             string `json:"Make"`
}

// returns (captureTime, Make, Model).
func (file filePath) readExifToolMeta() (jalali, string, string, error) {
	exifError := func(errNoEXIF error) (jalali, string, string, error) {
		return jalali{}, "", "", errNoEXIF
	}

	path := file.fullPath()
	exiftool, err := ensureExifTool()
	if err != nil {
		return exifError(errNoEXIF)
	}

	tmpFile, err := os.CreateTemp("", "exifargs-*.txt")
	if err != nil {
		return exifError(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write the path as UTF-8, followed by a newline.
	if _, err := tmpFile.WriteString(path + "\n"); err != nil {
		return exifError(err)
	}
	tmpFile.Close()

	cmd := exec.Command(exiftool, "-charset", "filename=utf8", "-@", tmpFile.Name(), "-json", "-DateTimeOriginal", "-Model", "-Make", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	out, err := cmd.Output()
	if err != nil {
		return exifError(errNoEXIF)
	}
	var result []exiftoolOutput
	if err := json.Unmarshal(out, &result); err != nil || len(result) == 0 {
		return exifError(errNoEXIF)
	}
	rec := result[0]

	makeStr := strings.TrimSpace(rec.Make)
	modelStr := strings.TrimSpace(rec.Model)

	if rec.DateTimeOriginal == "" && modelStr == "" && makeStr == "" {
		return exifError(errNoEXIF)
	}
	if rec.DateTimeOriginal == "" {
		return jalali{}, makeStr, modelStr, errNoDateTime
	}
	tm, err := time.Parse("2006:01:02 15:04:05", rec.DateTimeOriginal)
	if err != nil {
		return jalali{}, makeStr, modelStr, errBadDateFmt
	}
	return timeToJalali(tm), makeStr, modelStr, nil
}

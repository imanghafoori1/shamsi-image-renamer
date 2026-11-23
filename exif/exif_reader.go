package exif

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	. "renamer/file"
	. "renamer/jalali"
	"strings"
	"syscall"
	"time"
)

var (
	errNoEXIF     = errors.New("no EXIF data")
	ErrNoDateTime = errors.New("no DateTimeOriginal")
	ErrBadDateFmt = errors.New("bad date format")
)

type exiftoolOutput struct {
	DateTimeOriginal string `json:"DateTimeOriginal"`
	Model            string `json:"Model"`
	Make             string `json:"Make"`
}

// returns (captureTime, Make, Model).
func ReadExifToolMeta(file FilePath) (Jalali, string, string, error) {
	exifError := func(errNoEXIFs error) (Jalali, string, string, error) {
		return Jalali{}, "", "", errNoEXIFs
	}

	path := file.FullPath()
	exiftool, err := EnsureExifTool()
	if err != nil {
		return exifError(errNoEXIF)
	}

	tmpFile, err := os.CreateTemp("", "exifargs-*.txt")
	if err != nil {
		return exifError(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the path as UTF-8, followed by a newline.
	if _, err := tmpFile.WriteString(path + "\n"); err != nil {
		return exifError(err)
	}
	tmpFile.Close()

	cmd := exec.Command(exiftool, "-charset", "filename=utf8", "-@", tmpFile.Name(), "-json", "-DateTimeOriginal", "-Model", "-Make")
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
		return Jalali{}, makeStr, modelStr, ErrNoDateTime
	}
	tm, err := time.Parse("2006:01:02 15:04:05", rec.DateTimeOriginal)
	if err != nil {
		return Jalali{}, makeStr, modelStr, ErrBadDateFmt
	}
	return TimeToJalali(tm), makeStr, modelStr, nil
}

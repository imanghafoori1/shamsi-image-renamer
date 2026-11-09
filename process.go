package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jalaali/go-jalaali"
)

var filenameDateRe = regexp.MustCompile(`(20[0-4]\d(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01]))`)

// exiftoolOutput is the subset of ExifTool's -json output we need.
type exiftoolOutput struct {
	DateTimeOriginal string `json:"DateTimeOriginal"`
	Model            string `json:"Model"`
	Make             string `json:"Make"`
}

func processDir(dir string, prependModel bool, format DateFormat, recursive bool, addTags bool) (*Report, error) {
	report := &Report{}
	if err := walkDir(dir, prependModel, format, recursive, addTags, report); err != nil {
		return report, err
	}
	return report, nil
}

func walkDir(dir string, prependModel bool, format DateFormat, recursive bool, addTags bool, report *Report) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if !recursive {
				continue
			}
			sub := filepath.Join(dir, entry.Name())
			if err := walkDir(sub, prependModel, format, true, addTags, report); err != nil {
				report.addSkipped(sub, "cannot read directory: "+err.Error())
			}
			continue
		}
		if !isSupported(entry.Name()) {
			continue
		}
		processFile(dir, entry.Name(), prependModel, format, addTags, report)
	}
	return nil
}

func isSupported(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".heic", ".heif", ".mp4", ".mov":
		return true
	}
	return false
}

func dateFromFilename(fileName string) (time.Time, error) {
	m := filenameDateRe.FindAllStringSubmatch(fileName, -1)
	if len(m) == 0 {
		return time.Time{}, errNoDateTime
	}
	dateStr := m[len(m)-1][1]
	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, errBadDateFmt
	}
	return t, nil
}

func processFile(dir, fileName string, prependModel bool, format DateFormat, addTags bool, report *Report) {
	fullPath := filepath.Join(dir, fileName)

	var tm time.Time
	var rawModel string
	var err error

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(fileName); ferr == nil {
			tm = ft
			rawModel = ""
			err = nil
		} else {
			tm, rawModel, err = readExifToolMeta(fullPath)
		}
	default:
		tm, rawModel, err = readExifToolMeta(fullPath)
	}
	if err != nil {
		report.addSkipped(fileName, err.Error())
		return
	}

	canonicalPrefix := ""
	if prependModel {
		canonicalPrefix = modelToPrefix(rawModel)
	}

	jy, jm, jd, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
	jmInt := int(jm)
	canonicalSuffix := format.Suffix(jy, jmInt, jd)

	fileExt := filepath.Ext(fileName)
	stem := strings.TrimSuffix(fileName, fileExt)

	stem = stripKnownDecorations(
		stem,
		buildPrefixStripSet(rawModel),
		buildSuffixStripSet(jy, jmInt, jd),
	)

	if stem == "" {
		report.addSkipped(fileName, "would become empty name")
		return
	}

	newName := canonicalPrefix + stem + canonicalSuffix + fileExt
	newPath := filepath.Join(dir, newName)

	if newName == fileName {
		report.addSkipped(fileName, "already renamed")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		report.addSkipped(fileName, "target already exists")
		return
	}

	if err := os.Rename(fullPath, newPath); err != nil {
		report.addSkipped(fileName, "rename failed")
		return
	}

	if addTags {
		if tagErr := addNewTags(newPath, jy, jmInt); tagErr != nil {
			fmt.Fprintf(os.Stderr, "WARNING: tags not added to %s: %v\n", newName, tagErr)
			report.addSkipped(newName, "renamed, but tags failed: "+tagErr.Error())
			return
		}
	}

	report.addRenamed(fileName, newName)
}

// readExifToolMeta extracts the capture date and camera model from any
// supported file using the bundled ExifTool.
func readExifToolMeta(path string) (time.Time, string, error) {
	exiftool, err := ensureExifTool()
	if err != nil {
		return time.Time{}, "", errNoEXIF
	}

	cmd := exec.Command(exiftool,
		"-json",
		"-DateTimeOriginal",
		"-Model",
		"-Make",
		path,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, "", errNoEXIF
	}

	var result []exiftoolOutput
	if err := json.Unmarshal(out, &result); err != nil || len(result) == 0 {
		return time.Time{}, "", errNoEXIF
	}

	rec := result[0]

	if rec.DateTimeOriginal == "" {
		return time.Time{}, "", errNoDateTime
	}
	tm, err := time.Parse("2006:01:02 15:04:05", rec.DateTimeOriginal)
	if err != nil {
		return time.Time{}, "", errBadDateFmt
	}

	model := strings.TrimSpace(rec.Model)
	if model == "" {
		model = strings.TrimSpace(rec.Make)
	}

	return tm, model, nil
}

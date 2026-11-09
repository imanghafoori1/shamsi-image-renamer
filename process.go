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

type exiftoolOutput struct {
	DateTimeOriginal string `json:"DateTimeOriginal"`
	Model            string `json:"Model"`
	Make             string `json:"Make"`
}

// processDir renames all supported media files in dir.
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
	var make, model string
	var err error

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(fileName); ferr == nil {
			tm = ft
			make, model = "", ""
			err = nil
		} else {
			tm, make, model, err = readExifToolMeta(fullPath)
		}
	default:
		tm, make, model, err = readExifToolMeta(fullPath)
	}
	if err != nil {
		report.addSkipped(fileName, err.Error())
		return
	}

	canonicalPrefix := ""
	if prependModel {
		canonicalPrefix = modelToPrefix(model)
	}

	jy, jm, jd, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
	jmInt := int(jm)
	canonicalSuffix := format.Suffix(jy, jmInt, jd)

	fileExt := filepath.Ext(fileName)
	stem := strings.TrimSuffix(fileName, fileExt)

	// Only strip model prefixes when the checkbox is on.
	var prefixes []string
	if prependModel {
		prefixes = buildPrefixStripSet(model)
	}
	stem = stripKnownDecorations(stem, prefixes, buildSuffixStripSet(jy, jmInt, jd))

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
		if tagErr := addNewTags(newPath, jy, jmInt, make, model); tagErr != nil {
			fmt.Fprintf(os.Stderr, "WARNING: tags not added to %s: %v\n", newName, tagErr)
			report.addTagFailed(newName, tagErr.Error())
			return
		}
		report.addTagged(newName)
	}

	report.addRenamed(fileName, newName)
}

// readExifToolMeta returns (captureTime, Make, Model).
func readExifToolMeta(path string) (time.Time, string, string, error) {
	exiftool, err := ensureExifTool()
	if err != nil {
		return time.Time{}, "", "", errNoEXIF
	}
	cmd := exec.Command(exiftool, "-json", "-DateTimeOriginal", "-Model", "-Make", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, "", "", errNoEXIF
	}
	var result []exiftoolOutput
	if err := json.Unmarshal(out, &result); err != nil || len(result) == 0 {
		return time.Time{}, "", "", errNoEXIF
	}
	rec := result[0]

	makeStr := strings.TrimSpace(rec.Make)
	modelStr := strings.TrimSpace(rec.Model)

	if rec.DateTimeOriginal == "" && modelStr == "" && makeStr == "" {
		return time.Time{}, "", "", errNoEXIF
	}
	if rec.DateTimeOriginal == "" {
		return time.Time{}, makeStr, modelStr, errNoDateTime
	}
	tm, err := time.Parse("2006:01:02 15:04:05", rec.DateTimeOriginal)
	if err != nil {
		return time.Time{}, makeStr, modelStr, errBadDateFmt
	}
	return tm, makeStr, modelStr, nil
}

// removeModelPrefixes renames files in dir by stripping a leading model
// prefix, if one is recognized for the file. Returns a report.
func removeModelPrefixes(dir string, recursive bool) (*Report, error) {
	report := &Report{}
	return report, walkRemove(dir, recursive, true, false, report)
}

// removeDateSuffixes renames files in dir by stripping a trailing date
// suffix, if one is recognized for the file.
func removeDateSuffixes(dir string, recursive bool) (*Report, error) {
	report := &Report{}
	return report, walkRemove(dir, recursive, false, true, report)
}

func walkRemove(dir string, recursive, stripPrefixes, stripSuffixes bool, report *Report) error {
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
			if err := walkRemove(sub, recursive, stripPrefixes, stripSuffixes, report); err != nil {
				report.addSkipped(sub, "cannot read directory: "+err.Error())
			}
			continue
		}
		if !isSupported(entry.Name()) {
			continue
		}
		removeFileDecorations(dir, entry.Name(), stripPrefixes, stripSuffixes, report)
	}
	return nil
}

func removeFileDecorations(dir, fileName string, stripPrefixes, stripSuffixes bool, report *Report) {
	fullPath := filepath.Join(dir, fileName)

	var tm time.Time
	var model string
	var err error

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(fileName); ferr == nil {
			tm = ft
			model = ""
			err = nil
		} else {
			tm, _, model, err = readExifToolMeta(fullPath)
		}
	default:
		tm, _, model, err = readExifToolMeta(fullPath)
	}
	if err != nil {
		report.addSkipped(fileName, err.Error())
		return
	}

	jy, jm, jd, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
	jmInt := int(jm)

	var prefixes, suffixes []string
	if stripPrefixes {
		prefixes = buildPrefixStripSet(model)
	}
	if stripSuffixes {
		suffixes = buildSuffixStripSet(jy, jmInt, jd)
	}

	fileExt := filepath.Ext(fileName)
	stem := strings.TrimSuffix(fileName, fileExt)
	newStem := stripKnownDecorations(stem, prefixes, suffixes)
	if newStem == stem {
		report.addSkipped(fileName, "nothing to remove")
		return
	}
	if newStem == "" {
		report.addSkipped(fileName, "would become empty name")
		return
	}
	newName := newStem + fileExt
	newPath := filepath.Join(dir, newName)
	if _, err := os.Stat(newPath); err == nil {
		report.addSkipped(fileName, "target already exists")
		return
	}
	if err := os.Rename(fullPath, newPath); err != nil {
		report.addSkipped(fileName, "rename failed")
		return
	}
	report.addRenamed(fileName, newName)
}

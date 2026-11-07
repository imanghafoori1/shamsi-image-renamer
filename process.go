package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jalaali/go-jalaali"
	"github.com/rwcarlsen/goexif/exif"
)

// processDir renames all JPG/JPEG files in dir according to the chosen
// options. If recursive is true, subdirectories are walked as well.
// Returns a Report describing what happened.
func processDir(dir string, prependModel bool, format DateFormat, recursive bool) (*Report, error) {
	report := &Report{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return report, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if recursive {
				sub := filepath.Join(dir, entry.Name())
				if err := walkDir(sub, prependModel, format, report); err != nil {
					// Record the unreadable subdir as a skip, but keep going.
					report.addSkipped(sub, "cannot read directory: "+err.Error())
				}
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		processFile(dir, entry.Name(), prependModel, format, report)
	}
	return report, nil
}

// walkDir recurses into subdir, processing JPG/JPEG files and descending
// further into any subdirectories it finds.
func walkDir(dir string, prependModel bool, format DateFormat, report *Report) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			sub := filepath.Join(dir, entry.Name())
			if err := walkDir(sub, prependModel, format, report); err != nil {
				report.addSkipped(sub, "cannot read directory: "+err.Error())
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		processFile(dir, entry.Name(), prependModel, format, report)
	}
	return nil
}

// processFile handles a single file and appends results to report.
func processFile(dir, fileName string, prependModel bool, format DateFormat, report *Report) {
	fullPath := filepath.Join(dir, fileName)
	f, err := os.Open(fullPath)
	if err != nil {
		report.addSkipped(fileName, "cannot open")
		return
	}

	x, err := exif.Decode(f)
	f.Close() // Close before renaming: Windows cannot rename an open file.
	if err != nil || x == nil {
		report.addSkipped(fileName, "no EXIF data")
		return
	}

	tm, err := readDateTaken(x)
	if err != nil {
		report.addSkipped(fileName, err.Error())
		return
	}

	rawModel, _ := readModel(x)

	canonicalPrefix := ""
	if prependModel {
		canonicalPrefix = modelToPrefix(rawModel)
	}

	jy, jm, jd, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
	canonicalSuffix := format.Suffix(jy, jm, jd)

	fileExt := filepath.Ext(fileName)
	stem := strings.TrimSuffix(fileName, fileExt)

	stem = stripKnownDecorations(
		stem,
		buildPrefixStripSet(rawModel),
		buildSuffixStripSet(jy, jm, jd),
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
	report.addRenamed(fileName, newName)
}

// readDateTaken parses EXIF DateTimeOriginal into a time.Time.
func readDateTaken(x *exif.Exif) (time.Time, error) {
	tag, err := x.Get(exif.DateTimeOriginal)
	if err != nil {
		return time.Time{}, errNoDateTime
	}
	dateStr, err := tag.StringVal()
	if err != nil {
		return time.Time{}, errBadDateTag
	}
	tm, err := time.Parse("2006:01:02 15:04:05", dateStr)
	if err != nil {
		return time.Time{}, errBadDateFmt
	}
	return tm, nil
}

// readModel returns the raw and sanitized camera model strings (both may be "").
func readModel(x *exif.Exif) (raw, clean string) {
	tag, err := x.Get(exif.Model)
	if err != nil {
		return "", ""
	}
	s, err := tag.StringVal()
	if err != nil {
		return "", ""
	}
	raw = strings.TrimSpace(s)
	clean = sanitizeModel(s)
	return raw, clean
}

// buildPrefixStripSet returns every model-prefix spelling a previous run
// could have prepended for this file, given its raw EXIF model. Includes
// spellings produced by earlier builds (pre-normalize, intermediate, raw).
func buildPrefixStripSet(rawModel string) []string {
	raw := strings.TrimSpace(rawModel)
	if raw == "" {
		return nil
	}

	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}

	// 1. Current-build spelling (respects the "NIKON D<digits>" preserve rule).
	add(modelToPrefix(raw))

	// 2. Legacy: fully sanitized, NOT normalized (pre-normalize builds).
	add(sanitizeModel(raw) + "_")

	// 3. Legacy: sanitized + normalized (intermediate builds).
	add(normalizeModel(sanitizeModel(raw)) + "_")

	// 4. Legacy: raw + "_" (older builds that kept spaces for other brands).
	add(raw + "_")

	return out
}

// buildSuffixStripSet returns every suffix spelling a previous run could
// have appended for this file's date, across all supported formats.
func buildSuffixStripSet(jy int, jm jalaali.Month, jd int) []string {
	out := make([]string, 0, len(dateFormats))
	for _, f := range dateFormats {
		out = append(out, f.Suffix(jy, jm, jd))
	}
	return out
}
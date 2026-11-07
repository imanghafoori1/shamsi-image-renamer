package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gomantics/imx"
	"github.com/jalaali/go-jalaali"
	"github.com/rwcarlsen/goexif/exif"
)

var filenameDateRe = regexp.MustCompile(`(20[0-4]\d(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01]))`)

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
	case ".jpg", ".jpeg":
		tm, rawModel, err = readJPEGMeta(fullPath)
	case ".heic", ".heif":
		tm, rawModel, err = readIMXMeta(fullPath)
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(fileName); ferr == nil {
			tm = ft
			rawModel = ""
			err = nil
		} else {
			tm, rawModel, err = readIMXMeta(fullPath)
		}
	default:
		report.addSkipped(fileName, "unsupported format")
		return
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

	if addTags {
		if tagErr := addBridgeTags(newPath, jy, jm); tagErr != nil {
			fmt.Fprintf(os.Stderr, "WARNING: tags not added to %s: %v\n", newName, tagErr)
			report.addSkipped(newName, "renamed, but tags failed: "+tagErr.Error())
			return
		}
	}

	report.addRenamed(fileName, newName)
}

// readJPEGMeta reads EXIF from JPEG files using goexif.
func readJPEGMeta(path string) (time.Time, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, "", errNoEXIF
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil || x == nil {
		return time.Time{}, "", errNoEXIF
	}

	tm, err := readJPEGDate(x)
	if err != nil {
		return time.Time{}, "", err
	}

	model := readJPEGModel(x)
	return tm, model, nil
}

func readJPEGDate(x *exif.Exif) (time.Time, error) {
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

func readJPEGModel(x *exif.Exif) string {
	tag, err := x.Get(exif.Model)
	if err != nil {
		return ""
	}
	s, err := tag.StringVal()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// readIMXMeta reads metadata from HEIC files using imx.
func readIMXMeta(path string) (time.Time, string, error) {
	meta, err := imx.MetadataFromFile(path)
	if err != nil || meta == nil {
		return time.Time{}, "", errNoEXIF
	}

	tm, err := readIMXDate(meta)
	if err != nil {
		return time.Time{}, "", err
	}

	model := readIMXModel(meta)
	return tm, model, nil
}

func readIMXDate(meta *imx.Metadata) (time.Time, error) {
	if tag, ok := meta.Tag(imx.TagDateTimeOriginal); ok {
		if s, ok := tag.Value.(string); ok && s != "" {
			if tm, err := time.Parse("2006:01:02 15:04:05", s); err == nil {
				return tm, nil
			}
		}
	}
	return time.Time{}, errNoDateTime
}

func readIMXModel(meta *imx.Metadata) string {
	if tag, ok := meta.Tag(imx.TagModel); ok {
		if s, ok := tag.Value.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	if tag, ok := meta.Tag(imx.TagMake); ok {
		if s, ok := tag.Value.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// buildPrefixStripSet returns every model-prefix spelling a previous run
// could have prepended for this file, given its raw EXIF model.
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

	add(modelToPrefix(raw))
	add(sanitizeModel(raw) + "_")
	add(normalizeModel(sanitizeModel(raw)) + "_")
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

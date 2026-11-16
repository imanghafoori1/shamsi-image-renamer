package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jalaali/go-jalaali"
)

var filenameDateRe = regexp.MustCompile(`(20[0-4]\d(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01]))`)

// ProgressFunc is called after every 10 processed files.
type ProgressFunc func(done, total int, currentName string)

type filePath struct {
	dir  string
	name string
}

func (p *filePath) fullPath() string {
	return filepath.Join(p.dir, p.name)
}

func (p *filePath) getExt() string {
	return strings.ToLower(filepath.Ext(p.name))
}

func processDir(options options, progress ProgressFunc) (*Report, error) {
	report := &Report{}

	total := countSupported(options.folders)
	done := 0

	onFile := func(path filePath) {
		processFile(path, options, report)
	}

	if err := walkDir(options.folders, report, progress, total, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

// countSupported walks folders once and returns the number of files that
// processFile would attempt to handle. Recursive only if recursive is true.
func countSupported(dirs folders) int {
	entries, err := os.ReadDir(dirs.rootDir)
	if err != nil {
		return 0
	}
	n := 0
	for _, entry := range entries {
		if entry.IsDir() {
			if dirs.recursive {
				n += countSupported(folders{
					filepath.Join(dirs.rootDir, entry.Name()),
					true,
				})
			}
			continue
		}
		if isSupported(entry.Name()) {
			n++
		}
	}
	return n
}

// FileFunc is called once per supported file. The caller decides what to do
// with the file (rename, tag, strip decorations, etc.).
type FileFunc func(file filePath)

func walkDir(dirs folders, report *Report, progress ProgressFunc, total int, done *int, onFile FileFunc) error {
	dir := dirs.rootDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !dirs.recursive {
				continue
			}
			sub := filepath.Join(dir, entry.Name())
			if err := walkDir(folders{sub, true}, report, progress, total, done, onFile); err != nil {
				report.addSkipped(sub, "cannot read directory: "+err.Error())
			}
			continue
		}
		if !isSupported(entry.Name()) {
			continue
		}
		onFile(filePath{dir, entry.Name()})
		*done++
		if progress != nil && *done%10 == 0 {
			progress(*done, total, entry.Name())
		}
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

func processFile(oldFile filePath, uiOptions options, report *Report) {
	prependModel := uiOptions.prependModel
	format := uiOptions.format
	addTags := uiOptions.addTags

	var tm time.Time
	var made, model string
	var err error

	ext := oldFile.getExt()
	fullPath := oldFile.fullPath()
	fileName := oldFile.name

	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(fileName); ferr == nil {
			tm = ft
			made, model = "", ""
			err = nil
		} else {
			tm, made, model, err = readExifToolMeta(fullPath)
		}
	default:
		tm, made, model, err = readExifToolMeta(fullPath)
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

	fileExt := oldFile.getExt()
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
	newFile := filePath{
		oldFile.dir,
		newName,
	}

	newPath := newFile.fullPath()

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
		if tagErr := addNewTags(newPath, jy, jmInt, made, model); tagErr != nil {
			report.addTagFailed(newName, tagErr.Error())
			return
		}
		report.addTagged(newName)
	}

	report.addRenamed(fileName, newName)
}

// removeModelPrefixes strips a recognized camera-model prefix from every
// supported file's stem, if one is present.
func removeModelPrefixes(dirs folders) (*Report, error) {
	return removeDecorations(dirs, true, false)
}

// removeDateSuffixes strips a recognized Jalali date suffix from every
// supported file's stem, if one is present.
func removeDateSuffixes(dirs folders) (*Report, error) {
	return removeDecorations(dirs, false, true)
}

// removeDecorations runs the walk once, invoking removeFileDecorations on
// every supported file with the given strip flags.
func removeDecorations(dirs folders, stripPrefixes, stripSuffixes bool) (*Report, error) {
	report := &Report{}
	onFile := func(path filePath) {
		removeFileDecorations(path, stripPrefixes, stripSuffixes, report)
	}
	var done int
	if err := walkDir(dirs, report, nil, 0, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

func removeFileDecorations(path filePath, stripPrefixes, stripSuffixes bool, report *Report) {
	var tm time.Time
	var model string
	var err error

	fullPath := path.fullPath()
	ext := path.getExt()
	fileName := path.name
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

	stem := strings.TrimSuffix(fileName, ext)
	newStem := stripKnownDecorations(stem, prefixes, suffixes)
	if newStem == stem {
		report.addSkipped(fileName, "nothing to remove")
		return
	}
	if newStem == "" {
		report.addSkipped(fileName, "would become empty name")
		return
	}
	newName := newStem + ext
	newPath := filepath.Join(path.dir, newName)
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

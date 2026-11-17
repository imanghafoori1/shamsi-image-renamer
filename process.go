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

func (file *filePath) fullPath() string {
	return filepath.Join(file.dir, file.name)
}

func (file *filePath) getExt() string {
	return strings.ToLower(filepath.Ext(file.name))
}

func (file *filePath) rename(newPath string) error {
	return os.Rename(file.fullPath(), newPath)
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
		file := filePath{name: entry.Name()}
		if file.isSupported() {
			n++
		}
	}
	return n
}

// FileFunc is called once per supported file. The caller decides what to do
// with the file (rename, tag, strip decorations, etc.).
type FileFunc func(file filePath)

type jalali struct {
	year  int
	month int
	day   int
}

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
			subDir := folders{sub, true}
			if err := walkDir(subDir, report, progress, total, done, onFile); err != nil {
				report.addSkipped(filePath{name: dir}, "cannot read directory: "+err.Error())
			}
			continue
		}

		file := filePath{dir, entry.Name()}
		if !file.isSupported() {
			continue
		}
		onFile(file)
		*done++
		if progress != nil && *done%10 == 0 {
			progress(*done, total, entry.Name())
		}
	}
	return nil
}

func (file *filePath) isSupported() bool {
	switch file.getExt() {
	case ".jpg", ".jpeg", ".heic", ".heif", ".mp4", ".mov":
		return true
	}
	return false
}

func (file *filePath) dateFromFilename() (jalali, error) {
	m := filenameDateRe.FindAllStringSubmatch(file.name, -1)
	if len(m) == 0 {
		return jalali{}, errNoDateTime
	}
	dateStr := m[len(m)-1][1]
	tm, err := time.Parse("20060102", dateStr)
	if err != nil {
		return jalali{}, errBadDateFmt
	}
	return timeToJalali(tm), nil
}

func processFile(oldFile filePath, uiOptions options, report *Report) {
	prependModel := uiOptions.prependModel
	format := uiOptions.format
	addTags := uiOptions.addTags

	var jala jalali
	var made, model string
	var err error

	switch oldFile.getExt() {
	case ".mp4", ".mov":
		if ft, ferr := oldFile.dateFromFilename(); ferr == nil {
			jala = ft
			made, model = "", ""
			err = nil
		} else {
			jala, made, model, err = oldFile.readExifToolMeta()
		}
	default:
		jala, made, model, err = oldFile.readExifToolMeta()
	}
	if err != nil {
		report.addSkipped(oldFile, err.Error())
		return
	}

	var canonicalPrefix string
	if prependModel {
		canonicalPrefix = modelToPrefix(model)
	} else {
		canonicalPrefix = ""
	}

	fileExt := oldFile.getExt()
	stem := strings.TrimSuffix(oldFile.name, fileExt)

	// Only strip model prefixes when the checkbox is on.
	var prefixes []string
	if prependModel {
		prefixes = buildPrefixStripSet(model)
	}
	stem = stripKnownDecorations(stem, prefixes, buildSuffixStripSet(jala))

	if stem == "" {
		report.addSkipped(oldFile, "would become empty name")
		return
	}

	newName := canonicalPrefix + stem + format.Suffix(jala) + fileExt
	newFile := filePath{
		oldFile.dir,
		newName,
	}

	newPath := newFile.fullPath()

	if newName == oldFile.name {
		report.addSkipped(oldFile, "already renamed")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		report.addSkipped(oldFile, "target already exists")
		return
	}

	if err := oldFile.rename(newPath); err != nil {
		report.addSkipped(oldFile, "rename failed")
		return
	}

	if addTags {
		if tagErr := newFile.addNewTags(jala, made, model); tagErr != nil {
			report.addTagFailed(oldFile, tagErr.Error())
			return
		}
		report.addTagged(newFile)
	}

	report.addRenamed(oldFile, newFile)
}

func timeToJalali(times time.Time) jalali {
	jy, jm, jd, _ := jalaali.ToJalaali(times.Year(), times.Month(), times.Day())
	return jalali{jy, int(jm), jd}
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

func removeFileDecorations(file filePath, stripPrefixes, stripSuffixes bool, report *Report) {
	var times jalali
	var model string
	var err error

	ext := file.getExt()
	fileName := file.name
	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := file.dateFromFilename(); ferr == nil {
			times = ft
			model = ""
			err = nil
		} else {
			times, _, model, err = file.readExifToolMeta()
		}
	default:
		times, _, model, err = file.readExifToolMeta()
	}
	if err != nil {
		report.addSkipped(file, err.Error())
		return
	}

	var prefixes, suffixes []string
	if stripPrefixes {
		prefixes = buildPrefixStripSet(model)
	}
	if stripSuffixes {
		suffixes = buildSuffixStripSet(times)
	}

	stem := strings.TrimSuffix(fileName, ext)
	newStem := stripKnownDecorations(stem, prefixes, suffixes)
	if newStem == stem {
		report.addSkipped(file, "nothing to remove")
		return
	}
	if newStem == "" {
		report.addSkipped(file, "would become empty name")
		return
	}
	newFile := filePath{file.dir, newStem + ext}
	if _, err := os.Stat(newFile.fullPath()); err == nil {
		report.addSkipped(file, "target already exists")
		return
	}

	if err := file.rename(newFile.fullPath()); err != nil {
		report.addSkipped(file, "rename failed")
		return
	}
	report.addRenamed(file, newFile)
}

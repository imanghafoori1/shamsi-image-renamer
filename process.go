package main

import (
	"os"
	"path/filepath"
	"regexp"
	. "renamer/file"
	. "renamer/jalali"
	"time"
)

var filenameDateRe = regexp.MustCompile(`(20[0-4]\d(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01]))`)

// ProgressFunc is called after every 10 processed files.
type ProgressFunc func(done, total int, currentName string)

func processDir(options options, progress ProgressFunc) (*Report, error) {
	report := &Report{}

	total := countSupported(options.folders)
	done := 0

	onFile := func(path FilePath) {
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
		file := FilePath{Name: entry.Name()}
		if file.IsSupported() {
			n++
		}
	}
	return n
}

// FileFunc is called once per supported file. The caller decides what to do
// with the file (rename, tag, strip decorations, etc.).
type FileFunc func(file FilePath)

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
				report.addSkipped(FilePath{Name: dir}, "cannot read directory: "+err.Error())
			}
			continue
		}

		file := FilePath{Dir: dir, Name: entry.Name()}
		if !file.IsSupported() {
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

func dateFromFilename(file FilePath) (Jalali, error) {
	m := filenameDateRe.FindAllStringSubmatch(file.Name, -1)
	if len(m) == 0 {
		return Jalali{}, errNoDateTime
	}
	dateStr := m[len(m)-1][1]
	tm, err := time.Parse("20060102", dateStr)
	if err != nil {
		return Jalali{}, errBadDateFmt
	}
	return TimeToJalali(tm), nil
}

func processFile(oldFile FilePath, uiOptions options, report *Report) {
	prependModel := uiOptions.prependModel
	format := uiOptions.format
	addTags := uiOptions.addTags

	var jala Jalali
	var made, model string
	var err error

	switch oldFile.GetExt() {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(oldFile); ferr == nil {
			jala = ft
			made, model = "", ""
			err = nil
		} else {
			jala, made, model, err = readExifToolMeta(oldFile)
		}
	default:
		jala, made, model, err = readExifToolMeta(oldFile)
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

	// Only strip model prefixes when the checkbox is on.
	var prefixes []string
	if prependModel {
		prefixes = buildPrefixStripSet(model)
	}
	stem, _ := stripKnownDecorations(oldFile, prefixes, buildSuffixStripSet(jala))

	if stem == "" {
		report.addSkipped(oldFile, "would become empty name")
		return
	}

	newName := canonicalPrefix + stem + format.Suffix(jala) + oldFile.GetExt()
	newFile := FilePath{
		Dir:  oldFile.Dir,
		Name: newName,
	}

	newPath := newFile.FullPath()

	if newName == oldFile.Name {
		report.addSkipped(oldFile, "already renamed")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		report.addSkipped(oldFile, "target already exists")
		return
	}

	if err := oldFile.Rename(newPath); err != nil {
		report.addSkipped(oldFile, "rename failed")
		return
	}

	if addTags {
		if tagErr := addNewTags(newFile, jala, made, model); tagErr != nil {
			report.addTagFailed(oldFile, tagErr.Error())
			return
		}
		report.addTagged(newFile)
	}

	report.addRenamed(oldFile, newFile)
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
	onFile := func(path FilePath) {
		removeFileDecorations(path, stripPrefixes, stripSuffixes, report)
	}
	var done int
	if err := walkDir(dirs, report, nil, 0, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

func removeFileDecorations(file FilePath, stripPrefixes, stripSuffixes bool, report *Report) {
	var times Jalali
	var model string
	var err error

	ext := file.GetExt()
	switch ext {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(file); ferr == nil {
			times = ft
			model = ""
			err = nil
		} else {
			times, _, model, err = readExifToolMeta(file)
		}
	default:
		times, _, model, err = readExifToolMeta(file)
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

	newStem, change := stripKnownDecorations(file, prefixes, suffixes)
	if change {
		report.addSkipped(file, "nothing to remove")
		return
	}
	if newStem == "" {
		report.addSkipped(file, "would become empty name")
		return
	}
	newFile := FilePath{Dir: file.Dir, Name: newStem + ext}
	if _, err := os.Stat(newFile.FullPath()); err == nil {
		report.addSkipped(file, "target already exists")
		return
	}

	if err := file.Rename(newFile.FullPath()); err != nil {
		report.addSkipped(file, "rename failed")
		return
	}
	report.addRenamed(file, newFile)
}

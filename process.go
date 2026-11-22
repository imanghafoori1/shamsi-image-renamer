package main

import (
	"os"
	"path/filepath"
	"regexp"
	. "renamer/file"
	. "renamer/jalali"
	. "renamer/ui"
	"time"
)

var filenameDateRe = regexp.MustCompile(`(20[0-4]\d(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01]))`)

// ProgressFunc is called after every 10 processed files.

func processDir(ops Options, progress ProgressFunc) (*Report, error) {
	report := &Report{}

	total := countSupported(ops.Folders)
	done := 0

	onFile := func(path FilePath) {
		processFile(path, ops, report)
	}

	if err := walkDir(ops.Folders, report, progress, total, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

// countSupported walks folders once and returns the number of files that
// processFile would attempt to handle. Recursive only if recursive is true.
func countSupported(dirs Folders) int {
	entries, err := os.ReadDir(dirs.RootDir)
	if err != nil {
		return 0
	}
	n := 0
	for _, entry := range entries {
		if entry.IsDir() {
			if dirs.Recursive {
				n += countSupported(Folders{
					filepath.Join(dirs.RootDir, entry.Name()),
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

func walkDir(dirs Folders, report *Report, progress ProgressFunc, total int, done *int, onFile FileFunc) error {
	dir := dirs.RootDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !dirs.Recursive {
				continue
			}
			sub := filepath.Join(dir, entry.Name())
			subDir := Folders{sub, true}
			if err := walkDir(subDir, report, progress, total, done, onFile); err != nil {
				report.AddSkipped(FilePath{Name: dir}, "cannot read directory: "+err.Error())
			}
			continue
		}

		file := FilePath{Dir: dir, Name: entry.Name()}
		if !file.IsSupported() {
			continue
		}
		onFile(file)
		*done++
		if *done%10 == 0 {
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

func removeModelPrefixes(dirs Folders) (*Report, error) {
	return removeDecorations(dirs, true, false)
}

func removeDateSuffixes(dirs Folders) (*Report, error) {
	return removeDecorations(dirs, false, true)
}

// removeDecorations runs the walk once, invoking removeFileDecorations on
// every supported file with the given strip flags.
func removeDecorations(dirs Folders, stripPrefixes, stripSuffixes bool) (*Report, error) {
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

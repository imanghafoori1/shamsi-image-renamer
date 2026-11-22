package namers

import (
	"os"
	. "renamer/exif"
	. "renamer/file"
	. "renamer/jalali"
	. "renamer/report"
	. "renamer/ui"

	"strings"
)

func modelToPrefix(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	return s + "_"
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
	add(raw + "_")

	return out
}

// buildSuffixStripSet returns every suffix spelling a previous run could
// have appended for this file's date, across all supported formats.
func buildSuffixStripSet(jala Jalali) []string {
	out := make([]string, 0, len(DateFormats))
	for _, f := range DateFormats {
		out = append(out, f.Suffix(jala))
	}
	return out
}

func RemoveFileDecorations(file FilePath, stripPrefixes, stripSuffixes bool, report *Report) {
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
			times, _, model, err = ReadExifToolMeta(file)
		}
	default:
		times, _, model, err = ReadExifToolMeta(file)
	}
	if err != nil {
		report.AddSkipped(file, err.Error())
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
		report.AddSkipped(file, "nothing to remove")
		return
	}
	if newStem == "" {
		report.AddSkipped(file, "would become empty name")
		return
	}
	newFile := FilePath{Dir: file.Dir, Name: newStem + ext}
	if _, err := os.Stat(newFile.FullPath()); err == nil {
		report.AddSkipped(file, "target already exists")
		return
	}

	if err := file.Rename(newFile.FullPath()); err != nil {
		report.AddSkipped(file, "rename failed")
		return
	}
	report.AddRenamed(file, newFile)
}

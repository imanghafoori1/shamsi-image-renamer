package main

import (
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
func buildSuffixStripSet(jy int, jm int, jd int) []string {
	out := make([]string, 0, len(dateFormats))
	for _, f := range dateFormats {
		out = append(out, f.Suffix(jy, jm, jd))
	}
	return out
}

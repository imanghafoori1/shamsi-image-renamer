package main

import (
	"sort"
	"strings"
)

// Predicate reports whether target occurs at the chosen end of s.
type Predicate func(s, target string) bool

// Remover removes target from the chosen end of s and returns the result.
type Remover func(s, target string) string

// stripRepeatedly removes matching strings from one end of stem, repeating
// until nothing changes. match reports whether a candidate occurs at the
// chosen end; trim removes it. When breakOnMatch is true, the inner loop
// restarts from the first candidate after a successful strip, which gives
// longest-first ordering its intended effect.
func stripRepeatedly(stem string, candidates []string, match Predicate, trim Remover) string {
	for {
		changed := false
		for _, c := range candidates {
			if c == "" {
				continue
			}
			if match(stem, c) {
				stem = trim(stem, c)
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}

	return stem
}

// stripKnownDecorations removes decoration strings that are known to be
// associated with this file, based on its immutable EXIF metadata.
//
// Both prefixes and suffixes are stripped repeatedly from the respective
// ends, so historical double-appends collapse back to the original stem.
// Longest suffix first, so "-1398-bahman-6" wins over "-1398-bahman".
func stripKnownDecorations(stem string, prefixesToStrip, suffixesToStrip []string) string {
	stem = stripRepeatedly(stem, prefixesToStrip, strings.HasPrefix, strings.TrimPrefix)

	sortedSuffixes := append([]string(nil), suffixesToStrip...)
	sort.Slice(sortedSuffixes, func(i, j int) bool {
		return len(sortedSuffixes[i]) > len(sortedSuffixes[j])
	})
	stem = stripRepeatedly(stem, sortedSuffixes, strings.HasSuffix, strings.TrimSuffix)

	return stem
}

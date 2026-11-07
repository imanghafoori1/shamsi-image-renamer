package main

import (
	"sort"
	"strings"
)

// stripKnownDecorations removes decoration strings that are known to be
// associated with this file, based on its immutable EXIF metadata.
//
// Both prefixes and suffixes are stripped repeatedly from the respective
// ends, so historical double-appends collapse back to the original stem.
// Longest suffix first, so "-1398-bahman-6" wins over "-1398-bahman".
func stripKnownDecorations(stem string, prefixesToStrip, suffixesToStrip []string) string {
	// Strip any known prefix, in a loop.
	for {
		changed := false
		for _, p := range prefixesToStrip {
			if p != "" && strings.HasPrefix(stem, p) {
				stem = strings.TrimPrefix(stem, p)
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Sort suffixes longest-first so overlaps are handled correctly.
	sortedSuffixes := append([]string(nil), suffixesToStrip...)
	sort.Slice(sortedSuffixes, func(i, j int) bool {
		return len(sortedSuffixes[i]) > len(sortedSuffixes[j])
	})

	for {
		changed := false
		for _, s := range sortedSuffixes {
			if s != "" && strings.HasSuffix(stem, s) {
				stem = strings.TrimSuffix(stem, s)
				changed = true
				break // restart scan with longest suffix
			}
		}
		if !changed {
			break
		}
	}

	return stem
}
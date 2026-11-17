package main

import (
	"fmt"
	"strings"
)

type Report struct {
	Renamed   []string
	Skipped   []string
	Tagged    []string
	TagFailed []string
}

func (r *Report) addRenamed(file, newName filePath) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", file.name, newName.name))
}

func (r *Report) addSkipped(file filePath, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", file.name, reason))
}

func (r *Report) addTagged(name filePath) {
	r.Tagged = append(r.Tagged, name.fullPath())
}

func (r *Report) addTagFailed(file filePath, reason string) {
	r.TagFailed = append(r.TagFailed, fmt.Sprintf("%s  (%s)", file.name, reason))
}

func (s *uiState) updateReport(r *Report) {
	s.summaryLabel.SetText(fmt.Sprintf("Renamed: %d      Skipped: %d      Tagged: %d",
		len(r.Renamed), len(r.Skipped), len(r.Tagged)))

	if len(r.Renamed) == 0 {
		s.renamedBox.SetText("(none)")
	} else {
		s.renamedBox.SetText(strings.Join(r.Renamed, "\r\n"))
	}

	if len(r.Skipped) == 0 {
		s.skippedBox.SetText("(none)")
	} else {
		s.skippedBox.SetText(strings.Join(r.Skipped, "\r\n"))
	}

	if len(r.Tagged) == 0 {
		s.taggedBox.SetText("(none)")
	} else {
		s.taggedBox.SetText(strings.Join(r.Tagged, "\r\n"))
	}
}

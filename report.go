package main

import (
	"fmt"
	. "renamer/file"
	"strings"
)

type Report struct {
	Renamed   []string
	Skipped   []string
	Tagged    []string
	TagFailed []string
}

func (r *Report) addRenamed(file, newName FilePath) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", file.Name, newName.Name))
}

func (r *Report) addSkipped(file FilePath, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", file.Name, reason))
}

func (r *Report) addTagged(name FilePath) {
	r.Tagged = append(r.Tagged, name.FullPath())
}

func (r *Report) addTagFailed(file FilePath, reason string) {
	r.TagFailed = append(r.TagFailed, fmt.Sprintf("%s  (%s)", file.Name, reason))
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

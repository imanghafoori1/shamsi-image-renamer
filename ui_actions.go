package main

import (
	"fmt"
)

// runRename validates input, runs processDir, and refreshes the report.
func runRename(s *uiState) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError("Please enter a Path")
		return
	}

	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	report, err := processDir(
		dir,
		s.prependModelCheck.Checked(),
		dateFormats[idx],
		s.recursiveCheck.Checked(),
		s.addTagsCheck.Checked(),
	)
	if err != nil {
		s.statusLabel.SetText("Error: " + err.Error())
		return
	}
	s.updateReport(report)
	s.statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d, tagged %d.",
		len(report.Renamed), len(report.Skipped), len(report.Tagged)))
}

// runRemove is the shared handler for the two cleanup buttons.
func runRemove(s *uiState, stripPrefixes, stripSuffixes bool) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError("Please enter a Path")
		return
	}

	var report *Report
	var err error
	switch {
	case stripPrefixes:
		report, err = removeModelPrefixes(dir, s.recursiveCheck.Checked())
	case stripSuffixes:
		report, err = removeDateSuffixes(dir, s.recursiveCheck.Checked())
	}
	if err != nil {
		s.statusLabel.SetText("Error: " + err.Error())
		return
	}
	s.updateReport(report)
	s.statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d.",
		len(report.Renamed), len(report.Skipped)))
}

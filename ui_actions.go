package main

import (
	"fmt"

	. "github.com/lxn/walk/declarative"
)

// buildActionsRow returns the centered "Rename Files" button.
func buildActionsRow(s *uiState) Widget {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			HSpacer{},
			PushButton{
				Text:      "Rename Files",
				MinSize:   Size{Width: 180, Height: 36},
				MaxSize:   Size{Width: 180, Height: 36},
				Font:      Font{PointSize: 11, Bold: true},
				OnClicked: func() {
					runRename(s)
				},
			},
			HSpacer{},
		},
	}
}

// runRename validates input, calls processDir, and updates the report.
func runRename(s *uiState) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.statusLabel.SetText("Please enter a path.")
		return
	}

	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	report, err := processDir(dir, s.prependModelCheck.Checked(), dateFormats[idx])
	if err != nil {
		s.statusLabel.SetText("Error: " + err.Error())
		return
	}

	s.updateReport(report)
	s.statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d.",
		len(report.Renamed), len(report.Skipped)))
}
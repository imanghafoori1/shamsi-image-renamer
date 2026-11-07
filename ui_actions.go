package main

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func buildActionsRow(s *uiState) Widget {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			HSpacer{},
			PushButton{
				Text:    "Rename Files",
				MinSize: Size{Width: 180, Height: 36},
				MaxSize: Size{Width: 180, Height: 36},
				Font:    buttonFont,
				OnClicked: func() {
					runRename(s)
				},
			},
			HSpacer{},
		},
	}
}

func runRename(s *uiState) {
	dir := s.dirEdit.Text()
	if dir == "" {
		// Big, bold, red error.
		s.statusLabel.SetFont(s.errorFont)
		s.statusLabel.SetTextColor(walk.RGB(200, 0, 0))
		s.statusLabel.SetText("Please enter a Path")
		return
	}

	// Reset to normal style on any subsequent run.
	s.statusLabel.SetFont(s.normalFont)
	s.statusLabel.SetTextColor(walk.RGB(0, 0, 0))

	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	report, err := processDir(dir, s.prependModelCheck.Checked(), dateFormats[idx], s.recursiveCheck.Checked(), s.addTagsCheck.Checked())
	if err != nil {
		s.statusLabel.SetText("Error: " + err.Error())
		return
	}

	s.updateReport(report)
	s.statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d.",
		len(report.Renamed), len(report.Skipped)))
}

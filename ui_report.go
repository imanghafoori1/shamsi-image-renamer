package main

import (
	. "github.com/lxn/walk/declarative"
)

// buildReportSection returns the summary label plus the two scrollable
// list boxes for renamed and skipped files.
func buildReportSection(s *uiState) Widget {
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			Label{AssignTo: &s.summaryLabel, Text: "Renamed: 0      Skipped: 0", Font: bigFont},
			Label{Text: "Renamed files:", Font: bigFont},
			TextEdit{
				AssignTo: &s.renamedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				MinSize:  Size{Height: 150},
			},
			Label{Text: "Skipped files:", Font: bigFont},
			TextEdit{
				AssignTo: &s.skippedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				MinSize:  Size{Height: 150},
			},
		},
	}
}
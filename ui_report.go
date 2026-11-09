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
			Label{
				AssignTo: &s.summaryLabel,
				Text:     "Renamed: 0      Skipped: 0      Tagged: 0",
				Font:     headerFont,
			},
			Label{Text: "Renamed files:", Font: headerFont},
			TextEdit{
				AssignTo: &s.renamedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				Font:     listFont,
				MinSize:  Size{Height: 130},
			},
			Label{Text: "Tagged files:", Font: headerFont},
			TextEdit{AssignTo: &s.taggedBox, ReadOnly: true, VScroll: true, Text: "(none)", Font: listFont, MinSize: Size{Height: 130}},
			Label{Text: "Skipped files:", Font: headerFont},
			TextEdit{
				AssignTo: &s.skippedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				Font:     listFont,
				MinSize:  Size{Height: 130},
			},
		},
	}
}

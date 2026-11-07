package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var bigFont = Font{PointSize: 11, Bold: true}

// uiState holds the widget references that must be shared between the
// builders and their event callbacks.
type uiState struct {
	dirEdit            *walk.LineEdit
	statusLabel        *walk.Label
	summaryLabel       *walk.Label
	renamedBox         *walk.TextEdit
	skippedBox         *walk.TextEdit
	prependModelCheck  *walk.CheckBox
	formatCombo        *walk.ComboBox
}

// runUI builds and runs the main window. Blocks until the window closes.
func runUI() {
	s := &uiState{}

	MainWindow{
		Title:   "JPG Shamsi Renamer",
		MinSize: Size{Width: 760, Height: 700},
		Layout:  VBox{},
		Children: []Widget{
			buildPathRow(s),
			buildOptionsRow(s),
			buildActionsRow(s),
			Label{AssignTo: &s.statusLabel, Text: "Ready."},
			Label{Text: ""},
			buildReportSection(s),
		},
	}.Run()
}

// updateReport refreshes the inline report section from a Report.
func (s *uiState) updateReport(r *Report) {
	s.summaryLabel.SetText(fmt.Sprintf("Renamed: %d      Skipped: %d",
		len(r.Renamed), len(r.Skipped)))

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
}
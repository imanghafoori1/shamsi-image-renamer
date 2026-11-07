package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Font presets. Adjust PointSize here to scale the whole UI.
var (
	headerFont = Font{PointSize: 12, Bold: true} // section titles
	bodyFont   = Font{PointSize: 11}             // labels, inputs, checkbox
	buttonFont = Font{PointSize: 12, Bold: true} // primary action
	listFont   = Font{PointSize: 11}             // report text boxes
)

// uiState holds the widget references that must be shared between the
// builders and their event callbacks.
type uiState struct {
	dirEdit           *walk.LineEdit
	statusLabel       *walk.Label
	summaryLabel      *walk.Label
	renamedBox        *walk.TextEdit
	skippedBox        *walk.TextEdit
	prependModelCheck *walk.CheckBox
	recursiveCheck    *walk.CheckBox
	formatCombo       *walk.ComboBox
	addTagsCheck      *walk.CheckBox
	errorFont         *walk.Font
	normalFont        *walk.Font
}

// runUI builds and runs the main window. Blocks until the window closes.
func runUI() {
	s := &uiState{}

	var err error

	s.errorFont, err = walk.NewFont("Segoe UI", 12, walk.FontBold)
	if err != nil {
		panic(err) // font creation failing at startup is fatal
	}

	s.normalFont, err = walk.NewFont("Segoe UI", 11, 0)
	if err != nil {
		panic(err)
	}

	MainWindow{
		Title:   "JPG Shamsi Renamer",
		MinSize: Size{Width: 560, Height: 760},
		Layout:  VBox{},
		Children: []Widget{
			buildPathRow(s),
			buildOptionsRow(s),
			buildActionsRow(s),
			Label{AssignTo: &s.statusLabel, Text: "Ready.", Font: bodyFont},
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

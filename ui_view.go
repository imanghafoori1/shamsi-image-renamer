package main

import (
	"sync"

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
	MainWindow        *walk.MainWindow
	dirEdit           *walk.LineEdit
	statusLabel       *walk.Label
	summaryLabel      *walk.Label
	renamedBox        *walk.TextEdit
	skippedBox        *walk.TextEdit
	taggedBox         *walk.TextEdit
	prependModelCheck *walk.CheckBox
	recursiveCheck    *walk.CheckBox
	formatCombo       *walk.ComboBox
	addTagsCheck      *walk.CheckBox
	errorFont         *walk.Font
	normalFont        *walk.Font
	running           bool
	progressMu        sync.Mutex
	progressPending   bool
	progressDone      int
	progressTotal     int
	progressName      string
}

func (s *uiState) setProgress(done, total int, currentName string) {
	s.progressMu.Lock()
	s.progressDone = done
	s.progressTotal = total
	s.progressName = currentName
	if s.progressPending {
		s.progressMu.Unlock()
		return
	}
	s.progressPending = true
	s.progressMu.Unlock()

	go func() {
		s.MainWindow.Synchronize(func() {
			s.progressMu.Lock()
			s.progressPending = false
			s.progressMu.Unlock()
			s.statusLabel.SetText(progressMessage(s.progressDone, s.progressTotal, s.progressName))
		})
	}()
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
		Title:    "JPG Shamsi Renamer",
		MinSize:  Size{Width: 360, Height: 760},
		MaxSize:  Size{Width: 360, Height: 760},
		Layout:   VBox{},
		AssignTo: &s.MainWindow,
		Children: []Widget{
			buildPathRow(s),
			buildOptionsRow(s),
			buildActionsRow(s),
			Label{AssignTo: &s.statusLabel, Text: "Ready.", Font: bodyFont},
			Label{Text: ""},
			buildReportSection(s),
		},
	}.Create()

	s.MainWindow.Run()
}

func (s *uiState) showError(msg string) {
	s.statusLabel.SetFont(s.errorFont)
	s.statusLabel.SetTextColor(walk.RGB(200, 0, 0))
	s.statusLabel.SetText(msg)
}

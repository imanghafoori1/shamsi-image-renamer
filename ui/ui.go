package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// uiState holds the widget references that must be shared between the
// builders and their event callbacks.
type UiState struct {
	MainWindow        *walk.MainWindow
	dirEdit           *walk.LineEdit
	StatusLabel       *walk.Label
	summaryLabel      *walk.Label
	renamedBox        *walk.TextEdit
	skippedBox        *walk.TextEdit
	taggedBox         *walk.TextEdit
	prependModelCheck *walk.CheckBox
	recursiveCheck    *walk.CheckBox
	FormatCombo       *walk.ComboBox
	addTagsCheck      *walk.CheckBox
	ErrorFont         *walk.Font
	NormalFont        *walk.Font
	running           bool
	progressMu        sync.Mutex
	progressPending   bool
	progressDone      int
	progressTotal     int
	progressName      string
}

type actionHandler struct {
	uiState *UiState
	message func(*Report) string
}

var (
	headerFont = Font{PointSize: 12, Bold: true} // section titles
	bodyFont   = Font{PointSize: 11}             // labels, inputs, checkbox
	listFont   = Font{PointSize: 11}             // report text boxes
	buttonFont = Font{PointSize: 12, Bold: true} // primary action
)

func (s *UiState) setProgress(done, total int, currentName string) {
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
			s.StatusLabel.SetText(progressMessage(s.progressDone, s.progressTotal, s.progressName))
		})
	}()
}

func (s *UiState) showError(msg string) {
	s.StatusLabel.SetFont(s.ErrorFont)
	s.StatusLabel.SetTextColor(walk.RGB(200, 0, 0))
	s.StatusLabel.SetText(msg)
}

func (s *UiState) updateReport(r *Report) {
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

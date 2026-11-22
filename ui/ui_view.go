package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type Renamer func(Options, ProgressFunc) WorkFunc

// runUI builds and runs the main window. Blocks until the window closes.
func RunUI(removeModelPrefixes, removeDateSuffixes RemoveFunc, renamer Renamer) {
	s := &UiState{}

	var err error

	s.ErrorFont, err = walk.NewFont("Segoe UI", 12, walk.FontBold)
	if err != nil {
		panic(err) // font creation failing at startup is fatal
	}

	s.NormalFont, err = walk.NewFont("Segoe UI", 11, 0)
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
			buildOptionsRow(s, renamer),
			buildActionsRow(s, removeModelPrefixes, removeDateSuffixes),
			Label{AssignTo: &s.StatusLabel, Text: "Ready.", Font: bodyFont},
			Label{Text: ""},
			buildReportSection(s),
		},
	}.Create()

	s.MainWindow.Run()
}

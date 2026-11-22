package ui

import (
	. "renamer/report"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type renamerFunc func(Options, ProgressFunc) WorkFunc

func RunUI(renamer renamerFunc, removeModelPrefixes, removeDateSuffixes removeFunc) {
	s := &uiState{}

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
			s.buildPathRow(),
			s.buildOptionsRow(renamer),
			s.buildActionsRow(removeModelPrefixes, removeDateSuffixes),
			Label{AssignTo: &s.StatusLabel, Text: "Ready.", Font: bodyFont},
			Label{Text: ""},
			s.buildReportSection(),
		},
	}.Create()

	s.MainWindow.Run()
}

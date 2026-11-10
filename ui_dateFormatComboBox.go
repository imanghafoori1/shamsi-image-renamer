package main

import (
	"fmt"

	. "github.com/lxn/walk/declarative"
)

// DateFormat describes one selectable suffix style.
type DateFormat struct {
	Label  string
	Suffix func(jy int, jm int, jd int) string
}

// Formats offered in the GUI. Index 0 is the default.
var dateFormats = []DateFormat{
	{
		Label: "1398-11-28  (year-month-day)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%04d-%02d-%02d", jy, jm, jd)
		},
	},
	{
		Label: "1398-bahman-28  (year-monthname-day)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%d-%s-%d", jy, monthNameLatin(jm), jd)
		},
	},
	{
		Label: "1398-bahman  (year-monthname)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%d-%s", jy, monthNameLatin(jm))
		},
	},
	{
		Label: "1398-winter  (year-season, Latin)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%d-%s", jy, seasonNameLatin(jm))
		},
	},
	{
		Label: "زمستان-1398  (season-year, Persian)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%s-%d", seasonNamePersian(jm), jy)
		},
	},
	{
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(jy int, jm int, jd int) string {
			return fmt.Sprintf("-%s-%d", monthNamePersian(jm), jy)
		},
	},
}

func dateFormatComboBox(s *uiState) ComboBox {
	formatLabels := make([]string, len(dateFormats))
	for i, f := range dateFormats {
		formatLabels[i] = f.Label
	}

	return ComboBox{
		AssignTo:     &s.formatCombo,
		Model:        formatLabels,
		CurrentIndex: 0,
		Font:         bodyFont,
		MinSize:      Size{Width: 340},
	}
}

package main

import (
	"fmt"

	. "github.com/lxn/walk/declarative"
)

// DateFormat describes one selectable suffix style.
type DateFormat struct {
	Label  string
	Suffix func(jd jalali) string
}

// Formats offered in the GUI. Index 0 is the default.
var dateFormats = []DateFormat{
	{
		Label: "1398-11-28  (year-month-day)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%04d-%02d-%02d", dj.year, dj.month, dj.day)
		},
	},
	{
		Label: "1398-bahman-28  (year-monthname-day)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%d-%s-%d", dj.year, monthNameLatin(dj.month), dj.day)
		},
	},
	{
		Label: "1398-bahman  (year-monthname)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%d-%s", dj.year, monthNameLatin(dj.month))
		},
	},
	{
		Label: "1398-winter  (year-season, Latin)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%d-%s", dj.year, seasonNameLatin(dj.month))
		},
	},
	{
		Label: "زمستان-1398  (season-year, Persian)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%s-%d", seasonNamePersian(dj.month), dj.year)
		},
	},
	{
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(dj jalali) string {
			return fmt.Sprintf("-%s-%d", monthNamePersian(dj.month), dj.year)
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

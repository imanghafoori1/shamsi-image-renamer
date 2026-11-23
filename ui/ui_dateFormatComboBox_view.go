package ui

import (
	"fmt"

	. "renamer/jalali"

	. "github.com/lxn/walk/declarative"
)

// DateFormat describes one selectable suffix style.
type DateFormat struct {
	Label  string
	Suffix func(jd Jalali) string
}

// Formats offered in the GUI. Index 0 is the default.
var DateFormats = []DateFormat{
	{
		Label: "1398-11-28  (year-month-day)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%04d-%02d-%02d", dj.Year, dj.Month, dj.Day)
		},
	},
	{
		Label: "1398-bahman-28  (year-monthname-day)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%d-%s-%d", dj.Year, dj.MonthNameLatin(), dj.Day)
		},
	},
	{
		Label: "1398-bahman  (year-monthname)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%d-%s", dj.Year, dj.MonthNameLatin())
		},
	},
	{
		Label: "1398-winter  (year-season, Latin)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%d-%s", dj.Year, dj.SeasonNameLatin())
		},
	},
	{
		Label: "زمستان-1398  (season-year, Persian)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%s-%d", dj.SeasonNamePersian(), dj.Year)
		},
	},
	{
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(dj Jalali) string {
			return fmt.Sprintf("-%s-%d", dj.MonthNamePersian(), dj.Year)
		},
	},
}

func (s *uiState) dateFormatComboBox() ComboBox {
	formatLabels := make([]string, len(DateFormats))
	for i, f := range DateFormats {
		formatLabels[i] = f.Label
	}

	return ComboBox{
		AssignTo:     &s.FormatCombo,
		Model:        formatLabels,
		CurrentIndex: 0,
		Font:         bodyFont,
		MinSize:      Size{Width: 340},
	}
}

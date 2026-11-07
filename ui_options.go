package main

import (
	. "github.com/lxn/walk/declarative"
)

// buildOptionsRow returns the "Prepend camera model" checkbox and the
// date-format dropdown, on a single row with the label to the right.
func buildOptionsRow(s *uiState) Widget {
	formatLabels := make([]string, len(dateFormats))
	for i, f := range dateFormats {
		formatLabels[i] = f.Label
	}

	return Composite{
		Layout: HBox{},
		Children: []Widget{
			CheckBox{
				AssignTo: &s.prependModelCheck,
				Text:     "Prepend camera model",
			},
			HSpacer{},
			Label{Text: "Date format:"},
			ComboBox{
				AssignTo:     &s.formatCombo,
				Model:        formatLabels,
				CurrentIndex: 0,
				MinSize:      Size{Width: 300},
			},
		},
	}
}
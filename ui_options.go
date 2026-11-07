package main

import (
	. "github.com/lxn/walk/declarative"
)

// buildOptionsRow returns the option checkboxes and the date-format dropdown.
func buildOptionsRow(s *uiState) Widget {
	formatLabels := make([]string, len(dateFormats))
	for i, f := range dateFormats {
		formatLabels[i] = f.Label
	}

	return Composite{
		Layout: VBox{},
		Children: []Widget{
			CheckBox{
				AssignTo: &s.prependModelCheck,
				Text:     "Prepend camera model",
				Font:     bodyFont,
			},
			CheckBox{
				AssignTo: &s.recursiveCheck,
				Text:     "Include subdirectories",
				Font:     bodyFont,
			},
			CheckBox{
				AssignTo: &s.addTagsCheck,
				Text:     "Add searchable tags",
				Font:     bodyFont,
			},
			Label{Text: "Date format:", Font: bodyFont},
			ComboBox{
				AssignTo:     &s.formatCombo,
				Model:        formatLabels,
				CurrentIndex: 0,
				Font:         bodyFont,
				// Remove MinSize here; let the VBox give it the full width.
			},
			HSpacer{},
		},
		MaxSize: Size{Width: 340},
		MinSize: Size{Width: 200},
	}
}

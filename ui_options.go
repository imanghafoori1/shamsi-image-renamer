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
			Composite{
				Layout: HBox{},
				Children: []Widget{
					CheckBox{
						AssignTo: &s.prependModelCheck,
						Text:     "Prepend camera model",
						Font:     bodyFont,
					},
					CheckBox{
						AssignTo: &s.addTagsCheck,
						Text:     "Add searchable tags (Bridge)",
						Font:     bodyFont,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Date format:", Font: bodyFont},
					ComboBox{
						AssignTo:     &s.formatCombo,
						Model:        formatLabels,
						CurrentIndex: 0,
						Font:         bodyFont,
						MinSize:      Size{Width: 340},
					},
					PushButton{
						Text:    "Rename Files",
						MinSize: Size{Width: 200, Height: 42},
						MaxSize: Size{Width: 200, Height: 42},
						Font:    buttonFont,
						OnClicked: func() {
							runRename(s)
						},
					},
				},
			},
		},
	}
}

// buildActionsRow returns the primary Rename button on its own row,
// followed by a row of cleanup buttons.
func buildActionsRow(s *uiState) Widget {
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			// Row 1: primary action.
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:    "Remove model",
						Font:    bodyFont,
						MinSize: Size{Width: 220, Height: 32},
						OnClicked: func() {
							runRemove(s, true, false)
						},
					},
					PushButton{
						Text:    "Remove shamsi date",
						Font:    bodyFont,
						MinSize: Size{Width: 220, Height: 32},
						OnClicked: func() {
							runRemove(s, false, true)
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

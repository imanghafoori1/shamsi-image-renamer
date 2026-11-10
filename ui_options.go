package main

import (
	. "github.com/lxn/walk/declarative"
)

// buildOptionsRow returns the option checkboxes and the date-format dropdown.
func buildOptionsRow(s *uiState) Widget {
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
						Text:     "Add searchable tags",
						Font:     bodyFont,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Date format:", Font: bodyFont},
					dateFormatComboBox(s),
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

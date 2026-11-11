package main

import (
	. "github.com/lxn/walk/declarative"
)

// buildOptionsRow returns the option checkboxes and the date-format dropdown.
func buildOptionsRow(s *uiState) Widget {
	message := func(report *Report) string {
		return doneTaggedMessage(len(report.Renamed), len(report.Skipped), len(report.Tagged))
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
							runRename(s, message)
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
	message := func(report *Report) string {
		return doneMessage(len(report.Renamed), len(report.Skipped))
	}
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:    "Remove model",
						Font:    bodyFont,
						MinSize: Size{Width: 220, Height: 32},
						OnClicked: func() {
							runRemove(s, removeModelPrefixes, message)
						},
					},
					PushButton{
						Text:    "Remove shamsi date",
						Font:    bodyFont,
						MinSize: Size{Width: 220, Height: 32},
						OnClicked: func() {
							runRemove(s, removeDateSuffixes, message)
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

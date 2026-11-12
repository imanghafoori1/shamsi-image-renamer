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

// options is a snapshot of the UI controls relevant to running an operation.
// It must be read on the GUI thread; the values are plain data and safe to
// hand to a worker goroutine afterward.
type folders struct {
	rootDir   string
	recursive bool
}
type options struct {
	folders      folders
	prependModel bool
	addTags      bool
	format       DateFormat
}

func readFormState(s *uiState) options {
	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	return options{
		folders: folders{
			s.dirEdit.Text(),
			s.recursiveCheck.Checked(),
		},
		prependModel: s.prependModelCheck.Checked(),
		addTags:      s.addTagsCheck.Checked(),
		format:       dateFormats[idx],
	}
}

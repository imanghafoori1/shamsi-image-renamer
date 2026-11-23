package ui

import (
	. "renamer/dirWalker"
	. "renamer/report"

	. "github.com/lxn/walk/declarative"
)

type Options struct {
	Folders      Folders
	PrependModel bool
	AddTags      bool
	Format       DateFormat
}

// returns the option checkboxes and the date-format dropdown.
func (s *uiState) buildOptionsRow(renamer renamerFunc) Widget {
	message := func(report *Report) string {
		return doneTaggedMessage(len(report.Renamed), len(report.Skipped), len(report.Tagged))
	}
	h := actionHandler{
		s, message,
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
					s.dateFormatComboBox(),
					PushButton{
						Text:    "Rename Files",
						MinSize: Size{Width: 200, Height: 42},
						MaxSize: Size{Width: 200, Height: 42},
						Font:    buttonFont,
						OnClicked: func() {
							h.runRename(renamer)
						},
					},
				},
			},
		},
	}
}

// returns the primary Rename button on its own row,
// followed by a row of cleanup buttons.
func (s *uiState) buildActionsRow(removeModelPrefixes, removeDateSuffixes removeFunc) Widget {
	message := func(report *Report) string {
		return doneMessage(len(report.Renamed), len(report.Skipped))
	}
	h := actionHandler{
		s, message,
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
							h.runRemove(removeModelPrefixes)
						},
					},
					PushButton{
						Text:    "Remove shamsi date",
						Font:    bodyFont,
						MinSize: Size{Width: 220, Height: 32},
						OnClicked: func() {
							h.runRemove(removeDateSuffixes)
						},
					},
					HSpacer{},
				},
			},
		},
	}
}

func readFormState(s *uiState) Options {
	idx := s.FormatCombo.CurrentIndex()
	if idx < 0 || idx >= len(DateFormats) {
		idx = 0
	}

	return Options{
		Folders: Folders{
			s.dirEdit.Text(),
			s.recursiveCheck.Checked(),
		},
		PrependModel: s.prependModelCheck.Checked(),
		AddTags:      s.addTagsCheck.Checked(),
		Format:       DateFormats[idx],
	}
}

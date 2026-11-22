package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// BuildPathRow returns the "Directory Path:" label, the text input, and the
// Browse button. The input box is stored in s.dirEdit.
func buildPathRow(s *UiState) Widget {
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			Label{Text: "Directory Path:", Font: headerFont},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{
						AssignTo: &s.dirEdit,
						Font:     bodyFont,
						MinSize:  Size{Height: 30},
					},
					PushButton{
						Text:    "Browse...",
						Font:    bodyFont,
						MinSize: Size{Width: 100, Height: 30},
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Select Directory"
							dlg.InitialDirPath = s.dirEdit.Text()

							if ok, err := dlg.ShowBrowseFolder(nil); err != nil {
								s.StatusLabel.SetText("Error: " + err.Error())
							} else if !ok {
								s.StatusLabel.SetText("Browse cancelled.")
							} else {
								s.dirEdit.SetText(dlg.FilePath)
								s.StatusLabel.SetText("Ready.")
							}
						},
					},
					CheckBox{
						AssignTo: &s.recursiveCheck,
						Text:     "Include subdirectories",
						Font:     bodyFont,
					},
				},
			},
		},
	}
}

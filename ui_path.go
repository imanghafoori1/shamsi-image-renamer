package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// buildPathRow returns the "Directory Path:" label, the text input, and the
// Browse button. The input box is stored in s.dirEdit.
func buildPathRow(s *uiState) Widget {
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			Label{Text: "Directory Path:", Font: bigFont},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{AssignTo: &s.dirEdit},
					PushButton{
						Text:    "Browse...",
						MinSize: Size{Width: 90},
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Select Directory"
							dlg.InitialDirPath = s.dirEdit.Text()

							if ok, err := dlg.ShowBrowseFolder(nil); err != nil {
								s.statusLabel.SetText("Error: " + err.Error())
							} else if !ok {
								s.statusLabel.SetText("Browse cancelled.")
							} else {
								s.dirEdit.SetText(dlg.FilePath)
								s.statusLabel.SetText("Ready.")
							}
						},
					},
				},
			},
		},
	}
}
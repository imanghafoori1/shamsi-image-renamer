package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var bigFont = Font{PointSize: 11, Bold: true}

func main() {
	var dirEdit *walk.LineEdit
	var statusLabel *walk.Label
	var summaryLabel *walk.Label
	var renamedBox *walk.TextEdit
	var skippedBox *walk.TextEdit
	var prependModelCheck *walk.CheckBox
	var formatCombo *walk.ComboBox

	updateReport := func(r *Report) {
		summaryLabel.SetText(fmt.Sprintf("Renamed: %d      Skipped: %d",
			len(r.Renamed), len(r.Skipped)))

		if len(r.Renamed) == 0 {
			renamedBox.SetText("(none)")
		} else {
			renamedBox.SetText(strings.Join(r.Renamed, "\r\n"))
		}

		if len(r.Skipped) == 0 {
			skippedBox.SetText("(none)")
		} else {
			skippedBox.SetText(strings.Join(r.Skipped, "\r\n"))
		}
	}

	formatLabels := make([]string, len(dateFormats))
	for i, f := range dateFormats {
		formatLabels[i] = f.Label
	}

	MainWindow{
		Title:   "JPG Shamsi Renamer",
		MinSize: Size{Width: 760, Height: 700},
		Layout:  VBox{},
		Children: []Widget{
			Label{Text: "Directory Path:", Font: bigFont},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{AssignTo: &dirEdit},
					PushButton{
						Text:    "Browse...",
						MinSize: Size{Width: 90},
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Select Directory"
							dlg.InitialDirPath = dirEdit.Text()

							if ok, err := dlg.ShowBrowseFolder(nil); err != nil {
								statusLabel.SetText("Error: " + err.Error())
							} else if !ok {
								statusLabel.SetText("Browse cancelled.")
							} else {
								dirEdit.SetText(dlg.FilePath)
								statusLabel.SetText("Ready.")
							}
						},
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					CheckBox{
						AssignTo: &prependModelCheck,
						Text:     "Prepend camera model",
					},
					HSpacer{},
					Label{Text: "Date format:"},
					ComboBox{
						AssignTo:     &formatCombo,
						Model:        formatLabels,
						CurrentIndex: 0,
						MinSize:      Size{Width: 300},
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:      "Rename Files",
						MinSize:   Size{Width: 180, Height: 36},
						MaxSize:   Size{Width: 180, Height: 36},
						Font:      Font{PointSize: 11, Bold: true},
						OnClicked: func() {
							dir := dirEdit.Text()
							if dir == "" {
								statusLabel.SetText("Please enter a path.")
								return
							}
							idx := formatCombo.CurrentIndex()
							if idx < 0 || idx >= len(dateFormats) {
								idx = 0
							}
							report, err := processDir(dir, prependModelCheck.Checked(), dateFormats[idx])
							if err != nil {
								statusLabel.SetText("Error: " + err.Error())
								return
							}
							updateReport(report)
							statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d.",
								len(report.Renamed), len(report.Skipped)))
						},
					},
					HSpacer{},
				},
			},

			Label{AssignTo: &statusLabel, Text: "Ready."},
			Label{Text: ""},

			Label{AssignTo: &summaryLabel, Text: "Renamed: 0      Skipped: 0", Font: bigFont},
			Label{Text: "Renamed files:", Font: bigFont},
			TextEdit{
				AssignTo: &renamedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				MinSize:  Size{Height: 150},
			},
			Label{Text: "Skipped files:", Font: bigFont},
			TextEdit{
				AssignTo: &skippedBox,
				ReadOnly: true,
				VScroll:  true,
				Text:     "(none)",
				MinSize:  Size{Height: 150},
			},
		},
	}.Run()
}
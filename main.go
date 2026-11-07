package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jalaali/go-jalaali"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/rwcarlsen/goexif/exif"
)

// Report holds the outcome of a rename run.
type Report struct {
	Renamed []string // "old -> new"
	Skipped []string // "filename: reason"
}

func (r *Report) addRenamed(oldName, newName string) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", oldName, newName))
}

func (r *Report) addSkipped(name, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", name, reason))
}

func main() {
	var dirEdit *walk.LineEdit
	var statusLabel *walk.Label

	MainWindow{
		Title:   "JPG Shamsi Renamer",
		MinSize: Size{Width: 560, Height: 200},
		Layout:  VBox{},
		Children: []Widget{
			Label{Text: "Directory Path:"},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{AssignTo: &dirEdit},
					PushButton{
						Text: "Browse...",
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
			PushButton{
				Text: "Rename Files",
				OnClicked: func() {
					dir := dirEdit.Text()
					if dir == "" {
						statusLabel.SetText("Please enter a path.")
						return
					}
					report, err := processDir(dir)
					if err != nil {
						statusLabel.SetText("Error: " + err.Error())
						return
					}
					statusLabel.SetText(fmt.Sprintf("Done. Renamed %d, skipped %d.",
						len(report.Renamed), len(report.Skipped)))
					showReport(report)
				},
			},
			Label{AssignTo: &statusLabel, Text: "Ready."},
		},
	}.Run()
}
func processDir(dir string) (*Report, error) {
	report := &Report{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return report, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		f, err := os.Open(fullPath)
		if err != nil {
			report.addSkipped(entry.Name(), "cannot open")
			continue
		}

		x, err := exif.Decode(f)
		f.Close()
		if err != nil || x == nil {
			report.addSkipped(entry.Name(), "no EXIF data")
			continue
		}

		tag, err := x.Get(exif.DateTimeOriginal)
		if err != nil {
			report.addSkipped(entry.Name(), "no DateTimeOriginal")
			continue
		}

		dateStr, err := tag.StringVal()
		if err != nil {
			report.addSkipped(entry.Name(), "unreadable date tag")
			continue
		}

		tm, err := time.Parse("2006:01:02 15:04:05", dateStr)
		if err != nil {
			report.addSkipped(entry.Name(), "bad date format")
			continue
		}

		jy, jm, _, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
		monthName := getPersianMonth(jm)

		suffix := fmt.Sprintf("-%s-%d", monthName, jy)
		fileExt := filepath.Ext(entry.Name())
		nameWithoutExt := strings.TrimSuffix(entry.Name(), fileExt)

		// Already carries exactly this suffix -> leave it alone.
		if strings.HasSuffix(nameWithoutExt, suffix) {
			report.addSkipped(entry.Name(), "already renamed")
			continue
		}

		newName := nameWithoutExt + suffix + fileExt
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); err == nil {
			report.addSkipped(entry.Name(), "target already exists")
			continue
		}

		if err := os.Rename(fullPath, newPath); err != nil {
			report.addSkipped(entry.Name(), "rename failed")
			continue
		}
		report.addRenamed(entry.Name(), newName)
	}
	return report, nil
}
// showReport opens a modal window listing what was renamed and what was skipped.
func showReport(r *Report) {
	var w *walk.MainWindow

	summary := fmt.Sprintf("Renamed: %d    Skipped: %d", len(r.Renamed), len(r.Skipped))

	renamedText := strings.Join(r.Renamed, "\r\n")
	if renamedText == "" {
		renamedText = "(none)"
	}
	skippedText := strings.Join(r.Skipped, "\r\n")
	if skippedText == "" {
		skippedText = "(none)"
	}

	MainWindow{
		AssignTo: &w,
		Title:    "Rename Report",
		MinSize:  Size{Width: 720, Height: 520},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: summary},
			Label{Text: "Renamed files:"},
			TextEdit{
				ReadOnly: true,
				Text:     renamedText,
				VScroll:  true,
				MinSize:  Size{Height: 180},
			},
			Label{Text: "Skipped files:"},
			TextEdit{
				ReadOnly: true,
				Text:     skippedText,
				VScroll:  true,
				MinSize:  Size{Height: 180},
			},
			PushButton{
				Text: "Close",
				OnClicked: func() {
					w.Close()
				},
			},
		},
	}.Create()

	w.Run()
}

func getPersianMonth(m jalaali.Month) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}
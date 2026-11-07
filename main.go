package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jalaali/go-jalaali"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/rwcarlsen/goexif/exif"
)

// Report holds the outcome of a rename run.
type Report struct {
	Renamed []string
	Skipped []string
}

func (r *Report) addRenamed(oldName, newName string) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", oldName, newName))
}

func (r *Report) addSkipped(name, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", name, reason))
}

var bigFont = Font{PointSize: 11, Bold: true}

// DateFormat describes one selectable suffix style.
type DateFormat struct {
	Label  string
	Suffix func(jy int, jm jalaali.Month, jd int) string
}

// Formats offered in the GUI. Index 0 is the default.
var dateFormats = []DateFormat{
	{
		Label: "1398-11-28  (year-month-day)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%04d-%02d-%02d", jy, int(jm), jd)
		},
	},
	{
		Label: "1398-bahman-28  (year-monthname-day)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%d-%s-%d", jy, monthNameLatin(jm), jd)
		},
	},
	{
		Label: "1398-bahman  (year-monthname)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%d-%s", jy, monthNameLatin(jm))
		},
	},
	{
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s-%d", monthNamePersian(jm), jy)
		},
	},
	{
		Label: "بهمن  (monthname only, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s", monthNamePersian(jm))
		},
	},
}

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

// sanitizeModel removes characters that are unsafe or ugly in filenames.
func sanitizeModel(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		"\\", "", "/", "", ":", "", "*", "", "?", "", "\"", "",
		"<", "", ">", "", "|", "", " ", "", "\t", "", "\r", "", "\n", "",
	)
	return replacer.Replace(s)
}

// stripKnownDecorations removes decoration strings that are known to be
// associated with this file, based on its immutable EXIF metadata.
//
// prefixesToStrip: every model-prefix spelling a previous run could have used
// suffixesToStrip: every date-suffix spelling a previous run could have used
//                  (all supported formats, for this file's exact date)
//
// Both are stripped repeatedly from the respective ends, so historical
// double-appends collapse back to the original stem. Longest suffix first,
// so "-1398-bahman-6" wins over "-1398-bahman".
func stripKnownDecorations(stem string, prefixesToStrip, suffixesToStrip []string) string {
	// Strip any known prefix, in a loop.
	for {
		changed := false
		for _, p := range prefixesToStrip {
			if p != "" && strings.HasPrefix(stem, p) {
				stem = strings.TrimPrefix(stem, p)
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Sort suffixes longest-first so overlaps are handled correctly.
	sortedSuffixes := append([]string(nil), suffixesToStrip...)
	sort.Slice(sortedSuffixes, func(i, j int) bool {
		return len(sortedSuffixes[i]) > len(sortedSuffixes[j])
	})

	for {
		changed := false
		for _, s := range sortedSuffixes {
			if s != "" && strings.HasSuffix(stem, s) {
				stem = strings.TrimSuffix(stem, s)
				changed = true
				break // restart scan with longest suffix
			}
		}
		if !changed {
			break
		}
	}

	return stem
}

func processDir(dir string, prependModel bool, format DateFormat) (*Report, error) {
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

		// --- Immutable metadata: date taken ---
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

		// --- Immutable metadata: camera model (raw + sanitized) ---
		rawModel := ""
		cleanModel := ""
		if modelTag, err := x.Get(exif.Model); err == nil {
			if modelStr, err := modelTag.StringVal(); err == nil {
				rawModel = strings.TrimSpace(modelStr)
				cleanModel = sanitizeModel(modelStr)
			}
		}

		canonicalPrefix := ""
		if prependModel && cleanModel != "" {
			canonicalPrefix = cleanModel + "_"
		}

		jy, jm, jd, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
		canonicalSuffix := format.Suffix(jy, jm, jd)

		fileExt := filepath.Ext(entry.Name())
		stem := strings.TrimSuffix(entry.Name(), fileExt)

		// --- Prefixes a previous run could have prepended for THIS file ---
		var prefixesToStrip []string
		if cleanModel != "" {
			prefixesToStrip = append(prefixesToStrip, cleanModel+"_")
		}
		if rawModel != "" && rawModel != cleanModel {
			prefixesToStrip = append(prefixesToStrip, rawModel+"_")
		}

		// --- All suffix spellings for THIS file's date, across every format ---
		var suffixesToStrip []string
		for _, f := range dateFormats {
			suffixesToStrip = append(suffixesToStrip, f.Suffix(jy, jm, jd))
		}

		stem = stripKnownDecorations(stem, prefixesToStrip, suffixesToStrip)

		if stem == "" {
			report.addSkipped(entry.Name(), "would become empty name")
			continue
		}

		newName := canonicalPrefix + stem + canonicalSuffix + fileExt
		newPath := filepath.Join(dir, newName)

		if newName == entry.Name() {
			report.addSkipped(entry.Name(), "already renamed")
			continue
		}

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

func monthNameLatin(m jalaali.Month) string {
	months := []string{
		"farvardin", "ordibehesht", "khordad", "tir", "mordad", "shahrivar",
		"mehr", "aban", "azar", "dey", "bahman", "esfand",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}

func monthNamePersian(m jalaali.Month) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}
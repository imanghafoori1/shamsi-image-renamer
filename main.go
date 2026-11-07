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

func main() {
	var dirEdit *walk.LineEdit
	var statusLabel *walk.Label

	MainWindow{
		Title:   "JPG Shamsi Renamer",
		MinSize: Size{Width: 500, Height: 200},
		Layout:  VBox{},
		Children: []Widget{
			Label{Text: "Directory Path:"},
			LineEdit{AssignTo: &dirEdit},
			PushButton{
				Text: "Rename Files",
				OnClicked: func() {
					dir := dirEdit.Text()
					if dir == "" {
						statusLabel.SetText("Please enter a path.")
						return
					}
					count, err := processDir(dir)
					if err != nil {
						statusLabel.SetText("Error: " + err.Error())
					} else {
						statusLabel.SetText(fmt.Sprintf("Done! Renamed %d files.", count))
					}
				},
			},
			Label{AssignTo: &statusLabel, Text: "Ready."},
		},
	}.Run()
}

func processDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
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
			continue
		}

		x, err := exif.Decode(f)
		if err != nil || x == nil {
			f.Close()
			continue
		}


		// Get DateTimeOriginal string and parse
		tag, err := x.Get(exif.DateTimeOriginal)
		if err != nil {
			f.Close()
			continue
		}
		// ... after obtaining tag ...

		dateStr, err := tag.StringVal()
		if err != nil {
			f.Close()
			continue
		}
		tm, err := time.Parse("2006:01:02 15:04:05", dateStr)
		if err != nil {
			f.Close()
			continue
		}

		jy, jm, _, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
		monthName := getPersianMonth(jm) // pass jm directly

		// ... subsequent logic ...

		// Build new filename: IMG_0003-مهر-۱۴۰۴.jpg
		nameWithoutExt := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		newName := fmt.Sprintf("%s-%s-%d%s", nameWithoutExt, monthName, jy, filepath.Ext(entry.Name()))
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if err := os.Rename(fullPath, newPath); err == nil {
				count++
			}
		}
	}
	return count, nil
}

func getPersianMonth(m jalaali.Month) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if m >= 1 && m <= 12 {
		return months[m-1]
	}
	return ""
}
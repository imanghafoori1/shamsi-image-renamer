package test

import (
	"renamer/file"
	"testing"

	"renamer/report"
)

func TestReportAddMethods(t *testing.T) {
	r := &report.Report{}

	r.AddRenamed(file.FilePath{Name: "old.jpg"}, file.FilePath{Name: "new.jpg"})
	r.AddSkipped(file.FilePath{Name: "x.jpg"}, "no EXIF data")
	r.AddTagged(file.FilePath{Name: "a.jpg"})
	r.AddTagFailed(file.FilePath{Name: "b.jpg"}, "exiftool failed")

	if len(r.Renamed) != 1 || r.Renamed[0] != "old.jpg  ->  new.jpg" {
		t.Errorf("Renamed = %v", r.Renamed)
	}
	if len(r.Skipped) != 1 || r.Skipped[0] != "x.jpg  (no EXIF data)" {
		t.Errorf("Skipped = %v", r.Skipped)
	}
	if len(r.Tagged) != 1 || r.Tagged[0] != "a.jpg" {
		t.Errorf("Tagged = %v", r.Tagged)
	}
	if len(r.TagFailed) != 1 || r.TagFailed[0] != "b.jpg  (exiftool failed)" {
		t.Errorf("TagFailed = %v", r.TagFailed)
	}
}

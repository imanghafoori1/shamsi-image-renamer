package main

import "testing"

func TestReportAddRenamed(t *testing.T) {
	r := &Report{}
	r.addRenamed("old.jpg", "new.jpg")
	want := "old.jpg  ->  new.jpg"
	if len(r.Renamed) != 1 || r.Renamed[0] != want {
		t.Errorf("Renamed = %v, want [%q]", r.Renamed, want)
	}
}

func TestReportAddSkipped(t *testing.T) {
	r := &Report{}
	r.addSkipped("x.jpg", "no EXIF data")
	want := "x.jpg  (no EXIF data)"
	if len(r.Skipped) != 1 || r.Skipped[0] != want {
		t.Errorf("Skipped = %v, want [%q]", r.Skipped, want)
	}
}

func TestReportAddTagged(t *testing.T) {
	r := &Report{}
	r.addTagged("a.jpg")
	if len(r.Tagged) != 1 || r.Tagged[0] != "a.jpg" {
		t.Errorf("Tagged = %v, want [a.jpg]", r.Tagged)
	}
}

func TestReportAddTagFailed(t *testing.T) {
	r := &Report{}
	r.addTagFailed("a.jpg", "exiftool failed")
	want := "a.jpg  (exiftool failed)"
	if len(r.TagFailed) != 1 || r.TagFailed[0] != want {
		t.Errorf("TagFailed = %v, want [%q]", r.TagFailed, want)
	}
}

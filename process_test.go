package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// listDir returns the names of regular files inside dir.
func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}

func TestProcessDir_BasicRename(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "Canon EOS 5D")

	report, err := processDir(dir, false, dateFormats[2]) // year-monthname
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %d; skipped=%v", len(report.Renamed), report.Skipped)
	}

	files := listDir(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
	want := "IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_PrependModel(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "Canon EOS 5D")

	report, err := processDir(dir, true, dateFormats[2])
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %d; skipped=%v", len(report.Renamed), report.Skipped)
	}
	files := listDir(t, dir)
	want := "CanonEOS5D_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_NoEXIFIsSkipped(t *testing.T) {
	dir := t.TempDir()
	// A JPG with SOI+EOI but no APP1 → exif.Decode returns an error.
	writeJPEGWithEXIF(t, filepath.Join(dir, "plain.jpg"), "", "")

	report, err := processDir(dir, false, dateFormats[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 0 {
		t.Errorf("expected 0 renamed, got %v", report.Renamed)
	}
	if len(report.Skipped) != 1 {
		t.Fatalf("expected 1 skipped, got %v", report.Skipped)
	}
	if !strings.Contains(report.Skipped[0], "no EXIF data") {
		t.Errorf("skip reason = %q, want 'no EXIF data'", report.Skipped[0])
	}

	files := listDir(t, dir)
	if files[0] != "plain.jpg" {
		t.Errorf("file must remain untouched, got %q", files[0])
	}
}

func TestProcessDir_Idempotent(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "")

	r1, _ := processDir(dir, false, dateFormats[2])
	if len(r1.Renamed) != 1 {
		t.Fatalf("first run: expected 1 renamed, got %v; skipped=%v", r1.Renamed, r1.Skipped)
	}

	r2, _ := processDir(dir, false, dateFormats[2])
	if len(r2.Renamed) != 0 {
		t.Errorf("second run: expected 0 renamed, got %v", r2.Renamed)
	}
	if len(r2.Skipped) != 1 {
		t.Fatalf("second run: expected 1 skipped, got %v", r2.Skipped)
	}
	if !strings.Contains(r2.Skipped[0], "already renamed") {
		t.Errorf("skip reason = %q, want 'already renamed'", r2.Skipped[0])
	}
}

func TestProcessDir_FormatSwitchCleansOldSuffix(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "")

	// Run 1: year-monthname → IMG_0003-1398-bahman.jpg
	if r, _ := processDir(dir, false, dateFormats[2]); len(r.Renamed) != 1 {
		t.Fatalf("run1: %v", r.Skipped)
	}
	// Run 2: year-month-day → should REPLACE, not accumulate.
	if r, _ := processDir(dir, false, dateFormats[0]); len(r.Renamed) != 1 {
		t.Fatalf("run2: %v", r.Skipped)
	}

	files := listDir(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %v", files)
	}
	want := "IMG_0003-1398-11-06.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_ModelPrependIdempotent(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	// Run 1 and 2 with model ON — must not double the prefix.
	processDir(dir, true, dateFormats[2])
	r2, _ := processDir(dir, true, dateFormats[2])
	if len(r2.Renamed) != 0 {
		t.Errorf("second run should not rename, got %v", r2.Renamed)
	}

	files := listDir(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %v", files)
	}
	want := "NIKOND7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_ModelOffAfterOn(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	// Run 1 with model.
	processDir(dir, true, dateFormats[2])
	// Run 2 without model — prefix must be stripped.
	r2, _ := processDir(dir, false, dateFormats[2])
	if len(r2.Renamed) != 1 {
		t.Fatalf("expected rename on run2, got %v; skipped=%v", r2.Renamed, r2.Skipped)
	}

	files := listDir(t, dir)
	want := "IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_RepairsDoubleDecoratedName(t *testing.T) {
	dir := t.TempDir()
	// Simulate a file left behind by the old buggy build.
	writeJPEGWithEXIF(t,
		filepath.Join(dir, "NIKOND7200_NIKOND7200_IMG_0003-بهمن-1398-1398-bahman-6-1398-11-06.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := processDir(dir, true, dateFormats[2])
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}

	files := listDir(t, dir)
	want := "NIKOND7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_SkipsNonJPG(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0001.jpg"),
		"2020:01:26 14:30:00", "")

	report, err := processDir(dir, false, dateFormats[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v", report.Renamed)
	}
	for _, s := range report.Skipped {
		if strings.Contains(s, "notes.txt") {
			t.Errorf("non-JPG must not appear in skipped: %q", s)
		}
	}
	// notes.txt must still exist untouched.
	if _, err := os.Stat(filepath.Join(dir, "notes.txt")); err != nil {
		t.Errorf("notes.txt was touched: %v", err)
	}
}

func TestProcessDir_TargetExistsIsSkipped(t *testing.T) {
	dir := t.TempDir()

	// Source file with EXIF date 2020-01-26 -> 1398-11-06 (بهمن 1398).
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "")

	// Pre-create the exact target name, with the SAME date, so it is
	// recognised as "already renamed" and left untouched — keeping the
	// collision in place for IMG_0003.jpg.
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003-1398-bahman.jpg"),
		"2020:01:26 14:30:00", "")

	report, err := processDir(dir, false, dateFormats[2]) // year-monthname
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Renamed) != 0 {
		t.Errorf("expected 0 renamed, got %v", report.Renamed)
	}

	var sawCollision, sawAlready bool
	for _, s := range report.Skipped {
		if strings.Contains(s, "target already exists") {
			sawCollision = true
		}
		if strings.Contains(s, "already renamed") {
			sawAlready = true
		}
	}
	if !sawCollision {
		t.Errorf("expected 'target already exists' skip, got %v", report.Skipped)
	}
	if !sawAlready {
		t.Errorf("expected 'already renamed' skip, got %v", report.Skipped)
	}

	// Both files must still be present, unchanged.
	files := listDir(t, dir)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %v", files)
	}
}
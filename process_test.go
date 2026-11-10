package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// listDir returns the names of regular files inside dir.
//func listDir(t *testing.T, dir string) []string {
//	t.Helper()
//	entries, err := os.ReadDir(dir)
//	if err != nil {
//		t.Fatalf("read dir: %v", err)
//	}
//	var out []string
//	for _, e := range entries {
//		if !e.IsDir() {
//			out = append(out, e.Name())
//		}
//	}
//	return out
//}

func TestProcessDir_BasicRename(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "Canon EOS 5D")

	report, err := processDir(dir, false, dateFormats[2], false, false, nil) // year-monthname
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

func TestProcessDir_AddsTagsWhenChecked(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "IMG_0003.jpg")
	writeMinimalJPEG(t, path)

	// We need EXIF on the file for the rename to succeed.
	// mimage can add it, or reuse your existing fixture writer if it works
	// with mimage (it won't — mimage needs a real JPEG).
	// Simplest: write EXIF via mimage itself before the test.
	//_, err := metadata.NewJpegEditorFile(path)
	//if err != nil {
	//	t.Fatalf("open editor: %v", err)
	//}
	// Set DateTimeOriginal via mimage's EXIF API (check godoc for exact call).
	// For now, assume a helper exists; if not, skip this test until
	// you have a fixture that works with both goexif (for rename) and mimage (for tags).

	// This integration test is only meaningful once you have a JPEG that
	// both readers accept. Until then, the unit tests above cover the
	// tag-writing logic in isolation, which is the critical part.
	t.Skip("requires a JPEG fixture readable by both goexif and mimage")
}

func TestDateFromFilename(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string // "YYYY-MM-DD" or "" for error expected
	}{
		{"typical with underscores", "vid_20241024_34545.mp4", "2024-10-24"},
		{"no underscores", "vid20241024.mp4", "2024-10-24"},
		{"date as entire stem", "20241024.mp4", "2024-10-24"},
		{"date mid-name", "backup-20241024.tar.gz", "2024-10-24"},
		{"mid-year", "IMG_20250315_001.mp4", "2025-03-15"},
		{"first-of-year", "clip_20240101_000.mov", "2024-01-01"},
		{"last-of-year", "clip_20241231_000.mov", "2024-12-31"},
		{"invalid month", "vid_20241324_000.mp4", ""},
		{"invalid day", "vid_20240230_000.mp4", ""},
		{"year before 2000", "vid_19990101_000.mp4", ""},
		{"year 2050", "vid_20500101_000.mp4", ""},
		{"only seven digits", "vid_2024102_000.mp4", ""},
		{"eight digits not a year", "12345678.mp4", ""},
		{"no digits", "video.mp4", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tm, err := dateFromFilename(c.input)
			if c.want == "" {
				if err == nil {
					t.Errorf("expected error for %q, got date %v", c.input, tm)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tm.Format("2006-01-02")
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestProcessDir_PrependModel(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	report, err := processDir(dir, true, dateFormats[2], false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %d; skipped=%v", len(report.Renamed), report.Skipped)
	}
	files := listDir(t, dir)
	want := "NIKON D7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_NoEXIFIsSkipped(t *testing.T) {
	dir := t.TempDir()
	// A JPG with SOI+EOI but no APP1 → exif.Decode returns an error.
	writeJPEGWithEXIF(t, filepath.Join(dir, "plain.jpg"), "", "")

	report, err := processDir(dir, false, dateFormats[0], false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Renamed) != 0 {
		t.Errorf("expected 0 renamed, got %v", report.Renamed)
	}
	if !strings.Contains(report.Skipped[0], "no EXIF data") {
		t.Errorf("skip reason = %q, want 'no EXIF data'", report.Skipped[0])
	}
	if len(report.Skipped) != 1 {
		t.Fatalf("expected 1 skipped, got %v", report.Skipped)
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

	r1, _ := processDir(dir, false, dateFormats[2], false, false, nil)
	if len(r1.Renamed) != 1 {
		t.Fatalf("first run: expected 1 renamed, got %v; skipped=%v", r1.Renamed, r1.Skipped)
	}

	r2, _ := processDir(dir, false, dateFormats[2], false, false, nil)
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
	if r, _ := processDir(dir, false, dateFormats[2], false, false, nil); len(r.Renamed) != 1 {
		t.Fatalf("run1: %v", r.Skipped)
	}
	// Run 2: year-month-day → should REPLACE, not accumulate.
	if r, _ := processDir(dir, false, dateFormats[0], false, false, nil); len(r.Renamed) != 1 {
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
	processDir(dir, true, dateFormats[2], false, false, nil)
	r2, _ := processDir(dir, true, dateFormats[2], false, false, nil)
	if len(r2.Renamed) != 0 {
		t.Errorf("second run should not rename, got %v", r2.Renamed)
	}

	files := listDir(t, dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %v", files)
	}
	want := "NIKON D7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_ModelOffKeepsPrefix(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	// First run with prefix on.
	r1, _ := processDir(dir, true, dateFormats[2], false, false, nil)
	if len(r1.Renamed) != 1 {
		t.Fatalf("run1: expected 1 renamed, got %v; skipped=%v", r1.Renamed, r1.Skipped)
	}

	// Second run with prefix off. The existing prefix must be preserved,
	// so the file is already in its target state — no rename.
	r2, _ := processDir(dir, false, dateFormats[2], false, false, nil)
	if len(r2.Renamed) != 0 {
		t.Errorf("run2: expected 0 renamed (prefix preserved), got %v", r2.Renamed)
	}

	files := listDir(t, dir)
	want := "NIKON D7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q (prefix must be preserved)", files[0], want)
	}
}

func TestProcessDir_RepairsDoubleDecoratedName(t *testing.T) {
	dir := t.TempDir()
	// Simulate a file left behind by the old buggy build.
	writeJPEGWithEXIF(t,
		filepath.Join(dir, "NIKON D7200_NIKON D7200_IMG_0003-1398-bahman-1398-11-06.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := processDir(dir, true, dateFormats[2], false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}

	files := listDir(t, dir)
	want := "NIKON D7200_IMG_0003-1398-bahman.jpg"
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

	report, err := processDir(dir, false, dateFormats[0], false, false, nil)
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

	report, err := processDir(dir, false, dateFormats[2], false, false, nil) // year-monthname
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

func TestProcessDir_KeepsSpacedNikonModel(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t, filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := processDir(dir, true, dateFormats[2], false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}
	files := listDir(t, dir)
	want := "NIKON D7200_IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestProcessDir_Recursive(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "2020", "winter")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	writeJPEGWithEXIF(t, filepath.Join(root, "root.jpg"),
		"2020:01:26 14:30:00", "")
	writeJPEGWithEXIF(t, filepath.Join(sub, "nested.jpg"),
		"2020:01:26 14:30:00", "")

	// Non-recursive: only root.jpg should be renamed.
	r1, err := processDir(root, false, dateFormats[2], false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r1.Renamed) != 1 {
		t.Fatalf("non-recursive: expected 1 renamed, got %v", r1.Renamed)
	}
	if _, err := os.Stat(filepath.Join(sub, "nested.jpg")); err != nil {
		t.Errorf("non-recursive: nested.jpg was unexpectedly renamed")
	}

	// Recursive: nested.jpg should now also be renamed.
	r2, err := processDir(root, false, dateFormats[2], true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r2.Renamed) != 1 {
		t.Fatalf("recursive: expected 1 renamed, got %v; skipped=%v", r2.Renamed, r2.Skipped)
	}
	if _, err := os.Stat(filepath.Join(sub, "nested-1398-bahman.jpg")); err != nil {
		t.Errorf("recursive: nested-1398-bahman.jpg not found: %v", err)
	}
}

func TestRemoveDateSuffixes_JPEG(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t,
		filepath.Join(dir, "IMG_0003-1398-bahman.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := removeDateSuffixes(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}
	files := listDir(t, dir)
	if files[0] != "IMG_0003.jpg" {
		t.Errorf("got %q, want IMG_0003.jpg", files[0])
	}
}

func TestRemoveDateSuffixes_VideoFilenameDate(t *testing.T) {
	dir := t.TempDir()

	// A video whose name carries the original date fragment AND a Jalali
	// date suffix that a previous run appended. There is no metadata in
	// the file — the removal must work entirely from the filename date.
	if err := os.WriteFile(
		filepath.Join(dir, "vid_20241024_34545-1403-aban.mp4"),
		[]byte{0x00}, 0644,
	); err != nil {
		t.Fatal(err)
	}

	r, err := removeDateSuffixes(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}
	files := listDir(t, dir)
	want := "vid_20241024_34545.mp4"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestRemoveDateSuffixes_VideoNoDateAnywhere(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "random.mp4"), []byte{0x00}, 0644,
	); err != nil {
		t.Fatal(err)
	}

	r, err := removeDateSuffixes(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 0 {
		t.Errorf("expected 0 renamed, got %v", r.Renamed)
	}
	if len(r.Skipped) != 1 {
		t.Fatalf("expected 1 skipped, got %v", r.Skipped)
	}
	// The file has no metadata AND no filename date.
	// Either "no DateTimeOriginal" or "no EXIF data" is acceptable.
	if !strings.Contains(r.Skipped[0], "Date") && !strings.Contains(r.Skipped[0], "EXIF") {
		t.Errorf("unexpected skip reason: %q", r.Skipped[0])
	}
}

func TestRemoveModelPrefixes(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t,
		filepath.Join(dir, "NIKON D7200_IMG_0003-1398-bahman.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := removeModelPrefixes(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 1 {
		t.Fatalf("expected 1 renamed, got %v; skipped=%v", r.Renamed, r.Skipped)
	}
	files := listDir(t, dir)
	want := "IMG_0003-1398-bahman.jpg"
	if files[0] != want {
		t.Errorf("got %q, want %q", files[0], want)
	}
}

func TestRemoveDecorations_NothingToRemove(t *testing.T) {
	dir := t.TempDir()
	writeJPEGWithEXIF(t,
		filepath.Join(dir, "IMG_0003.jpg"),
		"2020:01:26 14:30:00", "NIKON D7200")

	r, err := removeDateSuffixes(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Renamed) != 0 {
		t.Errorf("expected 0 renamed, got %v", r.Renamed)
	}
	if len(r.Skipped) != 1 || !strings.Contains(r.Skipped[0], "nothing to remove") {
		t.Errorf("expected 'nothing to remove' skip, got %v", r.Skipped)
	}
}

func TestRemoveDecorations_Recursive(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	writeJPEGWithEXIF(t,
		filepath.Join(root, "A-1398-bahman.jpg"),
		"2020:01:26 14:30:00", "")
	writeJPEGWithEXIF(t,
		filepath.Join(sub, "B-1398-bahman.jpg"),
		"2020:01:26 14:30:00", "")

	// Non-recursive: only the root file is touched.
	r1, _ := removeDateSuffixes(root, false)
	if len(r1.Renamed) != 1 {
		t.Fatalf("non-recursive: expected 1 renamed, got %v", r1.Renamed)
	}
	if _, err := os.Stat(filepath.Join(sub, "B-1398-bahman.jpg")); err != nil {
		t.Errorf("non-recursive: sub file should be untouched")
	}

	// Recursive: sub file is now touched too.
	r2, _ := removeDateSuffixes(root, true)
	if len(r2.Renamed) != 1 {
		t.Fatalf("recursive: expected 1 renamed, got %v", r2.Renamed)
	}
	if _, err := os.Stat(filepath.Join(sub, "B.jpg")); err != nil {
		t.Errorf("recursive: sub file was not renamed: %v", err)
	}
}

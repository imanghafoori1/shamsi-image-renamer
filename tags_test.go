package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// readKeywords returns the XMP-dc:Subject list from a file by running the
// bundled ExifTool with short output and splitting on the delimiters that
// ExifTool uses for list values.
func readKeywords(t *testing.T, exiftool, path string) []string {
	t.Helper()

	out, err := exec.Command(exiftool, "-XMP-dc:Subject", "-s3", path).Output()
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	raw := strings.TrimSpace(string(out))
	raw = strings.ReplaceAll(raw, ".", "\n")
	raw = strings.ReplaceAll(raw, ",", "\n")
	raw = strings.ReplaceAll(raw, ";", "\n")

	var got []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			got = append(got, line)
		}
	}
	return got
}

func TestAddNewTags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	writeMinimalJPEG(t, path)

	jy, jm := 1398, 11
	if err := addNewTags(path, jy, jm, "NIKON CORPORATION", "NIKON D7200"); err != nil {
		t.Fatalf("addNewTags: %v", err)
	}

	exiftool, err := ensureExifTool()
	if err != nil {
		t.Fatalf("exiftool: %v", err)
	}

	got := readKeywords(t, exiftool, path)
	want := []string{"1398", "winter", "bahman", "NIKON CORPORATION", "NIKON D7200"}

	if len(got) != len(want) {
		t.Fatalf("keyword count = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("keyword[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAddNewTags_NoDayInTags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	writeMinimalJPEG(t, path)

	jy, jm := 1398, 11
	if err := addNewTags(path, jy, jm, "Canon", "Canon EOS 5D"); err != nil {
		t.Fatalf("addNewTags: %v", err)
	}

	exiftool, err := ensureExifTool()
	if err != nil {
		t.Fatalf("exiftool: %v", err)
	}

	for _, kw := range readKeywords(t, exiftool, path) {
		if strings.Contains(kw, "28") {
			t.Errorf("day must not appear in tags, found %q", kw)
		}
	}
}

func TestAddNewTags_EmptyMakeModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	writeMinimalJPEG(t, path)

	// Empty make and model must not produce empty keyword entries.
	jy, jm := 1398, 11
	if err := addNewTags(path, jy, jm, "", ""); err != nil {
		t.Fatalf("addNewTags: %v", err)
	}

	exiftool, err := ensureExifTool()
	if err != nil {
		t.Fatalf("exiftool: %v", err)
	}

	got := readKeywords(t, exiftool, path)
	want := []string{"1398", "winter", "bahman"}
	if len(got) != len(want) {
		t.Fatalf("keyword count = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("keyword[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMakeTags(t *testing.T) {
	cases := []struct {
		name        string
		jy, jm      int
		make, model string
		want        []string
	}{
		{"full", 1398, 11, "NIKON CORPORATION", "NIKON D7200",
			[]string{"1398", "winter", "bahman", "NIKON CORPORATION", "NIKON D7200"}},
		{"no model", 1398, 11, "Canon", "",
			[]string{"1398", "winter", "bahman", "Canon"}},
		{"no make/model", 1398, 11, "", "",
			[]string{"1398", "winter", "bahman"}},
		{"month 1 (spring)", 1399, 1, "", "",
			[]string{"1399", "spring", "farvardin"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := makeTags(c.jy, c.jm, c.make, c.model)
			if len(got) != len(c.want) {
				t.Fatalf("len = %d, want %d (got %v)", len(got), len(c.want), got)
			}
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Errorf("tag[%d] = %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestBuildArgs(t *testing.T) {
	got := buildArgs("C:\\Temp\\args.txt", []string{"1398", "winter", "bahman"})
	want := []string{
		"-charset", "filename=utf8",
		"-overwrite_original",
		"-XMP-dc:Subject=1398",
		"-XMP-dc:Subject+=winter",
		"-XMP-dc:Subject+=bahman",
		"-@", "C:\\Temp\\args.txt",
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

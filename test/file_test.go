package test

import (
	"os"
	"path/filepath"
	"testing"

	"renamer/file"
)

func TestFilePathFullPath(t *testing.T) {
	f := file.FilePath{Dir: `D:\test`, Name: "IMG_0003.JPG"}
	if got := f.FullPath(); got != `D:\test\IMG_0003.JPG` {
		t.Errorf("FullPath = %q", got)
	}
}

func TestFilePathGetExt(t *testing.T) {
	if got := (file.FilePath{Name: "IMG_0003.JPG"}).GetExt(); got != ".jpg" {
		t.Errorf("GetExt = %q, want .jpg", got)
	}
}

func TestFilePathIsSupported(t *testing.T) {
	for _, name := range []string{"a.jpg", "a.JPG", "a.jpeg", "a.HEIC", "a.heif", "a.mp4", "a.MOV"} {
		if !(file.FilePath{Name: name}).IsSupported() {
			t.Errorf("%q should be supported", name)
		}
	}
	for _, name := range []string{"a.txt", "a.png", "a", "archive.jpg.zip"} {
		if (file.FilePath{Name: name}).IsSupported() {
			t.Errorf("%q should not be supported", name)
		}
	}
}

func TestFilePathRename(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "old.txt")
	dst := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	f := file.FilePath{Dir: dir, Name: "old.txt"}
	if err := f.Rename(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("new file not found: %v", err)
	}
}

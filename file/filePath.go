package file

import (
	"os"
	"path/filepath"
	"strings"
)

type FilePath struct {
	Dir  string
	Name string
}

func (file *FilePath) FullPath() string {
	return filepath.Join(file.Dir, file.Name)
}

func (file *FilePath) GetExt() string {
	return strings.ToLower(filepath.Ext(file.Name))
}

func (file *FilePath) Rename(newPath string) error {
	return os.Rename(file.FullPath(), newPath)
}

func (file *FilePath) IsSupported() bool {
	switch file.GetExt() {
	case ".jpg", ".jpeg", ".heic", ".heif", ".mp4", ".mov":
		return true
	}
	return false
}

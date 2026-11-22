package dirWalker

import (
	"os"
	"path/filepath"
	. "renamer/file"
	. "renamer/report"
)

type Folders struct {
	RootDir   string
	Recursive bool
}

// FileFunc is called once per supported file. The caller decides what to do
// with the file (rename, tag, strip decorations, etc.).
type FileFunc func(file FilePath)

func WalkDir(dirs Folders, report *Report, progress ProgressFunc, total int, done *int, onFile FileFunc) error {
	dir := dirs.RootDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !dirs.Recursive {
				continue
			}
			sub := filepath.Join(dir, entry.Name())
			subDir := Folders{RootDir: sub, Recursive: true}
			if err := WalkDir(subDir, report, progress, total, done, onFile); err != nil {
				report.AddSkipped(FilePath{Name: dir}, "cannot read directory: "+err.Error())
			}
			continue
		}

		file := FilePath{Dir: dir, Name: entry.Name()}
		if !file.IsSupported() {
			continue
		}
		onFile(file)
		*done++
		if *done%10 == 0 {
			progress(*done, total, entry.Name())
		}
	}
	return nil
}

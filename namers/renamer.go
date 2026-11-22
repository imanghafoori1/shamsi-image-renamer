package namers

import (
	"os"
	"path/filepath"
	. "renamer/dirWalker"
	. "renamer/file"
	. "renamer/report"
	. "renamer/ui"
)

func Renamer(ops Options, Pr ProgressFunc) (work WorkFunc) {
	return func() (*Report, error) {
		return processDir(ops, Pr)
	}
}

func processDir(ops Options, progress ProgressFunc) (*Report, error) {
	report := &Report{}

	total := countSupported(ops.Folders)
	done := 0

	onFile := func(path FilePath) {
		processFile(path, ops, report)
	}

	if err := WalkDir(ops.Folders, report, progress, total, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

// countSupported walks folders once and returns the number of files that
// processFile would attempt to handle. Recursive only if recursive is true.
func countSupported(dirs Folders) int {
	entries, err := os.ReadDir(dirs.RootDir)
	if err != nil {
		return 0
	}
	n := 0
	for _, entry := range entries {
		if entry.IsDir() {
			if dirs.Recursive {
				n += countSupported(Folders{
					filepath.Join(dirs.RootDir, entry.Name()),
					true,
				})
			}
			continue
		}
		file := FilePath{Name: entry.Name()}
		if file.IsSupported() {
			n++
		}
	}
	return n
}

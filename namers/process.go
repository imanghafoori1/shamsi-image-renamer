package namers

import (
	. "renamer/dirWalker"
	. "renamer/file"
	. "renamer/report"
)

func RemoveModelPrefixes(dirs Folders) (*Report, error) {
	return removeDecorations(dirs, true, false)
}

func RemoveDateSuffixes(dirs Folders) (*Report, error) {
	return removeDecorations(dirs, false, true)
}

// removeDecorations runs the walk once, invoking removeFileDecorations on
// every supported file with the given strip flags.
func removeDecorations(dirs Folders, stripPrefixes, stripSuffixes bool) (*Report, error) {
	report := &Report{}
	onFile := func(path FilePath) {
		RemoveFileDecorations(path, stripPrefixes, stripSuffixes, report)
	}
	var done int
	if err := WalkDir(dirs, report, nil, 0, &done, onFile); err != nil {
		return report, err
	}
	return report, nil
}

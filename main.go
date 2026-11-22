package main

import . "renamer/ui"

func renamer(ops Options, Pr ProgressFunc) (work WorkFunc) {
	return func() (*Report, error) {
		return processDir(ops, Pr)
	}
}
func main() {
	defer cleanupExifTool()
	RunUI(removeModelPrefixes, removeDateSuffixes, renamer)
}

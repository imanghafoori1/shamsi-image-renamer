package main

import . "renamer/ui"
import . "renamer/exif"
import . "renamer/namers"

func main() {
	defer CleanupExifTool()
	RunUI(Renamer, RemoveModelPrefixes, RemoveDateSuffixes)
}

package ui

import "fmt"

func doneMessage(renameCount, skippedCount int) string {
	return fmt.Sprintf("Done. Renamed %d, skipped %d.", renameCount, skippedCount)
}

func doneTaggedMessage(renameCount, skippedCount, taggedCount int) string {
	return fmt.Sprintf("Done. Renamed %d, skipped %d, tagged %d.",
		renameCount, skippedCount, taggedCount)
}

func startingMessage() string {
	return "Starting..."
}

func errorPathMessage() string {
	return "Please enter a Path"
}

func errorPrefix(err string) string {
	return "Error: " + err
}

func progressMessage(done, total int, currentName string) string {
	return fmt.Sprintf("Processing %d/%d: %s", done, total, currentName)
}

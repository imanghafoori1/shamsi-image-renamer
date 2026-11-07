package main

import "fmt"

// Report holds the outcome of a rename run.
type Report struct {
	Renamed []string
	Skipped []string
}

func (r *Report) addRenamed(oldName, newName string) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", oldName, newName))
}

func (r *Report) addSkipped(name, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", name, reason))
}
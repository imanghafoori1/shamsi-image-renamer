package main

import "fmt"

type Report struct {
	Renamed   []string
	Skipped   []string
	Tagged    []string
	TagFailed []string
}

func (r *Report) addRenamed(oldName, newName string) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", oldName, newName))
}

func (r *Report) addSkipped(name, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", name, reason))
}

func (r *Report) addTagged(name string) {
	r.Tagged = append(r.Tagged, name)
}

func (r *Report) addTagFailed(name, reason string) {
	r.TagFailed = append(r.TagFailed, fmt.Sprintf("%s  (%s)", name, reason))
}

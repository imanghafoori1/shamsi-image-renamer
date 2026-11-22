package report

import (
	"fmt"
	. "renamer/file"
)

type ProgressFunc func(done, total int, currentName string)

type Report struct {
	Renamed   []string
	Skipped   []string
	Tagged    []string
	TagFailed []string
}

func (r *Report) AddRenamed(file, newName FilePath) {
	r.Renamed = append(r.Renamed, fmt.Sprintf("%s  ->  %s", file.Name, newName.Name))
}

func (r *Report) AddSkipped(file FilePath, reason string) {
	r.Skipped = append(r.Skipped, fmt.Sprintf("%s  (%s)", file.Name, reason))
}

func (r *Report) AddTagged(name FilePath) {
	r.Tagged = append(r.Tagged, name.FullPath())
}

func (r *Report) AddTagFailed(file FilePath, reason string) {
	r.TagFailed = append(r.TagFailed, fmt.Sprintf("%s  (%s)", file.Name, reason))
}

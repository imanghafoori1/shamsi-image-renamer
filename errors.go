package main

import "errors"

var (
	errNoEXIF     = errors.New("no EXIF data")
	errNoDateTime = errors.New("no DateTimeOriginal")
	errBadDateTag = errors.New("unreadable date tag")
	errBadDateFmt = errors.New("bad date format")
)

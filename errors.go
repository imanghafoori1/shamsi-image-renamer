package main

import "errors"

var (
	errNoEXIF     = errors.New("no EXIF data")
	errNoDateTime = errors.New("no DateTimeOriginal")
	errBadDateFmt = errors.New("bad date format")
)

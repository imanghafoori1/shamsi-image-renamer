package main

import "errors"

var (
	errNoDateTime = errors.New("no DateTimeOriginal")
	errBadDateTag = errors.New("unreadable date tag")
	errBadDateFmt = errors.New("bad date format")
)
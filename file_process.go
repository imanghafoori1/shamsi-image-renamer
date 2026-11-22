package main

import (
	"os"
	. "renamer/file"
	. "renamer/jalali"
	. "renamer/ui"
)

func processFile(oldFile FilePath, uiOptions Options, report *Report) {
	prependModel := uiOptions.PrependModel
	format := uiOptions.Format
	addTags := uiOptions.AddTags

	var jala Jalali
	var made, model string
	var err error

	switch oldFile.GetExt() {
	case ".mp4", ".mov":
		if ft, ferr := dateFromFilename(oldFile); ferr == nil {
			jala = ft
			made, model = "", ""
			err = nil
		} else {
			jala, made, model, err = readExifToolMeta(oldFile)
		}
	default:
		jala, made, model, err = readExifToolMeta(oldFile)
	}
	if err != nil {
		report.AddSkipped(oldFile, err.Error())
		return
	}

	var canonicalPrefix string
	if prependModel {
		canonicalPrefix = modelToPrefix(model)
	} else {
		canonicalPrefix = ""
	}

	// Only strip model prefixes when the checkbox is on.
	var prefixes []string
	if prependModel {
		prefixes = buildPrefixStripSet(model)
	}
	stem, _ := stripKnownDecorations(oldFile, prefixes, buildSuffixStripSet(jala))

	if stem == "" {
		report.AddSkipped(oldFile, "would become empty name")
		return
	}

	newName := canonicalPrefix + stem + format.Suffix(jala) + oldFile.GetExt()
	newFile := FilePath{
		Dir:  oldFile.Dir,
		Name: newName,
	}

	newPath := newFile.FullPath()

	if newName == oldFile.Name {
		report.AddSkipped(oldFile, "already renamed")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		report.AddSkipped(oldFile, "target already exists")
		return
	}

	if err := oldFile.Rename(newPath); err != nil {
		report.AddSkipped(oldFile, "rename failed")
		return
	}

	if addTags {
		if tagErr := addNewTags(newFile, jala, made, model); tagErr != nil {
			report.AddTagFailed(oldFile, tagErr.Error())
			return
		}
		report.AddTagged(newFile)
	}

	report.AddRenamed(oldFile, newFile)
}

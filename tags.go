package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// addNewTags writes year, season, and month as XMP keywords using the
// bundled ExifTool. Adobe Bridge reads these from XMP-dc:Subject.
func addNewTags(path string, jy int, jm int) error {
	exiftool, err := ensureExifTool()
	if err != nil {
		return fmt.Errorf("exiftool not available: %w", err)
	}

	tags := []string{
		fmt.Sprintf("%d", jy),
		seasonNameLatin(jm),
		monthNameLatin(jm),
	}

	// "#" after the tag name forces ExifTool to write the value as a string,
	// even when it looks numeric (e.g. "1398"). Without it, XMP ends up with
	// a mixed array of numbers and strings, which is invalid per the spec
	// and may be silently ignored by Adobe Bridge.
	args := []string{"-overwrite_original", "-XMP-dc:Subject#=" + tags[0]}
	for _, tag := range tags[1:] {
		args = append(args, "-XMP-dc:Subject#+="+tag)
	}
	args = append(args, path)

	log.Printf("exiftool: %s %v", exiftool, args)

	cmd := exec.Command(exiftool, args...)

	// --- THIS IS THE CRITICAL PART ---
	// Tell Windows to not create a visible console window for this process.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exiftool failed: %w (output: %s)", err, string(output))
	}
	return nil
}

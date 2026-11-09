package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// addNewTags writes year, season, month, camera make, and camera model as
// XMP keywords using the bundled ExifTool.
func addNewTags(path string, jy, jm int, cameraMake, cameraModel string) error {
	exiftool, err := ensureExifTool()
	if err != nil {
		return fmt.Errorf("exiftool not available: %w", err)
	}

	tags := []string{
		fmt.Sprintf("%d", jy),
		seasonNameLatin(jm),
		monthNameLatin(jm),
	}
	if cameraMake != "" {
		tags = append(tags, cameraMake)
	}
	if cameraModel != "" {
		tags = append(tags, cameraModel)
	}

	args := []string{"-overwrite_original", "-XMP-dc:Subject=" + tags[0]}
	for _, tag := range tags[1:] {
		args = append(args, "-XMP-dc:Subject+="+tag)
	}
	args = append(args, path)

	log.Printf("exiftool: %s %v", exiftool, args)

	cmd := exec.Command(exiftool, args...)
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

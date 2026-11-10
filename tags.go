package main

import (
	"fmt"
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

	tags := makeTags(jy, jm, cameraMake, cameraModel)
	args := buildArgs(path, tags)

	output, err := executeExifCommand(exiftool, args)
	if err != nil {
		return fmt.Errorf("exiftool failed: %w (output: %s)", err, string(output))
	}
	return nil
}

func makeTags(jy int, jm int, cameraMake string, cameraModel string) []string {
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
	return tags
}

func executeExifCommand(exiftool string, args []string) ([]byte, error) {
	cmd := exec.Command(exiftool, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	return cmd.CombinedOutput()
}

func buildArgs(path string, tags []string) []string {
	args := []string{"-overwrite_original", "-XMP-dc:Subject=" + tags[0]}
	for _, tag := range tags[1:] {
		args = append(args, "-XMP-dc:Subject+="+tag)
	}
	args = append(args, path)
	return args
}

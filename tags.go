package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// addNewTags writes year, season, month, camera make, and camera model as
// XMP keywords using the bundled ExifTool.
//
// The target file path is passed through an argument file (-@) with
// -charset filename=utf8. This is required on Windows: non-ASCII paths are
// mangled by the Win32 command-line encoding before ExifTool can see them,
// and ExifTool's own -charset flag cannot recover them once corrupted.
func addNewTags(path string, jy, jm int, cameraMake, cameraModel string) error {
	exiftool, err := ensureExifTool()
	if err != nil {
		return fmt.Errorf("exiftool not available: %w", err)
	}

	// Write the UTF-8 path to a temp file that ExifTool will read via -@.
	argFile, err := os.CreateTemp("", "exiftool-args-*.txt")
	if err != nil {
		return fmt.Errorf("create arg file: %w", err)
	}
	defer os.Remove(argFile.Name())

	if _, err := argFile.WriteString(path + "\n"); err != nil {
		argFile.Close()
		return fmt.Errorf("write arg file: %w", err)
	}
	if err := argFile.Close(); err != nil {
		return fmt.Errorf("close arg file: %w", err)
	}

	tags := makeTags(jy, jm, cameraMake, cameraModel)
	args := buildArgs(argFile.Name(), tags)

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

// buildArgs constructs the ExifTool command line.
//
// argFilePath is the path to a file containing the target filename in UTF-8;
// it is passed via -@. The tags are passed as command-line arguments because
// they are ASCII (year, season, month names, and camera strings), and they
// do not suffer from the same encoding problem as the path.
func buildArgs(argFilePath string, tags []string) []string {
	args := []string{
		"-charset", "filename=utf8",
		"-overwrite_original",
		"-XMP-dc:Subject=" + tags[0],
	}
	for _, tag := range tags[1:] {
		args = append(args, "-XMP-dc:Subject+="+tag)
	}
	args = append(args, "-@", argFilePath)
	return args
}

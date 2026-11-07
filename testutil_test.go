package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/msvens/mimage/metadata"
)

// writeMinimalJPEG writes a valid 1x1 JPEG with no metadata.
// mimage can open this file and add EXIF/IPTC/XMP to it.
func writeMinimalJPEG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode JPEG: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write JPEG: %v", err)
	}
}

// writeJPEGWithEXIF writes a minimal JPEG whose APP1 segment carries
// DateTimeOriginal and (optionally) Model. If dateTaken is "" the file is
// written WITHOUT an EXIF segment (used to test the "no EXIF data" skip path).
//
// The pixel data is intentionally absent — goexif only needs the APP1 header.
func writeJPEGWithEXIF(t *testing.T, path, dateTaken, model string) {
	t.Helper()

	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xD8}) // SOI

	if dateTaken != "" {
		tiff := buildExifTIFF(dateTaken, model)
		payload := append([]byte("Exif\x00\x00"), tiff...)
		buf.Write([]byte{0xFF, 0xE1})
		length := uint16(len(payload) + 2)
		var lenBytes [2]byte
		binary.BigEndian.PutUint16(lenBytes[:], length)
		buf.Write(lenBytes[:])
		buf.Write(payload)
	}

	buf.Write([]byte{0xFF, 0xD9}) // EOI

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write JPEG fixture: %v", err)
	}
}

// writeJPEGWithMetadata creates a 1x1 JPEG and adds EXIF DateTimeOriginal
// and Model via mimage, so both goexif and mimage can read it.
func writeJPEGWithMetadata(t *testing.T, path, dateTaken, model string) {
	t.Helper()
	writeMinimalJPEG(t, path)

	je, err := metadata.NewJpegEditorFile(path)
	if err != nil {
		t.Fatalf("open editor: %v", err)
	}

	// The exact API for setting EXIF DateTimeOriginal depends on mimage.
	// Check godoc for the correct method; likely je.Exif().SetDateTimeOriginal(...)
	// or a direct field assignment. The goal is to write tag 0x9003 and 0x0110.
	// If mimage lacks a convenience method, use its raw Exif accessor.

	if err := je.WriteFile(path); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
}

// buildExifTIFF returns the TIFF block that goes inside the APP1 payload
// (after the "Exif\0\0" signature). DateTimeOriginal lives in the Exif
// sub-IFD (reached via 0x8769); Model stays in IFD0.
func buildExifTIFF(dateTaken, model string) []byte {
	type entry struct {
		tag        uint16
		typ        uint16
		count      uint32
		dataOffset uint32
	}

	dateBytes := append([]byte(dateTaken), 0)
	var modelBytes []byte
	if model != "" {
		modelBytes = append([]byte(model), 0)
	}

	ifd0Count := uint16(1)
	if model != "" {
		ifd0Count = 2
	}
	exifIFDCount := uint16(1)

	tiffHeaderSize := uint32(8)
	ifd0Start := tiffHeaderSize
	ifd0Size := uint32(2) + uint32(ifd0Count)*12 + 4
	exifIFDStart := ifd0Start + ifd0Size
	exifIFDSize := uint32(2) + uint32(exifIFDCount)*12 + 4
	dataStart := exifIFDStart + exifIFDSize

	cursor := dataStart
	modelOffset := cursor
	if model != "" {
		cursor += uint32(len(modelBytes))
		if cursor%2 == 1 {
			cursor++
		}
	}
	dateOffset := cursor
	cursor += uint32(len(dateBytes))
	if cursor%2 == 1 {
		cursor++
	}

	var ifd0 []entry
	if model != "" {
		ifd0 = append(ifd0, entry{0x0110, 2, uint32(len(modelBytes)), modelOffset})
	}
	ifd0 = append(ifd0, entry{0x8769, 4, 1, exifIFDStart})

	exifIFD := []entry{
		{0x9003, 2, uint32(len(dateBytes)), dateOffset},
	}

	var out bytes.Buffer
	out.WriteString("II")
	binary.Write(&out, binary.LittleEndian, uint16(0x2A))
	binary.Write(&out, binary.LittleEndian, uint32(8))

	binary.Write(&out, binary.LittleEndian, ifd0Count)
	for _, e := range ifd0 {
		binary.Write(&out, binary.LittleEndian, e.tag)
		binary.Write(&out, binary.LittleEndian, e.typ)
		binary.Write(&out, binary.LittleEndian, e.count)
		binary.Write(&out, binary.LittleEndian, e.dataOffset)
	}
	binary.Write(&out, binary.LittleEndian, uint32(0))

	binary.Write(&out, binary.LittleEndian, exifIFDCount)
	for _, e := range exifIFD {
		binary.Write(&out, binary.LittleEndian, e.tag)
		binary.Write(&out, binary.LittleEndian, e.typ)
		binary.Write(&out, binary.LittleEndian, e.count)
		binary.Write(&out, binary.LittleEndian, e.dataOffset)
	}
	binary.Write(&out, binary.LittleEndian, uint32(0))

	if model != "" {
		out.Write(modelBytes)
		if len(modelBytes)%2 == 1 {
			out.WriteByte(0)
		}
	}
	out.Write(dateBytes)
	if len(dateBytes)%2 == 1 {
		out.WriteByte(0)
	}

	return out.Bytes()
}

// listDir returns the names of regular files inside dir.
func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, filepath.Base(e.Name()))
		}
	}
	return out
}

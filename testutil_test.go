package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

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
		// APP1 marker
		buf.Write([]byte{0xFF, 0xE1})
		// length includes the 2 length bytes themselves
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
// buildExifTIFF returns the TIFF block that goes inside the APP1 payload
// (after the "Exif\0\0" signature).
//
// Layout:
//   IFD0  contains:
//     - Model (0x0110, ASCII)
//     - ExifIFDPointer (0x8769, LONG) -> offset of Exif sub-IFD
//   ExifIFD contains:
//     - DateTimeOriginal (0x9003, ASCII)
func buildExifTIFF(dateTaken, model string) []byte {
	type entry struct {
		tag        uint16
		typ        uint16 // 2=ASCII, 4=LONG
		count      uint32
		dataOffset uint32 // for ASCII: offset into data area; for LONG: the value itself
	}

	dateBytes := append([]byte(dateTaken), 0) // NUL-terminated
	var modelBytes []byte
	if model != "" {
		modelBytes = append([]byte(model), 0)
	}

	// --- Plan layout ---
	// TIFF header: 8 bytes
	// IFD0: 2 + n0*12 + 4
	// ExifIFD: 2 + n1*12 + 4
	// Data area for ASCII values.

	ifd0Count := uint16(1) // ExifIFDPointer
	if model != "" {
		ifd0Count = 2
	}
	exifIFDCount := uint16(1) // DateTimeOriginal

	tiffHeaderSize := uint32(8)

	ifd0Start := tiffHeaderSize
	ifd0Size := uint32(2) + uint32(ifd0Count)*12 + 4

	exifIFDStart := ifd0Start + ifd0Size
	exifIFDSize := uint32(2) + uint32(exifIFDCount)*12 + 4

	dataStart := exifIFDStart + exifIFDSize

	// Assign offsets in the data area.
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

	// --- Build IFD0 entries (sorted by tag: 0x0110 < 0x8769) ---
	var ifd0 []entry
	if model != "" {
		ifd0 = append(ifd0, entry{
			tag:        0x0110,
			typ:        2,
			count:      uint32(len(modelBytes)),
			dataOffset: modelOffset,
		})
	}
	ifd0 = append(ifd0, entry{
		tag:        0x8769, // ExifIFDPointer
		typ:        4,      // LONG
		count:      1,
		dataOffset: exifIFDStart,
	})

	// --- Build Exif sub-IFD entries ---
	exifIFD := []entry{
		{
			tag:        0x9003, // DateTimeOriginal
			typ:        2,      // ASCII
			count:      uint32(len(dateBytes)),
			dataOffset: dateOffset,
		},
	}

	var out bytes.Buffer

	// TIFF header (little-endian)
	out.WriteString("II")
	binary.Write(&out, binary.LittleEndian, uint16(0x2A))
	binary.Write(&out, binary.LittleEndian, uint32(8))

	// IFD0
	binary.Write(&out, binary.LittleEndian, ifd0Count)
	for _, e := range ifd0 {
		binary.Write(&out, binary.LittleEndian, e.tag)
		binary.Write(&out, binary.LittleEndian, e.typ)
		binary.Write(&out, binary.LittleEndian, e.count)
		binary.Write(&out, binary.LittleEndian, e.dataOffset)
	}
	binary.Write(&out, binary.LittleEndian, uint32(0)) // next IFD = none

	// Exif sub-IFD
	binary.Write(&out, binary.LittleEndian, exifIFDCount)
	for _, e := range exifIFD {
		binary.Write(&out, binary.LittleEndian, e.tag)
		binary.Write(&out, binary.LittleEndian, e.typ)
		binary.Write(&out, binary.LittleEndian, e.count)
		binary.Write(&out, binary.LittleEndian, e.dataOffset)
	}
	binary.Write(&out, binary.LittleEndian, uint32(0)) // next IFD = none

	// Data area
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
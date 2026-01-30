package processor

import (
	"encoding/binary"
	"time"
)

// CreateEXIFSegment creates a complete EXIF APP1 segment payload
// Format: "Exif\0\0" + TIFF Header + IFD0 + ExifIFD + data values
func CreateEXIFSegment(dateTime time.Time) ([]byte, error) {
	byteOrder := binary.LittleEndian // Use little-endian (most common)

	// Format DateTimeOriginal string
	dateTimeStr := FormatDateTimeOriginal(dateTime)
	dateTimeBytes := []byte(dateTimeStr)

	// Calculate offsets
	// TIFF header: 8 bytes
	// IFD0: 2 (count) + entries*12 + 4 (next IFD offset)
	// ExifIFD: 2 (count) + entries*12 + 4 (next IFD offset)
	// Data values follow IFDs

	ifd0Offset := 8 // After TIFF header
	exifIFDOffset := ifd0Offset + 2 + 4*12 + 4 // IFD0: count + 4 entries + next offset
	dateTimeOffset := exifIFDOffset + 2 + 1*12 + 4 // ExifIFD: count + 1 entry + next offset

	// Create IFD0 entries
	// Entry 1: ImageWidth (placeholder - use 0)
	// Entry 2: ImageLength (placeholder - use 0)
	// Entry 3: Orientation (default 1)
	// Entry 4: ExifIFD pointer
	ifd0Entries := []TagEntry{
		{TagID: tagImageWidth, TagType: typeLong, Count: 1, Value: 0},
		{TagID: tagImageLength, TagType: typeLong, Count: 1, Value: 0},
		{TagID: tagOrientation, TagType: typeShort, Count: 1, Value: 1},
		{TagID: tagExifIFD, TagType: typeLong, Count: 1, Value: uint32(exifIFDOffset)},
	}

	// Create ExifIFD entries
	// Entry 1: DateTimeOriginal
	exifIFDEntries := []TagEntry{
		{TagID: tagDateTimeOriginal, TagType: typeASCII, Count: uint32(len(dateTimeBytes)), Value: uint32(dateTimeOffset)},
	}

	// Build IFD0
	ifd0 := CreateIFD(ifd0Entries, 0, byteOrder) // 0 = no next IFD

	// Build ExifIFD
	exifIFD := CreateIFD(exifIFDEntries, 0, byteOrder) // 0 = no next IFD

	// Create TIFF header
	tiffHeader := CreateTIFFHeader(byteOrder, uint32(ifd0Offset))

	// Assemble everything
	var buf []byte

	// EXIF identifier
	buf = append(buf, []byte("Exif\x00\x00")...)

	// TIFF header
	buf = append(buf, tiffHeader...)

	// IFD0
	buf = append(buf, ifd0...)

	// ExifIFD
	buf = append(buf, exifIFD...)

	// Data values (DateTimeOriginal string)
	buf = append(buf, dateTimeBytes...)

	return buf, nil
}

// CreateTIFFHeader creates an 8-byte TIFF header
// Format: [Byte Order (2)] [Magic (2)] [IFD Offset (4)]
func CreateTIFFHeader(byteOrder binary.ByteOrder, ifdOffset uint32) []byte {
	buf := make([]byte, 8)

	if byteOrder == binary.LittleEndian {
		// "II" (Intel - little-endian)
		buf[0] = 'I'
		buf[1] = 'I'
		// Magic number 42
		binary.LittleEndian.PutUint16(buf[2:4], 42)
		// IFD offset
		binary.LittleEndian.PutUint32(buf[4:8], ifdOffset)
	} else {
		// "MM" (Motorola - big-endian)
		buf[0] = 'M'
		buf[1] = 'M'
		// Magic number 42
		binary.BigEndian.PutUint16(buf[2:4], 42)
		// IFD offset
		binary.BigEndian.PutUint32(buf[4:8], ifdOffset)
	}

	return buf
}

// CreateEXIFSegmentWithDefaults creates EXIF segment with default values
// This is a convenience function that uses sensible defaults
func CreateEXIFSegmentWithDefaults(dateTime time.Time, imageWidth, imageLength uint32) ([]byte, error) {
	byteOrder := binary.LittleEndian

	// Format DateTimeOriginal string
	dateTimeStr := FormatDateTimeOriginal(dateTime)
	dateTimeBytes := []byte(dateTimeStr)

	// Calculate offsets
	ifd0Offset := 8
	exifIFDOffset := ifd0Offset + 2 + 4*12 + 4 // IFD0: count + 4 entries + next offset
	dateTimeOffset := exifIFDOffset + 2 + 1*12 + 4 // ExifIFD: count + 1 entry + next offset

	// Create IFD0 entries
	ifd0Entries := []TagEntry{
		{TagID: tagImageWidth, TagType: typeLong, Count: 1, Value: imageWidth},
		{TagID: tagImageLength, TagType: typeLong, Count: 1, Value: imageLength},
		{TagID: tagOrientation, TagType: typeShort, Count: 1, Value: 1},
		{TagID: tagExifIFD, TagType: typeLong, Count: 1, Value: uint32(exifIFDOffset)},
	}

	// Create ExifIFD entries
	exifIFDEntries := []TagEntry{
		{TagID: tagDateTimeOriginal, TagType: typeASCII, Count: uint32(len(dateTimeBytes)), Value: uint32(dateTimeOffset)},
	}

	// Build IFD0
	ifd0 := CreateIFD(ifd0Entries, 0, byteOrder)

	// Build ExifIFD
	exifIFD := CreateIFD(exifIFDEntries, 0, byteOrder)

	// Create TIFF header
	tiffHeader := CreateTIFFHeader(byteOrder, uint32(ifd0Offset))

	// Assemble everything
	var buf []byte

	// EXIF identifier
	buf = append(buf, []byte("Exif\x00\x00")...)

	// TIFF header
	buf = append(buf, tiffHeader...)

	// IFD0
	buf = append(buf, ifd0...)

	// ExifIFD
	buf = append(buf, exifIFD...)

	// Data values (DateTimeOriginal string)
	buf = append(buf, dateTimeBytes...)

	return buf, nil
}

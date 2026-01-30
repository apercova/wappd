package processor

import (
	"encoding/binary"
	"time"
)

const (
	// Tag IDs
	tagImageWidth      = 0x0100
	tagImageLength     = 0x0101
	tagOrientation     = 0x0112
	tagExifIFD         = 0x8769
	tagDateTimeOriginal = 0x9003
	tagDateTimeDigitized = 0x9004
	tagDateTime        = 0x0132

	// Tag Types
	typeByte   = 1
	typeASCII  = 2
	typeShort  = 3
	typeLong   = 4
	typeRational = 5
)

// TagEntry represents a 12-byte EXIF tag entry
type TagEntry struct {
	TagID   uint16
	TagType uint16
	Count   uint32
	Value   uint32 // Value if <= 4 bytes, or offset if > 4 bytes
}

// CreateTagEntry creates a 12-byte tag entry
func CreateTagEntry(tagID, tagType uint16, count, valueOrOffset uint32, byteOrder binary.ByteOrder) []byte {
	buf := make([]byte, 12)
	byteOrder.PutUint16(buf[0:2], tagID)
	byteOrder.PutUint16(buf[2:4], tagType)
	byteOrder.PutUint32(buf[4:8], count)
	byteOrder.PutUint32(buf[8:12], valueOrOffset)
	return buf
}

// FormatDateTimeOriginal formats a time.Time as EXIF DateTimeOriginal string
// Format: "YYYY:MM:DD HH:MM:SS\0" (20 bytes total: 19 chars + null terminator)
func FormatDateTimeOriginal(t time.Time) string {
	return t.Format("2006:01:02 15:04:05") + "\x00"
}

// CreateIFD creates an IFD (Image File Directory) structure
// Returns: [entry count (2)] + [entries (12*N)] + [next IFD offset (4)]
func CreateIFD(entries []TagEntry, nextIFDOffset uint32, byteOrder binary.ByteOrder) []byte {
	buf := make([]byte, 2+len(entries)*12+4)
	
	// Entry count
	byteOrder.PutUint16(buf[0:2], uint16(len(entries)))
	
	// Tag entries
	offset := 2
	for _, entry := range entries {
		entryBytes := CreateTagEntry(entry.TagID, entry.TagType, entry.Count, entry.Value, byteOrder)
		copy(buf[offset:offset+12], entryBytes)
		offset += 12
	}
	
	// Next IFD offset
	byteOrder.PutUint32(buf[offset:offset+4], nextIFDOffset)
	
	return buf
}

// PackString packs a string into bytes at a given offset
func PackString(s string, offset uint32, byteOrder binary.ByteOrder) ([]byte, uint32) {
	data := []byte(s)
	return data, offset + uint32(len(data))
}

// PackUint16 packs a uint16 into bytes
func PackUint16(value uint16, byteOrder binary.ByteOrder) []byte {
	buf := make([]byte, 2)
	byteOrder.PutUint16(buf, value)
	return buf
}

// PackUint32 packs a uint32 into bytes
func PackUint32(value uint32, byteOrder binary.ByteOrder) []byte {
	buf := make([]byte, 4)
	byteOrder.PutUint32(buf, value)
	return buf
}

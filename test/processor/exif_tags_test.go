package processor_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/apercova/wappd/internal/processor"
)

func TestFormatDateTimeOriginal(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Specific date",
			input:    time.Date(2025, 1, 22, 15, 30, 45, 0, time.UTC),
			expected: "2025:01:22 15:30:45\x00",
		},
		{
			name:     "Another date",
			input:    time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC),
			expected: "2024:04:15 12:00:00\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.FormatDateTimeOriginal(tt.input)
			if got != tt.expected {
				t.Errorf("FormatDateTimeOriginal() = %q, want %q", got, tt.expected)
			}
			// Verify it's exactly 20 bytes (19 chars + null terminator)
			if len(got) != 20 {
				t.Errorf("FormatDateTimeOriginal() length = %d, want 20", len(got))
			}
		})
	}
}

func TestCreateTagEntry(t *testing.T) {
	byteOrder := binary.LittleEndian
	tagID := uint16(0x9003)
	tagType := uint16(2) // typeASCII
	count := uint32(20)
	value := uint32(100)

	entry := processor.CreateTagEntry(tagID, tagType, count, value, byteOrder)

	if len(entry) != 12 {
		t.Errorf("CreateTagEntry() length = %d, want 12", len(entry))
	}

	// Verify tag ID
	gotTagID := byteOrder.Uint16(entry[0:2])
	if gotTagID != tagID {
		t.Errorf("CreateTagEntry() tagID = 0x%04x, want 0x%04x", gotTagID, tagID)
	}

	// Verify tag type
	gotTagType := byteOrder.Uint16(entry[2:4])
	if gotTagType != tagType {
		t.Errorf("CreateTagEntry() tagType = %d, want %d", gotTagType, tagType)
	}

	// Verify count
	gotCount := byteOrder.Uint32(entry[4:8])
	if gotCount != count {
		t.Errorf("CreateTagEntry() count = %d, want %d", gotCount, count)
	}

	// Verify value
	gotValue := byteOrder.Uint32(entry[8:12])
	if gotValue != value {
		t.Errorf("CreateTagEntry() value = %d, want %d", gotValue, value)
	}
}

func TestCreateIFD(t *testing.T) {
	byteOrder := binary.LittleEndian
	entries := []processor.TagEntry{
		{TagID: 0x0100, TagType: 4, Count: 1, Value: 1920}, // typeLong
		{TagID: 0x0101, TagType: 4, Count: 1, Value: 1080}, // typeLong
	}
	nextIFDOffset := uint32(0)

	ifd := processor.CreateIFD(entries, nextIFDOffset, byteOrder)

	// IFD should be: 2 bytes (count) + 12*2 bytes (entries) + 4 bytes (next offset) = 30 bytes
	expectedLength := 2 + len(entries)*12 + 4
	if len(ifd) != expectedLength {
		t.Errorf("CreateIFD() length = %d, want %d", len(ifd), expectedLength)
	}

	// Verify entry count
	gotCount := byteOrder.Uint16(ifd[0:2])
	if gotCount != uint16(len(entries)) {
		t.Errorf("CreateIFD() entry count = %d, want %d", gotCount, len(entries))
	}

	// Verify next IFD offset
	gotNextOffset := byteOrder.Uint32(ifd[len(ifd)-4:])
	if gotNextOffset != nextIFDOffset {
		t.Errorf("CreateIFD() next offset = %d, want %d", gotNextOffset, nextIFDOffset)
	}
}

func TestCreateTIFFHeader(t *testing.T) {
	byteOrder := binary.LittleEndian
	ifdOffset := uint32(8)

	header := processor.CreateTIFFHeader(byteOrder, ifdOffset)

	if len(header) != 8 {
		t.Errorf("CreateTIFFHeader() length = %d, want 8", len(header))
	}

	// Verify byte order marker
	if header[0] != 'I' || header[1] != 'I' {
		t.Errorf("CreateTIFFHeader() byte order = %c%c, want II", header[0], header[1])
	}

	// Verify magic number
	magic := byteOrder.Uint16(header[2:4])
	if magic != 42 {
		t.Errorf("CreateTIFFHeader() magic = %d, want 42", magic)
	}

	// Verify IFD offset
	gotOffset := byteOrder.Uint32(header[4:8])
	if gotOffset != ifdOffset {
		t.Errorf("CreateTIFFHeader() IFD offset = %d, want %d", gotOffset, ifdOffset)
	}
}

func TestCreateTIFFHeader_BigEndian(t *testing.T) {
	byteOrder := binary.BigEndian
	ifdOffset := uint32(8)

	header := processor.CreateTIFFHeader(byteOrder, ifdOffset)

	// Verify byte order marker
	if header[0] != 'M' || header[1] != 'M' {
		t.Errorf("CreateTIFFHeader() byte order = %c%c, want MM", header[0], header[1])
	}

	// Verify magic number
	magic := byteOrder.Uint16(header[2:4])
	if magic != 42 {
		t.Errorf("CreateTIFFHeader() magic = %d, want 42", magic)
	}
}

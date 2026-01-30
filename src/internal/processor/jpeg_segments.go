package processor

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	markerSOI = 0xD8 // Start of Image
	markerEOI = 0xD9 // End of Image
	markerAPP1 = 0xE1 // APP1 segment (EXIF)
	markerAPP0 = 0xE0 // APP0 segment (JFIF)
	markerSOF0 = 0xC0 // Start of Frame (baseline)
	markerSOF1 = 0xC1 // Start of Frame (extended)
	markerSOF2 = 0xC2 // Start of Frame (progressive)
	markerSOF3 = 0xC3 // Start of Frame (lossless)
)

// JPEGSegment represents a JPEG segment
type JPEGSegment struct {
	Marker  byte   // Marker type (0xE1 for APP1, etc.)
	Length  uint16 // Segment length (including length bytes)
	Payload []byte // Segment data (excluding marker and length)
}

// ParseJPEGSegments parses a JPEG file and extracts all segments
func ParseJPEGSegments(data []byte) ([]JPEGSegment, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("invalid JPEG: file too short")
	}

	// Verify SOI marker
	if data[0] != 0xFF || data[1] != markerSOI {
		return nil, fmt.Errorf("invalid JPEG: missing SOI marker")
	}

	var segments []JPEGSegment
	pos := 2 // Start after SOI

	for pos < len(data) {
		// Find next marker (0xFF followed by non-0xFF byte)
		for pos < len(data)-1 && (data[pos] != 0xFF || data[pos+1] == 0xFF || data[pos+1] == 0x00) {
			pos++
		}

		if pos >= len(data)-1 {
			break
		}

		marker := data[pos+1]

		// EOI marks end of image
		if marker == markerEOI {
			break
		}

		// SOF markers indicate start of image data - stop parsing segments
		if marker >= markerSOF0 && marker <= markerSOF3 {
			break
		}

		// Read segment length (2 bytes, big-endian)
		if pos+3 >= len(data) {
			return nil, fmt.Errorf("invalid JPEG: incomplete segment length")
		}

		length := binary.BigEndian.Uint16(data[pos+2 : pos+4])
		if length < 2 {
			return nil, fmt.Errorf("invalid JPEG: invalid segment length")
		}

		// Extract payload (length includes the 2 length bytes)
		payloadStart := pos + 4
		payloadEnd := pos + 2 + int(length)
		if payloadEnd > len(data) {
			return nil, fmt.Errorf("invalid JPEG: segment extends beyond file")
		}

		payload := make([]byte, payloadEnd-payloadStart)
		copy(payload, data[payloadStart:payloadEnd])

		segments = append(segments, JPEGSegment{
			Marker:  marker,
			Length:  length,
			Payload: payload,
		})

		pos = payloadEnd
	}

	return segments, nil
}

// FindAPP1Segment finds the EXIF APP1 segment
func FindAPP1Segment(segments []JPEGSegment) (int, *JPEGSegment) {
	for i, seg := range segments {
		if seg.Marker == markerAPP1 && len(seg.Payload) >= 6 {
			// Check for EXIF identifier
			if string(seg.Payload[0:6]) == "Exif\x00\x00" {
				return i, &seg
			}
		}
	}
	return -1, nil
}

// ReassembleJPEG reassembles JPEG segments into a complete JPEG file
func ReassembleJPEG(segments []JPEGSegment, imageData []byte) []byte {
	var buf bytes.Buffer

	// Write SOI marker
	buf.Write([]byte{0xFF, markerSOI})

	// Write all segments
	for _, seg := range segments {
		buf.WriteByte(0xFF)
		buf.WriteByte(seg.Marker)
		
		lengthBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lengthBytes, seg.Length)
		buf.Write(lengthBytes)
		
		buf.Write(seg.Payload)
	}

	// Write image data (everything after segments)
	buf.Write(imageData)

	// Write EOI marker if not present
	if len(imageData) == 0 || !bytes.HasSuffix(imageData, []byte{0xFF, markerEOI}) {
		buf.Write([]byte{0xFF, markerEOI})
	}

	return buf.Bytes()
}

// InsertEXIFSegment inserts or replaces EXIF APP1 segment
func InsertEXIFSegment(data []byte, exifPayload []byte) ([]byte, error) {
	// Parse segments (this stops at SOF markers)
	segments, err := ParseJPEGSegments(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JPEG: %v", err)
	}

	// Find existing APP1 segment
	app1Index, _ := FindAPP1Segment(segments)

	// Calculate APP1 segment length (payload + 2 bytes for length field)
	app1Length := uint16(len(exifPayload) + 2)

	// Create new APP1 segment
	newAPP1 := JPEGSegment{
		Marker:  markerAPP1,
		Length:  app1Length,
		Payload: exifPayload,
	}

	// Replace existing APP1 or insert new one
	if app1Index >= 0 {
		// Replace existing
		segments[app1Index] = newAPP1
	} else {
		// Insert at the beginning (after SOI, before other segments)
		newSegments := make([]JPEGSegment, 0, len(segments)+1)
		newSegments = append(newSegments, newAPP1)
		newSegments = append(newSegments, segments...)
		segments = newSegments
	}

	// Calculate where segments end in original file
	segmentsEnd := 2 // Start after SOI
	for pos := 2; pos < len(data); {
		// Find marker
		if pos >= len(data)-1 {
			break
		}
		if data[pos] != 0xFF {
			pos++
			continue
		}
		
		marker := data[pos+1]
		
		// Stop at SOF markers (start of image data)
		if marker >= markerSOF0 && marker <= markerSOF3 {
			segmentsEnd = pos
			break
		}
		
		// Stop at EOI
		if marker == markerEOI {
			segmentsEnd = pos
			break
		}
		
		// Skip this segment
		if pos+3 < len(data) {
			length := binary.BigEndian.Uint16(data[pos+2 : pos+4])
			pos += 2 + int(length)
		} else {
			break
		}
	}

	// Extract image data (everything from segmentsEnd to end)
	imageData := data[segmentsEnd:]

	// Reassemble JPEG
	return ReassembleJPEG(segments, imageData), nil
}

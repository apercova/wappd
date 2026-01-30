package processor_test

import (
	"testing"

	"github.com/apercova/wappd/internal/processor"
)

func TestParseMP4Atoms(t *testing.T) {
	// Create a minimal MP4 structure: ftyp atom
	ftypData := []byte{
		0x00, 0x00, 0x00, 0x20, // Size (32 bytes)
		'f', 't', 'y', 'p', // Type: ftyp
		0x69, 0x73, 0x6F, 0x6D, // Major brand
		0x00, 0x00, 0x00, 0x00, // Minor version
		0x69, 0x73, 0x6F, 0x6D, // Compatible brand
		0x69, 0x73, 0x6F, 0x32, // Compatible brand
		0x6D, 0x70, 0x34, 0x31, // Compatible brand
		0x6D, 0x70, 0x34, 0x32, // Compatible brand
	}

	atoms, err := processor.ParseMP4Atoms(ftypData)
	if err != nil {
		t.Fatalf("ParseMP4Atoms() error = %v", err)
	}

	if len(atoms) == 0 {
		t.Error("ParseMP4Atoms() returned no atoms")
	}

	if atoms[0].Type != "ftyp" {
		t.Errorf("ParseMP4Atoms() first atom type = %s, want ftyp", atoms[0].Type)
	}
}

func TestFindAtom(t *testing.T) {
	atoms := []processor.Atom{
		{Type: "ftyp", Size: 32, Data: make([]byte, 24)},
		{Type: "moov", Size: 100, Data: make([]byte, 92)},
		{Type: "mdat", Size: 1000, Data: make([]byte, 992)},
	}

	atom := processor.FindAtom(atoms, "moov")
	if atom == nil {
		t.Error("FindAtom() did not find moov atom")
	}
	if atom.Type != "moov" {
		t.Errorf("FindAtom() type = %s, want moov", atom.Type)
	}
}

func TestFindAtom_NotFound(t *testing.T) {
	atoms := []processor.Atom{
		{Type: "ftyp", Size: 32, Data: make([]byte, 24)},
		{Type: "mdat", Size: 1000, Data: make([]byte, 992)},
	}

	atom := processor.FindAtom(atoms, "moov")
	if atom != nil {
		t.Error("FindAtom() should return nil when not found")
	}
}

func TestFindAtomRecursive(t *testing.T) {
	moovAtom := processor.Atom{
		Type: "moov",
		Size: 100,
		Data: make([]byte, 92),
		Children: []processor.Atom{
			{Type: "mvhd", Size: 100, Data: make([]byte, 92)},
			{Type: "trak", Size: 50, Data: make([]byte, 42)},
		},
	}

	atom := processor.FindAtomRecursive(moovAtom, "mvhd")
	if atom == nil {
		t.Error("FindAtomRecursive() did not find mvhd atom")
	}
	if atom.Type != "mvhd" {
		t.Errorf("FindAtomRecursive() type = %s, want mvhd", atom.Type)
	}
}

func TestUnixToQuickTime(t *testing.T) {
	// Unix timestamp for 2025-01-22 00:00:00 UTC
	unixTime := int64(1737504000)
	qtTime := processor.UnixToQuickTime(unixTime)

	// QuickTime epoch is 1904-01-01, so 2025-01-22 should be a large number
	if qtTime == 0 {
		t.Error("UnixToQuickTime() returned 0")
	}

	// Verify round-trip conversion
	backToUnix := processor.QuickTimeToUnix(qtTime)
	if backToUnix != unixTime {
		t.Errorf("Round-trip conversion failed: %d -> %d -> %d", unixTime, qtTime, backToUnix)
	}
}

func TestQuickTimeToUnix(t *testing.T) {
	// QuickTime timestamp for 2025-01-22 00:00:00 UTC
	// This is Unix timestamp + quickTimeEpochOffset
	unixTime := int64(1737504000)
	qtTime := processor.UnixToQuickTime(unixTime)

	convertedBack := processor.QuickTimeToUnix(qtTime)
	if convertedBack != unixTime {
		t.Errorf("QuickTimeToUnix() = %d, want %d", convertedBack, unixTime)
	}
}

func TestParseMP4Atoms_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Too short",
			data: []byte{0x00, 0x00, 0x00},
		},
		{
			name: "Invalid size",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF, 'f', 't', 'y', 'p'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.ParseMP4Atoms(tt.data)
			if err == nil {
				t.Errorf("ParseMP4Atoms() expected error for %s", tt.name)
			}
		})
	}
}

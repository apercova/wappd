package processor

import (
	"encoding/binary"
	"fmt"
)

const (
	// QuickTime epoch: January 1, 1904 00:00:00 UTC
	// Offset from Unix epoch (January 1, 1970) in seconds
	quickTimeEpochOffset = 2082844800
)

// Atom represents an MP4 atom/box
type Atom struct {
	Size     uint32 // Atom size (including header)
	Type     string // Atom type (4 characters)
	Data     []byte // Atom data (excluding header)
	Children []Atom // Child atoms (for container atoms)
}

// ParseMP4Atoms parses MP4 file and extracts atoms
func ParseMP4Atoms(data []byte) ([]Atom, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var atoms []Atom
	pos := 0

	for pos < len(data) {
		if pos+8 > len(data) {
			if len(atoms) == 0 {
				return nil, fmt.Errorf("data too short: need at least 8 bytes for atom header, got %d", len(data))
			}
			break // Not enough data for another atom header, but we have some atoms
		}

		// Read atom header
		size := binary.BigEndian.Uint32(data[pos : pos+4])
		atomType := string(data[pos+4 : pos+8])

		// Handle special size values
		if size == 0 {
			// Size 0 means extends to end of file
			size = uint32(len(data) - pos)
		} else if size == 1 {
			// Size 1 means extended size follows (64-bit)
			if pos+16 > len(data) {
				return nil, fmt.Errorf("invalid atom: extended size extends beyond file")
			}
			// For simplicity, we'll handle this case by reading the extended size
			// But for most cases, we can skip this complexity
			return nil, fmt.Errorf("extended size atoms not yet supported")
		}

		if int(size) > len(data)-pos {
			return nil, fmt.Errorf("invalid atom: size %d extends beyond file", size)
		}

		// Extract atom data (excluding 8-byte header)
		atomData := make([]byte, size-8)
		if size > 8 {
			copy(atomData, data[pos+8:pos+int(size)])
		}

		atom := Atom{
			Size: size,
			Type: atomType,
			Data: atomData,
		}

		// Parse child atoms for container atoms
		if isContainerAtom(atomType) && len(atomData) > 0 {
			children, err := parseChildAtoms(atomData)
			if err == nil {
				atom.Children = children
			}
		}

		atoms = append(atoms, atom)
		pos += int(size)
	}

	return atoms, nil
}

// isContainerAtom checks if an atom type is a container (has children)
func isContainerAtom(atomType string) bool {
	containerAtoms := map[string]bool{
		"moov": true, // Movie atom
		"trak": true, // Track atom
		"mdia": true, // Media atom
		"minf": true, // Media information atom
		"stbl": true, // Sample table atom
		"edts": true, // Edit atom
		"udta": true, // User data atom
	}
	return containerAtoms[atomType]
}

// parseChildAtoms parses child atoms from parent atom data
func parseChildAtoms(data []byte) ([]Atom, error) {
	var atoms []Atom
	pos := 0

	// Some container atoms have version/flags before children
	// For moov, skip version/flags (1 byte version + 3 bytes flags = 4 bytes)
	// But this varies by atom type, so we'll start parsing from the beginning
	// and handle version/flags per atom type as needed

	for pos < len(data) {
		if pos+8 > len(data) {
			break
		}

		size := binary.BigEndian.Uint32(data[pos : pos+4])
		atomType := string(data[pos+4 : pos+8])

		if size == 0 {
			size = uint32(len(data) - pos)
		} else if size == 1 {
			return nil, fmt.Errorf("extended size atoms not yet supported in children")
		}

		if int(size) > len(data)-pos {
			break // Invalid size
		}

		atomData := make([]byte, size-8)
		if size > 8 {
			copy(atomData, data[pos+8:pos+int(size)])
		}

		atom := Atom{
			Size: size,
			Type: atomType,
			Data: atomData,
		}

		// Recursively parse children if container
		if isContainerAtom(atomType) && len(atomData) > 0 {
			children, err := parseChildAtoms(atomData)
			if err == nil {
				atom.Children = children
			}
		}

		atoms = append(atoms, atom)
		pos += int(size)
	}

	return atoms, nil
}

// FindAtom finds an atom by type (first level only)
func FindAtom(atoms []Atom, atomType string) *Atom {
	for i := range atoms {
		if atoms[i].Type == atomType {
			return &atoms[i]
		}
	}
	return nil
}

// FindAtomRecursive finds an atom by type recursively in children
func FindAtomRecursive(atom Atom, atomType string) *Atom {
	if atom.Type == atomType {
		return &atom
	}
	for i := range atom.Children {
		if found := FindAtomRecursive(atom.Children[i], atomType); found != nil {
			return found
		}
	}
	return nil
}

// UnixToQuickTime converts Unix timestamp to QuickTime timestamp
func UnixToQuickTime(unixTime int64) uint32 {
	return uint32(unixTime + quickTimeEpochOffset)
}

// QuickTimeToUnix converts QuickTime timestamp to Unix timestamp
func QuickTimeToUnix(qtTime uint32) int64 {
	return int64(qtTime) - quickTimeEpochOffset
}

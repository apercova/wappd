package processor

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// UpdateVideoMetadata updates creation date in MP4/MOV/3GP video files
func UpdateVideoMetadata(filePath string, dateTime time.Time) error {
	// Read the video file
	data, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Verify it's an MP4/MOV/3GP file (starts with ftyp atom)
	if len(data) < 8 {
		return fmt.Errorf("file too short to be a valid MP4/MOV/3GP")
	}

	// Check for ftyp atom (first atom should be ftyp)
	firstType := string(data[4:8])
	if firstType != "ftyp" {
		return fmt.Errorf("file does not appear to be a valid MP4/MOV/3GP (missing ftyp atom)")
	}

	// Parse atoms
	atoms, err := ParseMP4Atoms(data)
	if err != nil {
		return fmt.Errorf("failed to parse MP4 atoms: %v", err)
	}

	// Find moov atom
	moovAtom := FindAtom(atoms, "moov")
	if moovAtom == nil {
		return fmt.Errorf("moov atom not found")
	}

	// Find mvhd atom within moov
	mvhdAtom := FindAtomRecursive(*moovAtom, "mvhd")
	if mvhdAtom == nil {
		return fmt.Errorf("mvhd atom not found in moov")
	}

	// Update mvhd creation time
	newData, err := updateMvhdCreationTime(data, *mvhdAtom, dateTime)
	if err != nil {
		return fmt.Errorf("failed to update mvhd: %v", err)
	}

	// Write file back
	info, err := getFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	err = writeFile(filePath, newData, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// updateMvhdCreationTime updates the creation time in mvhd atom
func updateMvhdCreationTime(data []byte, mvhdAtom Atom, dateTime time.Time) ([]byte, error) {
	// Find mvhd atom position in file
	mvhdPos, err := findAtomPosition(data, "mvhd")
	if err != nil {
		return nil, fmt.Errorf("failed to find mvhd position: %v", err)
	}

	// mvhd structure:
	// - Header: 8 bytes (size + type)
	// - Version: 1 byte (0 or 1)
	// - Flags: 3 bytes
	// - Creation time: 4 bytes (if version 0) or 8 bytes (if version 1)
	// - Modification time: 4 bytes (if version 0) or 8 bytes (if version 1)
	// - Timescale: 4 bytes
	// - Duration: 4 bytes (if version 0) or 8 bytes (if version 1)
	// - ... rest of mvhd data

	if len(mvhdAtom.Data) < 4 {
		return nil, fmt.Errorf("mvhd atom data too short")
	}

	version := mvhdAtom.Data[0]
	creationTimeOffset := 4 // After version (1) + flags (3)

	// Convert dateTime to QuickTime timestamp
	unixTime := dateTime.Unix()
	qtTime := UnixToQuickTime(unixTime)

	// Create new data copy
	newData := make([]byte, len(data))
	copy(newData, data)

	if version == 0 {
		// Version 0: 32-bit timestamps
		if mvhdPos+8+creationTimeOffset+4 > len(newData) {
			return nil, fmt.Errorf("mvhd atom extends beyond file")
		}
		binary.BigEndian.PutUint32(newData[mvhdPos+8+creationTimeOffset:mvhdPos+8+creationTimeOffset+4], qtTime)
		// Also update modification time (4 bytes after creation time)
		binary.BigEndian.PutUint32(newData[mvhdPos+8+creationTimeOffset+4:mvhdPos+8+creationTimeOffset+8], qtTime)
	} else if version == 1 {
		// Version 1: 64-bit timestamps
		if mvhdPos+8+creationTimeOffset+8 > len(newData) {
			return nil, fmt.Errorf("mvhd atom extends beyond file")
		}
		binary.BigEndian.PutUint64(newData[mvhdPos+8+creationTimeOffset:mvhdPos+8+creationTimeOffset+8], uint64(qtTime))
		// Also update modification time (8 bytes after creation time)
		binary.BigEndian.PutUint64(newData[mvhdPos+8+creationTimeOffset+8:mvhdPos+8+creationTimeOffset+16], uint64(qtTime))
	} else {
		return nil, fmt.Errorf("unsupported mvhd version: %d", version)
	}

	return newData, nil
}

// findAtomPosition finds the byte position of an atom in the file
func findAtomPosition(data []byte, atomType string) (int, error) {
	pos := 0

	for pos < len(data) {
		if pos+8 > len(data) {
			break
		}

		size := binary.BigEndian.Uint32(data[pos : pos+4])
		currentType := string(data[pos+4 : pos+8])

		if currentType == atomType {
			return pos, nil
		}

		if size == 0 {
			break
		} else if size == 1 {
			return -1, fmt.Errorf("extended size atoms not supported")
		}

		// If it's a container atom, search recursively
		if isContainerAtom(currentType) && size > 8 {
			childPos, err := findAtomInChildren(data[pos+8:pos+int(size)], atomType)
			if err == nil {
				return pos + 8 + childPos, nil
			}
		}

		pos += int(size)
	}

	return -1, fmt.Errorf("atom %s not found", atomType)
}

// findAtomInChildren searches for an atom in child data
func findAtomInChildren(data []byte, atomType string) (int, error) {
	pos := 0

	for pos < len(data) {
		if pos+8 > len(data) {
			break
		}

		size := binary.BigEndian.Uint32(data[pos : pos+4])
		currentType := string(data[pos+4 : pos+8])

		if currentType == atomType {
			return pos, nil
		}

		if size == 0 {
			break
		} else if size == 1 {
			return -1, fmt.Errorf("extended size atoms not supported")
		}

		// Recursively search in children
		if isContainerAtom(currentType) && size > 8 {
			childPos, err := findAtomInChildren(data[pos+8:pos+int(size)], atomType)
			if err == nil {
				return pos + 8 + childPos, nil
			}
		}

		pos += int(size)
	}

	return -1, fmt.Errorf("atom %s not found in children", atomType)
}

// Helper functions to abstract file operations (for testing/mocking)
var (
	readFile   = readFileImpl
	writeFile  = writeFileImpl
	getFileInfo = getFileInfoImpl
)

func readFileImpl(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func writeFileImpl(path string, data []byte, mode os.FileMode) error {
	return os.WriteFile(path, data, mode)
}

func getFileInfoImpl(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

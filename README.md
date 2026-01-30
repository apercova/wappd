# wappd - WhatsApp Photo Date Extractor

A lightweight CLI tool written in Go to extract and restore EXIF data and file timestamps from WhatsApp backups. WhatsApp removes all EXIF information from images and videos, but the filenames contain the creation date—**wappd** extracts this date and applies it to your media files.

## ✨ Features

### Core Functionality
- **Date Extraction**: Automatically extracts creation dates from WhatsApp filename patterns
- **EXIF Restoration**: Writes EXIF DateTimeOriginal metadata to JPEG images
- **Video Metadata**: Updates creation dates in MP4/MOV/3GP video files
- **Batch Processing**: Process entire directories or individual files
- **Custom Patterns**: Support for custom date extraction via regex or pattern matching
- **File Timestamps**: Optionally update file modification times

### File Format Support
- **Images**: JPG, JPEG, PNG, GIF, BMP, WebP
- **Videos**: MP4, MOV, AVI, MKV, FLV, M4V, 3GP

### Smart Features
- **Configuration Files**: Persistent settings via `wappd.json` config file
- **Dry-Run Mode**: Preview changes before applying them
- **Verbose Output**: Detailed processing information
- **Flexible Output**: Add suffix, override originals, or save to custom directory
- **EXIF Preservation**: Option to preserve or overwrite existing EXIF data

## 📦 Installation

### Build from Source

```bash
git clone https://github.com/apercova/wappd.git
cd wappd
go build -o wappd .
```

### Requirements
- Go 1.16 or later

### Build Options

You can customize the output binary name:

```bash
# Default name
go build -o wappd .

# Custom name
go build -o my-wappd .

# With path
go build -o bin/wappd .
```

## 🚀 Quick Start

**Process all media files in current directory:**
```bash
./wappd
```

**Process a specific directory:**
```bash
./wappd -d ./whatsapp_backup
```

**Preview changes without modifying files:**
```bash
./wappd -d ./media --dry-run
```

**Process with verbose output:**
```bash
./wappd -d ./media -v
```

## 📖 Usage Guide

### Basic Operations

#### Process Single File
```bash
./wappd -f IMG-20250122-WA0003.jpg
```

#### Process Directory
```bash
./wappd -d ./WhatsApp/Media
```

#### Update File Modification Time
```bash
./wappd -d ./media -m
```
The `-m` flag updates both EXIF creation date and file modification time to the extracted date.

#### Override Original Files
```bash
./wappd -d ./media -o
```
By default, processed files get a `_modified` suffix. Use `-o` to overwrite the originals.

#### Specify Output Directory
```bash
./wappd -d ./media -out ./processed_media
```
Creates new files in the specified directory. If the output directory equals the input directory, a suffix is automatically added.

### Advanced Features

#### Dry-Run Mode
Preview what changes would be made without actually modifying files:
```bash
./wappd -d ./media --dry-run
```

#### Verbose Output
Get detailed information about processing:
```bash
./wappd -d ./media -v
```

#### Overwrite Existing EXIF Data
By default, existing EXIF data is preserved. Use `-ow` to completely replace it:
```bash
./wappd -d ./media -ow
```

#### Custom Date Extraction Patterns

**Using regex pattern (named group `date`):**
```bash
./wappd -d ./media -e 'IMG-(?P<date>\d{8})-WA'
```

**Using pattern format:**
```bash
./wappd -d ./media -p 'IMG-{date}-WA'
```

#### Override Extracted Date
Specify an ISO format date (YYYY-MM-DD) to override automatic extraction for all files:
```bash
./wappd -d ./media -dt 2025-01-22
```

### Configuration File

wappd supports configuration files to set default options. Create a `wappd.json` file in your working directory:

```json
{
  "updateModified": true,
  "overwriteExif": false,
  "overrideOriginal": false,
  "outputDir": "./processed",
  "verbose": true
}
```

**Using default config file:**
```bash
# wappd.json in current directory is automatically loaded
./wappd -d ./media
```

**Using custom config file:**
```bash
./wappd -d ./media -cf ./my-config.json
# or
./wappd -d ./media --config-file ./my-config.json
```

**Config file behavior:**
- Config file values provide defaults
- CLI flags override config file values
- Config file is optional (not required)

**Available config options:**
- `updateModified` (boolean): Update file modification time
- `overwriteExif` (boolean): Overwrite existing EXIF data
- `overrideOriginal` (boolean): Override original files (no suffix)
- `outputDir` (string): Output directory path
- `verbose` (boolean): Verbose output

## 📋 Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f` | string | "" | Path to a specific file to process |
| `-d` | string | "." | Input directory (default: current directory) |
| `-cf`, `--config-file` | string | "" | Path to config file (default: wappd.json in working directory) |
| `-dt` | string | "" | ISO format date (YYYY-MM-DD) to override extraction |
| `-e` | string | "" | Custom regex pattern with named group `date` |
| `-p` | string | "" | Custom pattern format with `{date}` placeholder |
| `-m` | bool | false | Also update file's last modified date |
| `-ow` | bool | false | Overwrite existing EXIF data |
| `-o` | bool | false | Override original files (don't add suffix) |
| `-out` | string | "" | Output directory for processed files |
| `-v` | bool | false | Verbose output (show detailed processing information) |
| `--dry-run` | bool | false | Preview changes without modifying files |

## 📝 WhatsApp Filename Patterns

wappd recognizes the following WhatsApp filename patterns:

### Default Patterns

**Image Pattern:**
- `IMG-YYYYMMDD-WA####.ext`
- Example: `IMG-20250122-WA0003.jpg` → Date: 2025-01-22

**Video Pattern:**
- `VID-YYYYMMDD-WA####.ext`
- Example: `VID-20240415-WA0010.mp4` → Date: 2024-04-15

**WhatsApp Image with Time:**
- `WhatsApp Image YYYY-MM-DD at H.MM.SS AM\|PM.ext`
- Example: `WhatsApp Image 2025-01-22 at 3.30.45 PM.jpg` → Date: 2025-01-22T15:30:45

**WhatsApp Video with Time:**
- `WhatsApp Video YYYY-MM-DD at H.MM.SS AM\|PM.ext`
- Example: `WhatsApp Video 2024-04-15 at 10.15.30 AM.mp4` → Date: 2024-04-15T10:15:30

### Custom Patterns

You can define custom patterns using regex or pattern format:

**Regex Pattern Requirements:**
- Must include a named group called `date` that captures 8 digits in YYYYMMDD format
- Example: `(?P<date>\d{8})`

**Pattern Format:**
- Use `{date}` placeholder for the date portion
- Example: `Photo-{date}-Custom`

## 💡 Examples

### Basic Usage
```bash
# Process all media in current directory
./wappd

# Process specific directory
./wappd -d ./whatsapp_backup

# Process single file
./wappd -f IMG-20250122-WA0003.jpg
```

### Advanced Workflows
```bash
# Process WhatsApp backup, update timestamps, save to output folder
./wappd -d ./WhatsApp/Media -m -out ./restored_media

# Process with verbose output and dry-run first
./wappd -d ./media -v --dry-run

# Process single file, override date, update mod time
./wappd -f media.jpg -dt 2024-12-25 -m

# Use custom pattern with output directory
./wappd -d ./backup -p 'Photo-{date}-Custom' -out ./processed

# Process with config file and overwrite originals
./wappd -d ./media -cf ./config.json -o
```

### Configuration Examples
```bash
# Create wappd.json in your working directory
cat > wappd.json << EOF
{
  "updateModified": true,
  "outputDir": "./processed",
  "verbose": false
}
EOF

# Use the config file
./wappd -d ./media

# Override config file settings with CLI flags
./wappd -d ./media -o  # Override original files (ignores config)
```

## 🔧 Implementation Status

### ✅ Fully Implemented
- Date extraction from filenames (default WhatsApp patterns + custom formats)
- EXIF DateTimeOriginal writing for JPEG files
- Video metadata (creation date) for MP4/MOV/3GP files
- File copying and organization
- File modification timestamp updates
- Configuration file support (`wappd.json`)
- Dry-run mode
- Verbose output
- Batch processing
- Custom pattern matching

### ⚠️ Partial Implementation
- **PNG and other image formats**: File timestamps only (EXIF writing not yet implemented)
- **Video formats (AVI, MKV, FLV, M4V)**: File timestamps only (metadata writing for MP4/MOV/3GP only)

### 📋 Known Limitations

1. **Image Format Support:**
   - JPEG: Full EXIF support ✅
   - PNG, GIF, BMP, WebP: File timestamps only (EXIF writing not implemented)

2. **Video Format Support:**
   - MP4, MOV, 3GP: Full metadata support ✅
   - AVI, MKV, FLV, M4V: File timestamps only

3. **Pattern Matching:**
   - Regex patterns must include a named group called `date` that captures 8 digits in YYYYMMDD format
   - Example: `(?P<date>\d{8})`

## 🏗️ Architecture

```
wappd/
├── main.go                          # CLI entry point and flag parsing
├── internal/processor/
│   ├── processor.go                # Core file processing logic
│   ├── exif.go                     # EXIF metadata handling
│   ├── exif_writer.go              # EXIF segment creation and writing
│   ├── exif_tags.go                # EXIF tag formatting
│   ├── jpeg_segments.go            # JPEG segment parsing
│   ├── video_metadata.go           # Video metadata handling
│   ├── mp4_atoms.go                # MP4 atom parsing
│   ├── config.go                   # Configuration file handling
│   └── ...
├── test/processor/                  # Test files
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## 🧪 Testing

Run tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Fork the repository
- Create a feature branch
- Submit a pull request

Areas that could use contributions:
- Additional image format support (PNG EXIF, TIFF, RAW formats)
- Enhanced video format support
- Performance optimizations
- Additional test coverage

## 📄 License

MIT License - See LICENSE file for details

---

**Need help?** Run `./wappd -help` for a quick reference of all flags.

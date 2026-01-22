# wappd - WhatsApp Photo Date Extractor

A lightweight CLI tool written in Go to extract and restore EXIF data and file timestamps from WhatsApp backups. WhatsApp removes all EXIF information from images and videos, but the filenames contain the creation date—**wappd** extracts this date and applies it to your media files.

## Features

✨ **Core Functionality:**
- Extract creation dates from WhatsApp filename patterns
- Support for custom date formats via regex or pattern matching
- Batch process entire directories of media files
- Process individual files with `-f` flag
- Override extracted dates with ISO format dates
- Update file modification timestamps
- Preserve or overwrite existing EXIF data

📁 **File Support:**
- **Images:** JPG, JPEG, PNG, GIF, BMP, WebP
- **Videos:** MP4, MOV, AVI, MKV, FLV, M4V

⚙️ **Smart Output:**
- Add `_modified` suffix to processed files (default)
- Option to override original files with `-o` flag
- Specify custom output directory with `-out` flag
- Automatic creation of output directories

## Installation

### Build from Source

```bash
git clone https://github.com/apercova/wappd.git
cd wappd
go build -o wappd
```

### Requirements
- Go 1.16+

## Usage

### Basic Examples

**Process all media in current directory:**
```bash
./wappd
```

**Process specific directory:**
```bash
./wappd -d ./whatsapp_backup
```

**Process single file:**
```bash
./wappd -f IMG-20250122-WA0003.jpg
```

### Advanced Options

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
This creates new files in the specified directory. If the output directory equals the input directory, a suffix is automatically added.

#### Use Custom Date Format

**With regex pattern (named group `date`):**
```bash
./wappd -d ./media -e 'IMG-(?P<date>\d{8})-WA'
```

**With pattern format:**
```bash
./wappd -d ./media -p 'IMG-{date}-WA'
```

#### Override Extracted Date
```bash
./wappd -d ./media -dt 2025-01-22
```
Specify an ISO format date (YYYY-MM-DD) to override automatic extraction for all files.

#### Overwrite Existing EXIF Data
```bash
./wappd -d ./media -ow
```
By default, existing EXIF data is preserved. Use `-ow` to completely replace it with extracted date information.

### Combined Examples

```bash
# Process WhatsApp backup, update timestamps, save to output folder
./wappd -d ./WhatsApp/Media -m -out ./restored_media

# Process single file, override date, update mod time
./wappd -f media.jpg -dt 2024-12-25 -m

# Use custom pattern with output directory
./wappd -d ./backup -p 'Photo-{date}-Custom' -out ./processed
```

## Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-f` | string | "" | Path to a specific file to process |
| `-d` | string | "." | Input directory (default: current directory) |
| `-dt` | string | "" | ISO format date (YYYY-MM-DD) to override extraction |
| `-e` | string | "" | Custom regex pattern with named group `date` |
| `-p` | string | "" | Custom pattern format with `{date}` placeholder |
| `-m` | bool | false | Also update file's last modified date |
| `-ow` | bool | false | Overwrite existing EXIF data |
| `-o` | bool | false | Override original files (don't add suffix) |
| `-out` | string | "" | Output directory for processed files |

## WhatsApp Filename Format

Default pattern: `IMG-YYYYMMDD-WA####.ext`

Examples:
- `IMG-20250122-WA0003.jpg` → Date: 2025-01-22
- `IMG-20240415-WA0010.mp4` → Date: 2024-04-15

## Current Status

⚠️ **Current Implementation:**
- ✅ Date extraction from filenames (default WhatsApp pattern + custom formats)
- ✅ File copying and organization
- ✅ File modification timestamp updates
- ⚠️ EXIF metadata writing (basic framework, limited implementation)
- ⚠️ Video metadata handling (file timestamps only, requires ffmpeg for full metadata)

## Known Limitations

1. **EXIF Writing:**
   - Currently updates file modification time, which serves as a fallback for date information
   - Full EXIF writing to JPEG, PNG, and other formats requires more extensive implementation or external tools

2. **Video Metadata:**
   - Video metadata (MP4, MOV, etc.) requires specialized parsers or external tools like `ffmpeg`
   - Current implementation updates file modification time only

3. **Regex Patterns:**
   - Must include a named group called `date` that captures 8 digits in YYYYMMDD format
   - Example: `(?P<date>\d{8})`

## Future Improvements

- [ ] Full EXIF APP1 segment writing for JPEG files
- [ ] Video metadata support via ffmpeg integration (optional feature)
- [ ] Batch processing with progress bar
- [ ] Configuration file support
- [ ] Dry-run mode to preview changes
- [ ] Detailed logging with verbosity levels
- [ ] Support for additional image formats (TIFF, RAW, etc.)

## Architecture

```
wappd/
├── main.go                          # CLI entry point and flag parsing
├── internal/processor/
│   ├── processor.go                # Core file processing logic
│   ├── exif.go                     # EXIF metadata handling
│   └── ...
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## Contributing

Feel free to fork, modify, and submit pull requests!

## License

MIT License - See LICENSE file for details

---

**Need help?** Run `./wappd -help` for a quick reference of all flags.

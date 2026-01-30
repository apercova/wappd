#!/bin/bash
# Build script for wappd with version information

set -e

# Get version from argument or git tag, default to "dev"
if [ -n "$1" ]; then
    VERSION="$1"
elif git describe --tags --exact-match HEAD 2>/dev/null >/dev/null; then
    VERSION=$(git describe --tags --exact-match HEAD | sed 's/^v//')
elif git describe --tags HEAD 2>/dev/null >/dev/null; then
    VERSION=$(git describe --tags HEAD | sed 's/^v//')
else
    VERSION="dev"
fi

# Get git commit hash
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Get build date
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X github.com/apercova/wappd/version.Version=$VERSION \
         -X github.com/apercova/wappd/version.GitCommit=$GIT_COMMIT \
         -X github.com/apercova/wappd/version.BuildDate=$BUILD_DATE"

# Output directory
OUTPUT_DIR="${OUTPUT_DIR:-.}"
BINARY_NAME="${BINARY_NAME:-wappd}"

# Build output path
if [ "$OUTPUT_DIR" = "." ]; then
    OUTPUT_PATH="../$BINARY_NAME"
    RUN_PATH="./$BINARY_NAME"
else
    OUTPUT_PATH="../$OUTPUT_DIR/$BINARY_NAME"
    RUN_PATH="./$OUTPUT_DIR/$BINARY_NAME"
fi

echo "Building wappd..."
echo "  Version: $VERSION"
echo "  Commit: $GIT_COMMIT"
echo "  Build Date: $BUILD_DATE"
echo "  Output: $OUTPUT_PATH"
echo ""

cd src
go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" .

echo ""
echo "Build complete!"
echo "Run '$RUN_PATH -version' to verify version information."

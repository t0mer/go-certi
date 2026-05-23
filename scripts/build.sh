#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
BUILD_MODE="${BUILD_MODE:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
MODULE="github.com/t0mer/go-certi"

mkdir -p "$OUTPUT_DIR"

LDFLAGS="-X ${MODULE}/internal/version.Version=${VERSION}"
if [[ "$BUILD_MODE" == "prod" ]]; then
    LDFLAGS="-s -w ${LDFLAGS}"
fi

declare -a TARGETS=(
    "linux   amd64  amd64  "
    "linux   arm64  arm64  "
    "linux   armv7  arm    7"
    "linux   armhf  arm    6"
    "linux   arm    arm    5"
    "windows amd64  amd64  "
    "windows arm64  arm64  "
)

for target in "${TARGETS[@]}"; do
    read -r os display_arch goarch goarm <<< "$target"
    ext=""
    [[ "$os" == "windows" ]] && ext=".exe"
    outfile="${OUTPUT_DIR}/go-certi_${VERSION}_${os}_${display_arch}${ext}"

    echo "Building ${outfile}..."
    env CGO_ENABLED=0 GOOS="$os" GOARCH="$goarch" GOARM="$goarm" \
        go build -ldflags "$LDFLAGS" -o "$outfile" ./cmd/go-certi
done

echo "Build complete. Binaries in ${OUTPUT_DIR}/"

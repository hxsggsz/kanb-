#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 3 ]; then
  echo "Usage: $0 <goos> <goarch> <version>" >&2
  echo "Example: $0 linux amd64 v1.2.0" >&2
  exit 1
fi

GOOS="$1"
GOARCH="$2"
VERSION="$3"

cd "$(dirname "$0")/.."

OUT_DIR="dist/kanba-${VERSION}-${GOOS}-${GOARCH}"
mkdir -p "$OUT_DIR"

echo "Building kanba ${VERSION} for ${GOOS}/${GOARCH}..."
GOOS="$GOOS" GOARCH="$GOARCH" go build \
  -ldflags "-X kanba/cmd.Version=${VERSION}" \
  -o "${OUT_DIR}/kanba" \
  .

tar -czf "dist/kanba-${VERSION}-${GOOS}-${GOARCH}.tar.gz" -C "dist" "kanba-${VERSION}-${GOOS}-${GOARCH}"

echo "Built ${OUT_DIR}/kanba and dist/kanba-${VERSION}-${GOOS}-${GOARCH}.tar.gz"

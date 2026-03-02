#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?Usage: release.sh <version>}"
TAG="v${VERSION#v}"
VERSION="${TAG#v}"

echo "Building rctl ${VERSION}..."

DIST="dist"
rm -rf "$DIST"
mkdir -p "$DIST"

for pair in darwin/arm64 darwin/amd64; do
  OS="${pair%/*}"
  ARCH="${pair#*/}"
  GOOS=$OS GOARCH=$ARCH go build \
    -ldflags "-X main.version=${VERSION}" \
    -o "${DIST}/rctl" ./cmd/rctl
  tar -czf "${DIST}/rctl-${OS}-${ARCH}.tar.gz" -C "${DIST}" rctl
  rm "${DIST}/rctl"
done

echo "Creating release ${TAG}..."
gh release create "${TAG}" \
  "${DIST}"/rctl-*.tar.gz \
  --title "rctl ${VERSION}" \
  --generate-notes

echo ""
echo "SHA256 checksums:"
shasum -a 256 "${DIST}"/*.tar.gz

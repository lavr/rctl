# RFC-003: GitHub Actions & Binary Release

## Status

Implemented (2026-03-02)

## Motivation

The repo is public. We need:

1. **CI** — run tests and lint on every push/PR to catch regressions early
2. **Release** — build cross-platform binaries and publish them as GitHub Release assets when a version tag is pushed

Currently, building and releasing is fully manual. Automating it via GitHub Actions removes human error and makes the release process reproducible.

## Design

### CI workflow

Triggered on every push to `master` and on pull requests.

Steps:
1. Checkout code
2. Set up Go
3. Run `go vet ./...`
4. Run `go test ./...`

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test ./... -count=1
```

### Release workflow

Triggered when a version tag (`v*`) is pushed. Builds binaries for macOS (ARM64 + AMD64), creates a GitHub Release, and uploads the artifacts.

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run tests
        run: go test ./... -count=1

      - name: Build binaries
        run: |
          VERSION="${GITHUB_REF_NAME#v}"
          mkdir -p dist

          for pair in darwin/arm64 darwin/amd64; do
            OS="${pair%/*}"
            ARCH="${pair#*/}"
            GOOS=$OS GOARCH=$ARCH go build \
              -ldflags "-X main.version=${VERSION}" \
              -o "dist/rctl" ./cmd/rctl
            tar -czf "dist/rctl-${OS}-${ARCH}.tar.gz" -C dist rctl
            rm dist/rctl
          done

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*.tar.gz
          generate_release_notes: true
```

### Manual release script

For local releases without CI (e.g. testing):

```bash
# scripts/release.sh
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
```

### Makefile target

```makefile
release:
	@test -n "$(VERSION)" || (echo "Usage: make release VERSION=0.1.0" && exit 1)
	./scripts/release.sh $(VERSION)
```

### Release process

Automated (preferred):

1. `git tag v0.1.0 && git push origin v0.1.0`
2. GitHub Actions builds binaries and creates the release automatically

Manual (fallback):

1. `make release VERSION=0.1.0`

### Binaries

Two target platforms (macOS):

| Platform | File |
|----------|------|
| macOS ARM64 (Apple Silicon) | `rctl-darwin-arm64.tar.gz` |
| macOS AMD64 (Intel) | `rctl-darwin-amd64.tar.gz` |

Linux can be added later by extending the `for` loop.

Each archive contains a single `rctl` file.

## Code changes

### 1. `.github/workflows/ci.yml` (new)

CI workflow — test and vet on push/PR.

### 2. `.github/workflows/release.yml` (new)

Release workflow — build binaries and publish GitHub Release on tag push.

### 3. `scripts/release.sh` (new)

Manual release script for local use.

### 4. `Makefile`

Add `release` target.

## Out of scope

- Linux binaries — add later if needed
- Code signing / notarization for macOS
- Goreleaser — the build matrix is small enough, a simple script is sufficient
- Docker images

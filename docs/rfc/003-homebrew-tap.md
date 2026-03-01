# RFC-003: Homebrew Tap

## Status

Draft

## Motivation

Currently, installing rctl is manual (`go build` + `scripts/install.sh`). The goal is:

```
brew install lavr/tap/rctl
```

The tap `lavr/homebrew-tap` already exists with a formula for `ansible-ssh`. We add a formula for rctl there.

## Prerequisites

The `lavr/rctl` repository is **private**. This determines the strategy choice.

### Problem: private repos and Homebrew

The standard Homebrew approach is to download a source tarball from GitHub (`archive/refs/tags/v*.tar.gz`). For private repos this requires `HOMEBREW_GITHUB_API_TOKEN` with access to private repos — inconvenient and fragile.

**Solution:** build binaries in advance and publish them as GitHub Release assets. The formula downloads a ready-made binary, not source code.

## Design

### Release process

1. Create a tag: `git tag v0.1.0 && git push origin v0.1.0`
2. Run the build script: `scripts/release.sh v0.1.0`
3. The script:
   - Builds binaries for `darwin-arm64` and `darwin-amd64`
   - Creates a GitHub Release via `gh release create`
   - Uploads binaries as assets
4. Update the formula in `lavr/homebrew-tap`

### Binaries

Two target platforms (macOS):

| Platform | File |
|----------|------|
| macOS ARM64 (Apple Silicon) | `rctl-darwin-arm64.tar.gz` |
| macOS AMD64 (Intel) | `rctl-darwin-amd64.tar.gz` |

Linux can be added later.

Each archive contains a single `rctl` file.

### Formula

`Formula/rctl.rb` in `lavr/homebrew-tap`:

```ruby
class Rctl < Formula
  desc "Run commands in client+domain context"
  homepage "https://github.com/lavr/rctl"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/lavr/rctl/releases/download/v0.1.0/rctl-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/lavr/rctl/releases/download/v0.1.0/rctl-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    bin.install "rctl"
  end

  test do
    assert_match "rctl version", shell_output("#{bin}/rctl version")
  end
end
```

**Note:** for private repos the user must configure access:
```bash
export HOMEBREW_GITHUB_API_TOKEN=ghp_...  # token with repo scope
```

### Shell completion

The formula includes completion installation:

```ruby
def install
  bin.install "rctl"

  generate_completions_from_executable(bin/"rctl", "completion")
end
```

Homebrew will automatically pick up `rctl completion bash`, `rctl completion zsh`, `rctl completion fish`.

### scripts/release.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?Usage: release.sh <version>}"
TAG="v${VERSION#v}"
VERSION="${TAG#v}"

echo "Building rctl ${VERSION}..."

DIST="dist"
rm -rf "$DIST"
mkdir -p "$DIST"

# Build for macOS ARM64
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o "${DIST}/rctl" ./cmd/rctl
tar -czf "${DIST}/rctl-darwin-arm64.tar.gz" -C "${DIST}" rctl
rm "${DIST}/rctl"

# Build for macOS AMD64
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o "${DIST}/rctl" ./cmd/rctl
tar -czf "${DIST}/rctl-darwin-amd64.tar.gz" -C "${DIST}" rctl
rm "${DIST}/rctl"

# Create GitHub release
echo "Creating release ${TAG}..."
gh release create "${TAG}" \
  "${DIST}/rctl-darwin-arm64.tar.gz" \
  "${DIST}/rctl-darwin-amd64.tar.gz" \
  --title "rctl ${VERSION}" \
  --notes "Release ${VERSION}"

# Print SHA256 for formula
echo ""
echo "SHA256 checksums (for formula):"
shasum -a 256 "${DIST}"/*.tar.gz
```

### Updating the formula

After `release.sh` — copy the sha256 into the formula and commit to `lavr/homebrew-tap`. This can be automated with a `scripts/update-formula.sh` script later.

## Code changes to rctl

### 1. `scripts/release.sh` (new)

Build and release publishing script.

### 2. `Makefile`

Add target:
```makefile
release:
	@test -n "$(VERSION)" || (echo "Usage: make release VERSION=0.1.0" && exit 1)
	./scripts/release.sh $(VERSION)
```

### 3. `Formula/rctl.rb` in `lavr/homebrew-tap` (new)

Homebrew formula with binary artifacts.

## Out of scope

- CI/CD automation (GitHub Actions) — manual release is sufficient for now
- Linux binaries — add later if needed
- Automatic formula updates — `update-formula.sh` script can be added later
- Publishing to homebrew-core — repo is private

## User installation

```bash
# One-time — add tap
brew tap lavr/tap

# Install
brew install lavr/tap/rctl

# Upgrade
brew upgrade lavr/tap/rctl
```

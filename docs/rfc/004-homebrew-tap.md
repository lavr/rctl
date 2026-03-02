# RFC-004: Homebrew Tap

## Status

Implemented (2026-03-02)

## Motivation

Currently, installing rctl is manual (`go build` + `scripts/install.sh`). The goal is:

```
brew install lavr/tap/rctl
```

The tap `lavr/homebrew-tap` already exists with a formula for `ansible-ssh`. We add a formula for rctl there.

## Design

The formula downloads pre-built binaries from GitHub Releases (see [RFC-003](003-github-actions-release.md)) instead of building from source — faster install, no Go toolchain required.

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

### Shell completion

The formula includes completion installation:

```ruby
def install
  bin.install "rctl"

  generate_completions_from_executable(bin/"rctl", "completion")
end
```

Homebrew will automatically pick up `rctl completion bash`, `rctl completion zsh`, `rctl completion fish`.

### Updating the formula

After a release (see [RFC-003](003-github-actions-release.md)) — copy the sha256 from release artifacts into the formula and commit to `lavr/homebrew-tap`. This can be automated with a `scripts/update-formula.sh` script later.

## Code changes

### `Formula/rctl.rb` in `lavr/homebrew-tap` (new)

Homebrew formula with binary artifacts.

## Out of scope

- Automatic formula updates — `update-formula.sh` script can be added later
- Publishing to homebrew-core — use tap for now

## User installation

```bash
# One-time — add tap
brew tap lavr/tap

# Install
brew install lavr/tap/rctl

# Upgrade
brew upgrade lavr/tap/rctl
```

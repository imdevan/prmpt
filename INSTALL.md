# Installation Guide

## Homebrew Installation (Recommended)

Once the Homebrew tap is set up, you can install prompter using:

```bash
brew install imdevan/prompter/prompter
```

## Manual Installation

### Download Pre-built Binaries

1. Go to the [Releases page](https://github.com/imdevan/prompter-cli/releases)
2. Download the appropriate binary for your platform:
   - macOS (Intel): `prompter-darwin-amd64`
   - macOS (Apple Silicon): `prompter-darwin-arm64`
   - Linux (Intel): `prompter-linux-amd64`
   - Linux (ARM64): `prompter-linux-arm64`

3. Make the binary executable and move it to your PATH:

```bash
# Example for macOS Apple Silicon
chmod +x prompter-darwin-arm64
sudo mv prompter-darwin-arm64 /usr/local/bin/prompter
```

### Build from Source

Requirements:
- Go 1.21 or later
- Git

```bash
# Clone the repository
git clone https://github.com/imdevan/prompter-cli.git
cd prompter-cli

# Build and install
make build
sudo make install
```

## Verification

Verify the installation by running:

```bash
prompter --version
```

You should see output similar to:
```
prompter version 1.0.0
  commit: abc1234
  built: 2024-01-01T12:00:00Z
  go version: go1.21.0
  platform: darwin/arm64
```

## Configuration

Create a configuration file at `~/.config/prompter/config.toml`:

```toml
prompts_location = "~/.config/prompter/prompts"
editor = "nvim"
default_pre = ""
default_post = ""
fix_file = "/tmp/prompter-fix.txt"
max_file_size_bytes = 65536
max_total_bytes = 262144
allow_oversize = false
directory_strategy = "git"
target = "clipboard"
```

## Uninstallation

### Homebrew
```bash
brew uninstall prompter
brew untap imdevan/prompter
```

### Manual
```bash
sudo rm /usr/local/bin/prompter
rm -rf ~/.config/prompter
```

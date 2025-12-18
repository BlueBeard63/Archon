# Archon Installation Guide

## Quick Install

To install Archon as a global command accessible from anywhere:

```bash
./install.sh
```

This will:
- Build the Archon TUI binary
- Install it to `~/.local/bin/archon` (user install) or `/usr/local/bin/archon` (root install)
- Create the config directory at `~/.config/archon`

After installation, you can run Archon from anywhere by simply typing:

```bash
archon
```

## Installation Locations

### User Installation (Recommended)
When running as a normal user, Archon is installed to:
```
~/.local/bin/archon
```

**Important:** Make sure `~/.local/bin` is in your PATH. If not, add this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.bashrc  # or source ~/.zshrc
```

### System-wide Installation
When running as root (with sudo), Archon is installed to:
```
/usr/local/bin/archon
```

This makes it available for all users on the system.

## Manual Installation

If you prefer to install manually:

```bash
# Build the binary
go build -o archon

# Copy to a directory in your PATH
cp archon ~/.local/bin/archon
# or for system-wide:
sudo cp archon /usr/local/bin/archon

# Make sure it's executable
chmod +x ~/.local/bin/archon
```

## Uninstallation

To uninstall Archon:

```bash
./uninstall.sh
```

This removes the binary but keeps your configuration files in `~/.config/archon`.

To completely remove everything including config:

```bash
./uninstall.sh
rm -rf ~/.config/archon
```

## Configuration

Archon stores its configuration in:
```
~/.config/archon/config.toml
```

Site-specific configs are stored in:
```
~/.config/archon/sites/<domain>/<site-name>/config.toml
```

Node configs are stored in:
```
~/.config/archon/nodes/<node-name>/config.toml
```

## Requirements

- Go 1.21 or later (for building)
- Linux/macOS/WSL

## First Run

On first run, Archon will create the config directory and a default configuration file. You can then:

1. Add nodes (servers running archon-node)
2. Add domains
3. Create and deploy sites

## Getting Help

Run `archon` and press `?` for keyboard shortcuts and help.

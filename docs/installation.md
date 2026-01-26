# Installation Guide

This guide covers all methods of installing ASK on your system.

## Requirements

- **Operating System**: macOS, Linux, or Windows
- **Go** (optional): 1.24+ if building from source

---

## macOS (Homebrew)

The easiest way to install ASK on macOS:

```bash
brew tap yeasy/ask
brew install ask
```

To update:

```bash
brew upgrade ask
```

---

## Linux

### Download Binary

Download the latest release for your architecture:

```bash
# AMD64
curl -LO https://github.com/yeasy/ask/releases/latest/download/ask_linux_amd64.tar.gz
tar xzf ask_linux_amd64.tar.gz
sudo mv ask /usr/local/bin/

# ARM64
curl -LO https://github.com/yeasy/ask/releases/latest/download/ask_linux_arm64.tar.gz
tar xzf ask_linux_arm64.tar.gz
sudo mv ask /usr/local/bin/
```

### Verify Installation

```bash
ask version
```

---

## Windows

### Download Binary

1. Download `ask_windows_amd64.zip` from [Releases](https://github.com/yeasy/ask/releases)
2. Extract the zip file
3. Add the directory to your PATH

### Using PowerShell

```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/yeasy/ask/releases/latest/download/ask_windows_amd64.zip" -OutFile "ask.zip"
Expand-Archive -Path "ask.zip" -DestinationPath "C:\tools\ask"

# Add to PATH (run as Administrator)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\tools\ask", "Machine")
```

---

## Build from Source

If you have Go 1.21+ installed:

```bash
git clone https://github.com/yeasy/ask.git
cd ask
make build
sudo mv ask /usr/local/bin/
```

Or using `go install`:

```bash
go install github.com/yeasy/ask@latest
```

---

## Verify Installation

After installation, verify ASK is working:

```bash
ask version
ask --help
```

## Next Steps

1. [Initialize your first project](commands.md#ask-init)
2. [Search for skills](commands.md#ask-search)
3. [Install a skill](commands.md#ask-install)

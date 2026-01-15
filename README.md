# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>Just ask, and your agent shall receive.</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

**ASK** (Agent Skills Kit) is a command-line interface (CLI) designed to be the package manager for AI Agent Skills. Similar to how `brew` manages macOS packages or `npm` manages Node.js dependencies, `ask` helps you discover, install, and manage capabilities for your AI agents.

## Features

- **📦 Package Management**: Install, uninstall, and list skills with ease.
- **🔍 Discovery**: Search GitHub for skills from multiple sources (topic-based and directory-based).
- **🌐 Multiple Sources**: Built-in support for the `agent-skill` topic and [Anthropics Skills](https://github.com/anthropics/skills).
- **⚡️ Fast & Native**: Built with Go, compiling to a single static binary with no runtime dependencies.
- **🛠 Project-Local**: Manages dependencies via `ask.yaml`, keeping your projects reproducible.

## Installation

### macOS (Homebrew)
```bash
brew tap yeasy/ask
brew install ask
```

### Manual
Download the latest release from the [Releases](https://github.com/yeasy/ask/releases) page, or build from source:

```bash
git clone https://github.com/yeasy/ask.git
cd ask
make build
mv ask /usr/local/bin/
```

## Usage

### 1. Initialize a Project
Start by initializing `ask` in your agent's root directory:
```bash
ask init
```
This creates an `ask.yaml` file to track your skills.

### 2. Search for Skills
Find skills relevant to your needs:
```bash
ask search browser
# Returns skills like: browser-use, web-surfer, etc.
```

### 3. Install a Skill
Install a skill directly from GitHub:
```bash
ask install browser-use/browser-use
```
This will:
- Clone the repository into `./skills/browser-use`
- Add the dependency to `ask.yaml`

### 4. Uninstall a Skill
Remove a skill you no longer need:
```bash
ask uninstall browser-use
```
or
```bash
ask uninstall browser-use/browser-use
```
This will:
- Remove the directory `skills/browser-use`
- Remove the dependency from `ask.yaml`

### 5. List Installed Skills
See what your agent is equipped with:
```bash
ask list
```

## Directory Structure
When you use `ask`, your project will look like this:

```text
my-agent/
├── ask.yaml          # Manifest file
├── main.py           # Your agent code
└── skills/           # Managed skills directory
    ├── browser-use/
    └── web-surfer/
```

## Skill Sources
By default, `ask` searches these sources:

1. **Community** (`topic`): GitHub repos with the `agent-skill` topic.
2. **Anthropics** (`dir`): Skills from [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills).

You can customize sources in your `ask.yaml`:

```yaml
version: "1.0"
skills:
  - browser-use
sources:
  - name: community
    type: topic
    url: agent-skill
  - name: anthropics
    type: dir
    url: anthropics/skills/skills
```

## Development
This project includes a `Makefile` to simplify common development tasks:

- `make build`: Compiles the binary to `ask`.
- `make test`: Runs unit tests.
- `make clean`: Removes the binary and runs `go clean`.
- `make run`: Runs `go run main.go`.
- `make deps`: Downloads dependencies (`go mod download`).
- `make fmt`: Formats code (`go fmt ./...`).
- `make vet`: Vets code (`go vet ./...`).
- `make install`: Installs the binary to `$GOPATH/bin` (`go install`).

## Contributing
We welcome contributions! Please check out our [Contribution Guidelines](CONTRIBUTING.md).

## License
MIT

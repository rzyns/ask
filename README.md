# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>Just ask, and your agent shall receive.</strong>
</p>

<p align="center">
  <img src="assets/usage.webp" alt="ASK Usage Demo" width="800"/>
</p>

<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/yeasy/ask/actions"><img src="https://img.shields.io/github/actions/workflow/status/yeasy/ask/release.yml?style=flat-square" alt="Build"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> · 
  <a href="#-commands">Commands</a> · 
  <a href="#-skill-sources">Sources</a> · 
  <a href="docs/README.md">Docs</a>
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

**ASK** (Agent Skills Kit) is the package manager for AI Agent Skills. Just like `brew` manages macOS packages or `npm` manages Node.js dependencies, `ask` helps you discover, install, and manage capabilities for your AI agents.

## ✨ Key Features

### 📦 Smart Package Management
Install, uninstall, update, and list skills with intuitive commands. Version locking via `ask.lock` ensures reproducible environments across teams.

### 🔍 Multi-Source Discovery  
Search skills from multiple sources simultaneously — community topics, official repositories, and scientific domains. See which skills you already have installed.

### ⚡ Lightning Fast
Built with Go, compiling to a single static binary with zero runtime dependencies. Parallel search across all sources. Git sparse checkout for minimal downloads.

### 🔒 Version Locking
Pin specific versions with `skill@v1.0` syntax. Track exact commits in `ask.lock` for reproducible installations.

### 📊 Progress Tracking
Real-time progress bars during installation and updates. Clear feedback on what's happening.

---

## 🚀 Quick Start

**① Install ASK**

```bash
# macOS (Homebrew)
brew tap yeasy/ask
brew install ask

# Or build from source
git clone https://github.com/yeasy/ask.git && cd ask
make build && mv ask /usr/local/bin/
```

**② Initialize Your Project**

```bash
ask init    # Creates ask.yaml in current directory
```

**③ Search & Install Skills**

```bash
ask skill search browser      # Search across all sources
ask skill install browser-use # Install a skill
ask skill list                # View installed skills
```

---

## 📋 Commands

| Command | Description |
|---------|-------------|
| `ask init` | Initialize project, create `ask.yaml` |
| **Skill Management** | |
| `ask skill search <keyword>` | Search skills across all sources |
| `ask skill install <skill>` | Install a skill to `./skills/` |
| `ask skill install skill@v1.0` | Install specific version |
| `ask skill uninstall <skill>` | Remove a skill |
| `ask skill list` | List installed skills |
| `ask skill info <skill>` | Show skill details |
| `ask skill update [skill]` | Update skill(s) to latest |
| `ask skill outdated` | Check for updates |
| `ask skill create <name>` | Create new skill template |
| **Repository Management** | |
| `ask repo list` | List skill sources |
| `ask repo add <url>` | Add custom source |
| `ask repo remove <name>` | Remove a source |

---

## 🌐 Skill Sources

ASK searches these sources by default:

| Source | Type | Description |
|--------|------|-------------|
| **Community** | topic | GitHub repos with `agent-skill` topic |
| **Anthropics** | dir | Official [anthropics/skills](https://github.com/anthropics/skills) |
| **MCP-Servers** | dir | [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers) |
| **Scientific** | dir | Research skills from [K-Dense-AI](https://github.com/K-Dense-AI/claude-scientific-skills) |
| **Superpowers** | dir | Core skills from [obra/superpowers](https://github.com/obra/superpowers) |
| **OpenAI** | dir | Codex skills from [openai/skills](https://github.com/openai/skills) |

### Add Custom Sources

```yaml
# ask.yaml
repos:
  - name: my-team
    type: dir
    url: my-org/agent-skills/skills
```

---

## 📂 Project Structure

After using ASK, your project looks like:

```
my-agent/
├── ask.yaml          # Manifest file
├── ask.lock          # Version lock file
├── main.py           # Your agent code
└── .agent/
    └── skills/       # Managed skills
        ├── browser-use/
        └── web-surfer/
```

---

## 🛠 Development

```bash
make build     # Build binary
make test      # Run tests
make fmt       # Format code
make vet       # Check code
make install   # Install to $GOPATH/bin
```

---

## 📚 Documentation

See the [docs/](docs/README.md) directory for detailed documentation:

- [Installation Guide](docs/installation.md)
- [Command Reference](docs/commands.md)
- [Skill Sources](docs/skill-sources.md)
- [SKILL.md Format](docs/skill-format.md)
- [Configuration](docs/configuration.md)

---

## 🤝 Contributing

We welcome contributions! Please see our [Contribution Guidelines](CONTRIBUTING.md).

---

## 📄 License

MIT

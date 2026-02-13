# ASK: Agent Skills Kit for Enterprise AI

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>The Missing Package Manager for Agent Skills</strong>
</p>

<p align="center">
  Just ask, the agents are ready!
</p>

<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <a href="https://goreportcard.com/report/github.com/yeasy/ask"><img src="https://goreportcard.com/badge/github.com/yeasy/ask?style=flat-square" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go" alt="Go Version">
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

<p align="center">
  <a href="#-quick-start">🚀 Quick Start</a> •
  <a href="#-key-features">✨ Features</a> •
  <a href="#-commands">📋 Commands</a> •
  <a href="docs/README.md">📚 Documentation</a>
</p>

---

**ASK** (Agent Skills Kit) is the package manager for AI Agent skills. Just like `brew` manages macOS packages, `pip` manages Python packages, or `npm` manages Node.js dependencies, `ask` helps you discover, install, and lock skills for your agents (Claude, Cursor, Codex, etc.).





## ✨ Key Features

| Feature | Description |
| :--- | :--- |
| **📦 Smart Management** | Install, update, and remove skills with ease. Includes `ask.lock` for reproducible builds. |
| **🔍 Multi-Source** | Unified search across GitHub and official repos (Anthropic, OpenAI, etc.). You can add more skill sources. |
| **🤖 Agent Agnostic** | Works with **any** agent. Auto-detects configuration for **Claude**, **Cursor**, **Codex**, and adapts to your custom agents. |
| **⚡ Blazing Fast** | Written in Go. Parallel downloads, sparse checkouts, and zero runtime dependencies. |
| **🔌 Offline Mode** | Full offline support with `--offline`. Perfect for air-gapped or secure environments. |
| **🌎 Global & Local** | Manage project-specific skills (`.agent/skills`) or user-wide tools (`~/.ask/skills`). |
| **🛡️ Security Audit** | Built-in security scanner checks skills for secrets, dangerous commands, and malware using entropy analysis. |
| **🖥️ Desktop & Web** | Beautiful UI available as `ask serve` web server or native desktop app via [Wails](https://wails.io). |

## 🖥️ Web UI & Desktop App

<p align="center">
  <img src="docs/images/skills.png" alt="ASK Skills Manager" width="800"/>
</p>

ASK provides a beautiful web interface for skill discovery and management — available as a **web server** (`ask serve`) or a **native desktop app**.

| Feature | Description |
| :--- | :--- |
| **📊 Visual Dashboard** | Overview of installed skills, repos, and system stats |
| **🔍 Skill Browser** | Search, filter, and install skills with rich metadata |
| **📦 Repository Manager** | Add and sync skill sources from GitHub |
| **🛡️ Security Audit** | View generated safety reports |

### Quick Start
```bash
# Web Server
ask serve

# Desktop App (requires Wails CLI)
wails build && ./build/bin/ask-desktop
```

📖 [Explore the Web UI Documentation →](docs/web-ui.md)


### 1. Install

**Homebrew (macOS/Linux):**
```bash
brew tap yeasy/tap
brew install yeasy/tap/ask              # CLI version
brew install --cask yeasy/tap/ask-desktop  # Desktop App (macOS only)
```

> [!NOTE]
> **macOS Users**: When opening `ask-desktop` for the first time, if you see an "unidentified developer" warning, please go to **System Settings > Privacy & Security**, and click **"Open Anyway"** in the Security section.

**Go Install:**
```bash
go install github.com/yeasy/ask@latest
```

**Binary / Manual Install (Windows / Linux / Desktop):**
Download the latest pre-compiled binary or Desktop App for your system from [Releases](https://github.com/yeasy/ask/releases).


### 2. Initialize
Enter your project directory and run:
```bash
ask init
```
This creates an `ask.yaml` configuration file.

### 3. Use

```bash
# Search for skills
ask search mcp

# Install a skill (by name or repo, `ask add` is an alias for `ask install`)
ask install anthropics/mcp-builder
ask install superpowers

# Install a skill from a root-level repository
ask install op7418/Youtube-clipper-skill

# Install specific version
ask install mcp-builder@v1.0.0

# Install for specific agent
ask install mcp-builder --agent claude
ask install mcp-builder --agent claude cursor

# Security Check
ask check .
ask check anthropics/mcp-builder -o report.html

# Restore skills from ask.lock or ask.yaml (if no arguments provided)
ask install

# Start Web UI
ask serve

# Install skills from a specific repository
ask skill install --repo anthropics pdf
# Install all skills from a specific repository
ask skill install --repo anthropics
```

## 📋 Commands

### Skill Management
| Command | Description |
| :--- | :--- |
| `ask skill search <keyword>` | Search across all sources |
| `ask skill install <name>` | Install skill(s) |
| `ask skill list` | List installed skills |
| `ask skill uninstall <name>` | Remove a skill |
| `ask skill update` | Update skills to latest version |
| `ask skill outdated` | Check for newer versions |
| `ask skill check <path>` | Security scan + SKILL.md format validation |
| `ask skill prompt [paths]` | Generate XML for agent system prompts |

### Repository Management
| Command | Description |
| :--- | :--- |
| `ask repo list` | Show configured repositories |
| `ask repo add <url>` | Add a custom skill source (run `ask repo sync` after to download) |
| `ask repo sync` | Download/update repos to local cache (`~/.ask/repos`) |

### System Commands
| Command | Description |
| :--- | :--- |
| `ask doctor` | Diagnose and report on ASK health (config, skills, cache, system) |
| `ask serve` | Start web UI for visual skill management |

## 🌐 Skill Sources

ASK comes pre-configured with trusted sources:

| Source | Description |
| :--- | :--- |
| **Anthropic** | Official [anthropics/skills](https://github.com/anthropics/skills) |
| **Community** | Top-rated community skills (GitHub `agent-skill` and `agent-skills` topics) |
| **Composio** | [ComposioHQ/awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) collection |
| **OpenAI** | Official [openai/skills](https://github.com/openai/skills) |
| **Vercel** | [vercel-labs/agent-skills](https://github.com/vercel-labs/agent-skills) AI SDK skills |

### Optional Repositories

For specific needs, you can add these additional sources:

| Repository | Command to Add | Description |
| :--- | :--- | :--- |
| **Scientific** | `ask repo add K-Dense-AI/claude-scientific-skills` | Data science & research skills |
| **MATLAB** | `ask repo add matlab/skills` | Official MATLAB integration |
| **Superpowers** | `ask repo add obra/superpowers` | Full dev workflow with sub-agents |
| **Planning** | `ask repo add OthmanAdi/planning-with-files` | File-based persistent planning |
| **UI/UX Pro** | `ask repo add nextlevelbuilder/ui-ux-pro-max-skill` | 57 UI styles, 95 color schemes |
| **NotebookLM** | `ask repo add PleasePrompto/notebooklm-skill` | Auto-upload to NotebookLM |
| **AI DrawIO** | `ask repo add GBSOSS/ai-drawio` | Flowchart & diagram generation |
| **PPT Skills** | `ask repo add op7418/NanoBanana-PPT-Skills` | Dynamic PPT generation |
| **Antigravity** | `ask repo add sickn33/antigravity-awesome-skills` | Collection of 600+ skills for Claude Code & Cursor |


## 🏗️ Architecture & Layout

For detailed architecture diagrams and installation layout, see [Architecture Guide](docs/architecture.md).

## 🐞 Debugging

To see detailed operational logs (scanning, updating, searching), set `ASK_LOG=debug`:

```bash
export ASK_LOG=debug
ask skill install browser-use
```

## ⌨️ Shell Completion

ASK supports intelligent tab completion for skill names, repository names, and agent flags.

**Setup (one-time):**
```bash
# Bash
ask completion bash > $(brew --prefix)/etc/bash_completion.d/ask

# Zsh
ask completion zsh > "${fpath[1]}/_ask"

# Fish
ask completion fish > ~/.config/fish/completions/ask.fish
```

**Features:**
- `ask skill install <TAB>` - Complete from cached skills
- `ask skill uninstall <TAB>` - Complete from installed skills  
- `ask repo sync <TAB>` - Complete from configured repositories
- `ask install --agent <TAB>` - Complete agent names (claude, cursor, codex, etc.)

## 📊 Security Audit Reports

<img src="reports/anthropics.png" width="300">
<img src="reports/openai.png" width="300">
<img src="reports/composio.png" width="300">
<img src="reports/vercel.png" width="300">
<img src="reports/superpowers.png" width="300">

See detailed security audit reports generated for top skill repositories:

- [🛡️ Anthropic Security Audit Report](reports/anthropics.html)
- [🛡️ OpenAI Security Audit Report](reports/openai.html)
- [🛡️ Composio Security Audit Report](reports/composio.html)
- [🛡️ Vercel Security Audit Report](reports/vercel.html)
- [🛡️ Superpowers Security Audit Report](reports/superpowers.html)

## 🤝 Contributing
Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License
MIT License. See [LICENSE](LICENSE) for details.

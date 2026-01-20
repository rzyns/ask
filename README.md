# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>The Missing Package Manager for AI Agents</strong>
</p>

<p align="center">
  Just ask, and your agent shall receive.
</p>

<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <a href="https://goreportcard.com/report/github.com/yeasy/ask"><img src="https://goreportcard.com/badge/github.com/yeasy/ask?style=flat-square" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
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

**ASK** (Agent Skills Kit) is the package manager for AI Agent capabilities. Just like `brew` manages macOS packages or `npm` manages Node.js dependencies, `ask` helps you discover, install, and lock skills for your AI agents (Claude, Cursor, Codex, etc.).

```mermaid
graph LR
    User[User/Agent] -->|ask skill search| Sources[GitHub/Community]
    Sources -->|Discover| Skills{Skills}
    User -->|ask skill install| Skills
    Skills -->|Download & Lock| Agent[.agent/skills/]
    
    style User fill:#4a9eff,color:white
    style Sources fill:#ff6b6b,color:white
    style Agent fill:#90ee90,color:black
```

<p align="center">
  <img src="assets/demo.png" alt="ASK CLI Demo" width="700"/>
</p>

## ✨ Key Features

| Feature | Description |
| :--- | :--- |
| **📦 Smart Management** | Install, update, and remove skills with ease. Includes `ask.lock` for reproducible builds. |
| **🔍 Multi-Source** | Unified search across GitHub and official repos (Anthropic, OpenAI, MATLAB). |
| **🤖 Multi-Agent** | Auto-detects and installs for **Claude** (`.claude/`), **Cursor** (`.cursor/`), **Codex** (`.codex/`), and more. |
| **⚡ Blazing Fast** | Written in Go. Parallel downloads, sparse checkouts, and zero runtime dependencies. |
| **🔌 Offline Mode** | Full offline support with `--offline`. Perfect for air-gapped or secure environments. |
| **🌍 Global & Local** | Manage project-specific skills (`.agent/skills`) or user-wide tools (`~/.ask/skills`). |

## 🚀 Quick Start

### 1. Install

**Homebrew (macOS/Linux):**
```bash
brew tap yeasy/ask
brew install ask
```

**Go Install:**
```bash
go install github.com/yeasy/ask@latest
```

### 2. Initialize
Enter your project directory and run:
```bash
ask init
```
This creates an `ask.yaml` configuration file.

### 3. Use
```bash
# Search for skills
ask skill search browser

# Install a skill (by name or repo)
ask skill install browser-use
ask skill install superpowers

# Install specific version
ask skill install browser-use@v1.0.0

# Install for specific agent
ask skill install browser-use --agent claude
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

### Repository Management
| Command | Description |
| :--- | :--- |
| `ask repo list` | Show configured repositories |
| `ask repo add <url>` | Add a custom skill source |

## 🌐 Skill Sources

ASK comes pre-configured with trusted sources:

| Source | Description |
| :--- | :--- |
| **Anthropic** | Official [anthropics/skills](https://github.com/anthropics/skills) |
| **Community** | Top-rated community skills (GitHub `agent-skill` topic) |
| **Composio** | [ComposioHQ/awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) collection |
| **MATLAB** | Official [matlab/skills](https://github.com/matlab/skills) |
| **OpenAI** | Official [openai/skills](https://github.com/openai/skills) |
| **Superpowers** | [obra/superpowers](https://github.com/obra/superpowers) core library |
| **Vercel** | [vercel-labs/agent-skills](https://github.com/vercel-labs/agent-skills) AI SDK skills |

## 📂 Installation Layout

Default structure after installation:
```text
my-project/
├── ask.yaml          # Project config
├── ask.lock          # Lockfile (commit hashes)
└── .agent/           
    └── skills/       # Default install location
        ├── browser-use/
        └── web-surfer/
```

**Agent-Specific Paths:**
- **Claude**: `.claude/skills/`
- **Cursor**: `.cursor/skills/`
- **Codex**: `.codex/skills/`

## 🤝 Contributing
Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License
MIT License. See [LICENSE](LICENSE) for details.

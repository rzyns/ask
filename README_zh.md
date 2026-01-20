# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>打造 AI 智能体的各种能力 - 缺少的那个包管理器</strong>
</p>

<p align="center">
  只需请求，您的智能体即可获得所需技能。
</p>

<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

<p align="center">
  <a href="#-快速开始">🚀 快速开始</a> •
  <a href="#-核心特性">✨ 核心特性</a> •
  <a href="#-常用命令">📋 常用命令</a> •
  <a href="docs/README.md">📚 文档</a>
</p>

---

**ASK** (Agent Skills Kit) 是专为 AI Agent 设计的技能包管理器。就像 `brew` 管理 macOS 软件、`npm` 管理 Node.js 依赖一样，`ask` 帮助您发现、安装和管理 AI 智能体的各种能力（支持 Claude, Cursor, Codex 等）。

```mermaid
graph LR
    User[用户/Agent] -->|ask skill search| Sources[GitHub/社区]
    Sources -->|发现| Skills{技能库}
    User -->|ask skill install| Skills
    Skills -->|下载 & 锁定| Agent[.agent/skills/]
    
    style User fill:#4a9eff,color:white
    style Sources fill:#ff6b6b,color:white
    style Agent fill:#90ee90,color:black
```

<p align="center">
  <img src="assets/demo.png" alt="ASK CLI Demo" width="700"/>
</p>

## ✨ 核心特性

| 特性 | 说明 |
| :--- | :--- |
| **📦 智能管理** | 轻松安装、升级、卸载技能。支持 `ask.lock` 版本锁定，确保环境一致性。 |
| **🔍 多源聚合** | 统一搜索 GitHub 及官方仓库 (Anthropic, OpenAI, MATLAB)。 |
| **🤖 多 Agent 支持** | 自动检测并适配 **Claude** (`.claude/`)、**Cursor** (`.cursor/`)、**Codex** (`.codex/`) 等。 |
| **⚡ 极速体验** | Go 语言编写，平行下载，稀疏检出（Sparse Checkout），无运行时依赖。 |
| **🔌 离线模式** | 支持 `--offline` 离线运行，使用本地缓存，适合安全受限环境。 |
| **🌍 全局与本地** | 支持项目级管理 (`.agent/skills`) 和用户级全局工具 (`~/.ask/skills`)。 |

## 🚀 快速开始

### 1. 安装

**Homebrew (macOS/Linux):**
```bash
brew tap yeasy/ask
brew install ask
```

**源码安装:**
```bash
git clone https://github.com/yeasy/ask.git
cd ask
make build && mv ask /usr/local/bin/
```

### 2. 初始化
进入项目目录并运行：
```bash
ask init
```
这将创建一个 `ask.yaml` 配置文件。

### 3. 使用
```bash
# 搜索技能
ask skill search browser

# 安装技能（按名称或仓库）
ask skill install browser-use

# 批量安装仓库中的所有技能
ask skill install superpowers

# 指定版本安装
ask skill install browser-use@v1.0.0

# 为特定 Agent 安装
ask skill install browser-use --agent claude
```

## 📋 常用命令

### 技能管理
| 命令 | 说明 |
| :--- | :--- |
| `ask skill search <关键字>` | 全网搜索技能 |
| `ask skill install <名称>` | 安装技能 |
| `ask skill list` | 列出已安装技能 |
| `ask skill uninstall <名称>` | 卸载技能 |
| `ask skill update` | 升级已安装技能 |
| `ask skill outdated` | 检查可用更新 |

### 仓库管理
| 命令 | 说明 |
| :--- | :--- |
| `ask repo list` | 查看配置的仓库源 |
| `ask repo add <url>` | 添加自定义源 |

## 🌐 技能来源

ASK 默认内置了以下受信源：

| 来源 | 说明 |
| :--- | :--- |
| **Community** | GitHub 社区高分技能 (`agent-skill` topic) |
| **Anthropic** | 官方库 [anthropics/skills](https://github.com/anthropics/skills) |
| **OpenAI** | 官方库 [openai/skills](https://github.com/openai/skills) |
| **MATLAB** | 官方库 [matlab/skills](https://github.com/matlab/skills) |
| **Superpowers** | 核心库 [obra/superpowers](https://github.com/obra/superpowers) |
| **Composio** | 精选集 [ComposioHQ/awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) |
| **Vercel** | AI SDK [vercel-labs/agent-skills](https://github.com/vercel-labs/agent-skills) |

## 📂 目录结构

安装后的默认结构：
```text
my-project/
├── ask.yaml          # 项目配置
├── ask.lock          # 版本锁定文件
└── .agent/           
    └── skills/       # 默认技能目录
        ├── browser-use/
        └── web-surfer/
```

**不同 Agent 的安装路径:**
- **Claude**: `.claude/skills/`
- **Cursor**: `.cursor/skills/`
- **Codex**: `.codex/skills/`

## 🤝 贡献参与
欢迎提交 PR 或 Issue！详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📄 许可证
MIT License. 详见 [LICENSE](LICENSE) 文件。

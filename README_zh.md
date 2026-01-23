# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>AI 智能体必备的技能管理器</strong>
</p>

<p align="center">
  只需 ask，智能体即可掌握新技能！
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

**ASK** (Agent Skills Kit) 是专为 AI Agent 设计的技能包管理器。就像 `brew` 管理 macOS 软件、`pip` 管理 Python 包、`npm` 管理 Node.js 依赖一样，`ask` 帮助您发现、安装和管理 AI 智能体的各种能力（支持 Claude, Cursor, Codex 等）。

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
ask skill search mcp

# 安装技能（按名称或仓库）
ask skill install mcp-builder

# 批量安装仓库中的所有技能
ask skill install superpowers

# 指定版本安装
ask skill install mcp-builder@v1.0.0

# 为特定 Agent 安装
ask skill install mcp-builder --agent claude
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
| `ask repo sync` | 同步仓库到本地缓存 |

## 🌐 技能来源

ASK 默认内置了以下受信源：

| 来源 | 说明 |
| :--- | :--- |
| **Anthropic** | 官方库 [anthropics/skills](https://github.com/anthropics/skills) |
| **Community** | GitHub 社区高分技能 (`agent-skill` 和 `agent-skills` topics) |
| **Composio** | 精选集 [ComposioHQ/awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) |
| **OpenAI** | 官方库 [openai/skills](https://github.com/openai/skills) |
| **Vercel** | AI SDK [vercel-labs/agent-skills](https://github.com/vercel-labs/agent-skills) |

### 可选技能仓库

如有特定需求，您可以添加以下额外来源：

| 仓库 | 添加命令 | 说明 |
| :--- | :--- | :--- |
| **SkillHub** | `ask repo add skillhub/skills` | [SkillHub.club](https://www.skillhub.club) 索引 |
| **Scientific** | `ask repo add K-Dense-AI/claude-scientific-skills` | 数据科学与研究技能 |
| **MATLAB** | `ask repo add matlab/skills` | 官方 MATLAB 集成 |
| **Superpowers** | `ask repo add obra/superpowers` | 全链路开发工作流 |
| **Planning** | `ask repo add OthmanAdi/planning-with-files` | 文件持久化规划 |
| **UI/UX Pro** | `ask repo add nextlevelbuilder/ui-ux-pro-max-skill` | 57种UI风格，95种配色 |
| **NotebookLM** | `ask repo add PleasePrompto/notebooklm-skill` | 自动上传到NotebookLM |
| **AI DrawIO** | `ask repo add GBSOSS/ai-drawio` | 流程图自动生成 |
| **PPT Skills** | `ask repo add op7418/NanoBanana-PPT-Skills` | 动态PPT生成 |

## 📂 目录结构

安装后的默认结构：
```text
my-project/
├── ask.yaml          # 项目配置
├── ask.lock          # 版本锁定文件
└── .agent/           
    └── skills/       # 默认技能目录
        ├── mcp-builder/
        └── writing-plans/
```

**不同 Agent 的安装路径:**
- **Claude**: `.claude/skills/`
- **Cursor**: `.cursor/skills/`
- **Codex**: `.codex/skills/`

## 🤝 贡献参与
欢迎提交 PR 或 Issue！详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📄 许可证
MIT License. 详见 [LICENSE](LICENSE) 文件。

# ASK: Agent Skills Kit for Enterprise AI

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>智能体必备的技能包管理器</strong>
</p>

<p align="center">
  只需一行命令，让智能体掌握无限可能。
</p>

<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go" alt="Go Version">
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

**ASK** (Agent Skills Kit) 是专为 AI Agent 设计的技能包管理器。就像 `brew` 管理 macOS 软件、`pip` 管理 Python 包一样，`ask` 帮助您高效发现、安装和管理 AI 智能体的各种能力（支持 Claude, Cursor, Codex 等）。





## ✨ 核心特性

| 特性 | 说明 |
| :--- | :--- |
| **📦 智能管理** | 轻松安装、升级和卸载。支持 `ask.lock` 版本锁定，确保环境一致性。 |
| **🔍 多源聚合** | 统一检索 GitHub 社区及官方仓库 (Anthropic, OpenAI 等)。支持添加更多自定义源。 |
| **🤖 Agent 无关** | 适用于 **任何** AI 智能体。自动适配 **Claude**, **Cursor**, **Codex**，并支持自定义 Agent 配置，不绑定特定厂商。 |
| **⚡ 极速体验** | 纯 Go 语言编写。支持并发下载、稀疏检出 (Sparse Checkout)，无运行时依赖，毫秒级响应。 |
| **🔌 离线模式** | 支持 `--offline` 离线模式，优先使用本地缓存，完美适配内网或安全受限环境。 |
| **🌎 全局与本地** | 灵活支持项目级 (`.agent/skills`) 和用户级 (`~/.ask/skills`) 隔离管理。 |
| **🛡️ 安全守卫** | 内置安全扫描引擎，通过熵值分析检测敏感信息泄漏、危险命令及恶意代码，为智能体保驾护航。 |

## 🖥️ Web UI & 桌面应用

<p align="center">
  <img src="docs/images/skills.png" alt="ASK 技能管理器" width="800"/>
</p>

ASK 提供精美的 Web 界面进行技能发现和管理 — 支持 **Web 服务器** (`ask serve`) 和 **原生桌面应用**。

| 功能 | 说明 |
| :--- | :--- |
| **📊 仪表盘** | 总览已安装技能、仓库和系统状态 |
| **🔍 技能浏览器** | 搜索、筛选并一键安装技能 |
| **📦 仓库管理** | 从 GitHub 添加并同步技能源 |
| **🛡️ 安全审计** | 查看生成的安全报告 |

### 快速启动
```bash
# Web 服务器
ask serve

# 桌面应用 (需要 Wails CLI)
wails build && ./build/bin/ask-desktop
```

📖 [Web UI 文档 →](docs/web-ui.md)

## 🚀 快速开始

### 1. 安装

**Homebrew (macOS/Linux):**
```bash
brew tap yeasy/tap
brew install yeasy/tap/ask              # 命令行版本
brew install --cask yeasy/tap/ask-desktop  # 桌面应用 (仅 macOS)
```

> [!NOTE]
> **macOS 用户请注意**：首次打开 `ask-desktop` 时若提示"无法验证开发者"，请前往 **系统设置 > 隐私与安全性**，在"安全性"区域点击 **"仍要打开" (Open Anyway)** 即可正常运行。

**源码安装:**
```bash
git clone https://github.com/yeasy/ask.git
cd ask
make build && mv ask /usr/local/bin/
make build-desktop # 构建桌面应用
```

**二进制 / 手动安装 (Windows / Linux):**
请前往 [Releases](https://github.com/yeasy/ask/releases) 页面下载对应系统的预编译二进制文件。



### 2. 初始化
进入项目目录并运行：
```bash
ask init
```
这将创建一个 `ask.yaml` 配置文件。

### 3. 使用
```bash
# 搜索 Skill
ask search mcp

# 安装 Skill (通过名称或仓库，支持使用 ask add 别名)
ask install anthropics/mcp-builder
ask add superpowers

# 安装根目录类型的 Skill (如 Youtube Clipper)
ask install op7418/Youtube-clipper-skill

# 安装指定版本
ask install mcp-builder@v1.0.0

# 为指定 Agent 安装
ask install mcp-builder --agent claude

# 安全检查
ask check .
ask check anthropics/mcp-builder -o report.html

# 启动 Web 管理界面
ask serve

# 从 ask.lock 或 ask.yaml 还原安装技能（不带参数运行）
ask install

# 从指定仓库安装技能
ask skill install --repo anthropics pdf
# 安装指定仓库下的所有技能
ask skill install --repo anthropics
```

## 📋 命令参考

### Skill 管理
| 命令 | 说明 |
| :--- | :--- |
| `ask search <keyword>` | 在所有源中搜索 |
| `ask install <name>` | 安装 Skill (别名: `add`, `i`) |
| `ask list` | 列出已安装的 Skill |
| `ask uninstall <name>` | 卸载 Skill |
| `ask update` | 更新 Skill 到最新版本 |
| `ask outdated` | 检查可用更新 |
| `ask check <path>` | 安全扫描 + SKILL.md 格式验证 |
| `ask skill prompt [paths]` | 生成 XML 格式供 Agent 系统提示使用 |

### 仓库管理
| 命令 | 说明 |
| :--- | :--- |
| `ask repo list` | 显示已配置的仓库 |
| `ask repo add <url>` | 添加自定义 Skill 源 (添加后可使用 `--sync` 或手动运行 `ask repo sync` 下载) |
| `ask repo sync` | 同步仓库到本地缓存 (`~/.ask/repos`) |

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
| **Scientific** | `ask repo add K-Dense-AI/claude-scientific-skills` | 数据科学与研究技能 |
| **MATLAB** | `ask repo add matlab/skills` | 官方 MATLAB 集成 |
| **Superpowers** | `ask repo add obra/superpowers` | 全链路开发工作流 |
| **Planning** | `ask repo add OthmanAdi/planning-with-files` | 文件持久化规划 |
| **UI/UX Pro** | `ask repo add nextlevelbuilder/ui-ux-pro-max-skill` | 57种UI风格，95种配色 |
| **NotebookLM** | `ask repo add PleasePrompto/notebooklm-skill` | 自动上传到NotebookLM |
| **AI DrawIO** | `ask repo add GBSOSS/ai-drawio` | 流程图自动生成 |
| **PPT Skills** | `ask repo add op7418/NanoBanana-PPT-Skills` | 动态PPT生成 |
| **Antigravity** | `ask repo add sickn33/antigravity-awesome-skills` | 600+ 个 Claude Code & Cursor 智能体技能合集 |


## 🏗️ 架构与布局

详细的架构图和安装布局说明，请参阅 [架构设计指南](docs/architecture_zh.md)。

## 🐞 调试

要查看详细的操作日志（如扫描、更新、搜索），请设置 `ASK_LOG=debug`：

```bash
export ASK_LOG=debug
ask skill install browser-use
```

## ⌨️ Shell 自动补全

ASK 支持智能 Tab 补全，可补全技能名称、仓库名称和 agent 参数。

**设置 (一次性):**
```bash
# Bash
ask completion bash > $(brew --prefix)/etc/bash_completion.d/ask

# Zsh
ask completion zsh > "${fpath[1]}/_ask"

# Fish
ask completion fish > ~/.config/fish/completions/ask.fish
```

**支持功能:**
- `ask skill install <TAB>` - 从缓存中补全技能名
- `ask skill uninstall <TAB>` - 从已安装技能中补全
- `ask repo sync <TAB>` - 从已配置仓库中补全
- `ask install --agent <TAB>` - 补全 agent 名称 (claude, cursor, codex 等)

## 📊 安全审计报告

<img src="reports/anthropics.png" width="300">
<img src="reports/openai.png" width="300">
<img src="reports/composio.png" width="300">
<img src="reports/vercel.png" width="300">
<img src="reports/superpowers.png" width="300">

完整安全审计报告：

- [🛡️ Anthropic 安全审计报告](reports/anthropics.html)
- [🛡️ OpenAI 安全审计报告](reports/openai.html)
- [🛡️ Composio 安全审计报告](reports/composio.html)
- [🛡️ Vercel 安全审计报告](reports/vercel.html)
- [🛡️ Superpowers 安全审计报告](reports/superpowers.html)

## 🤝 贡献参与
欢迎提交 PR 或 Issue！详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📄 许可证
MIT License. 详见 [LICENSE](LICENSE) 文件。

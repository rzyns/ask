# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>只需询问，您的智能体即可获得。</strong>
</p>



<p align="center">
  <a href="https://github.com/yeasy/ask/releases"><img src="https://img.shields.io/github/v/release/yeasy/ask?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://github.com/yeasy/ask/blob/main/LICENSE"><img src="https://img.shields.io/github/license/yeasy/ask?style=flat-square" alt="License"></a>
  <a href="https://github.com/yeasy/ask/stargazers"><img src="https://img.shields.io/github/stars/yeasy/ask?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/yeasy/ask/actions"><img src="https://img.shields.io/github/actions/workflow/status/yeasy/ask/release.yml?style=flat-square" alt="Build"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
</p>

<p align="center">
  <a href="#-快速开始">快速开始</a> · 
  <a href="#-命令列表">命令列表</a> · 
  <a href="#-技能来源">技能来源</a> · 
  <a href="docs/README.md">文档</a>
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

**ASK**（Agent Skills Kit，智能体技能工具包）是 AI 智能体技能的包管理器。如同 `brew` 管理 macOS 软件包、`npm` 管理 Node.js 依赖一样，`ask` 帮助您发现、安装和管理智能体的能力扩展。

## 🏆 为什么选择 ASK？

| 特性 | ASK | 手动管理 | 其他工具 |
|------|-----|---------|----------|
| **多源搜索** | ✅ 统一搜索 GitHub、Anthropic、MCP、OpenAI | ❌ 需分别查找各平台 | ⚠️ 通常限于特定来源 |
| **多工具支持** | ✅ Claude/Cursor/Codex/OpenCode | ❌ | ⚠️ 通常仅支持单一工具 |
| **版本锁定** | ✅ `ask.lock` 精确到 commit | ❌ | ⚠️ 部分支持 |
| **离线模式** | ✅ `--offline` 标志 | ❌ | ⚠️ |
| **安装速度** | ⚡ 并行 + 稀疏检出 | 🐢 | 🐢 |
| **全局/项目级** | ✅ 两者皆可 | ⚠️ | ⚠️ |
| **零依赖** | ✅ 单个静态二进制 | N/A | ⚠️ 通常需要运行时 |

---

## ✨ 核心特性



### 📦 智能包管理
使用直观的命令安装、卸载、更新和列出技能。通过 `ask.lock` 进行版本锁定，确保团队间环境的可复现性。

### 🔍 多源搜索  
同时从多个来源搜索技能 — 社区主题、官方仓库和科研领域。清楚显示已安装的技能。

### ⚡ 极速体验
使用 Go 构建，编译为单个静态二进制文件，零运行时依赖。并行搜索所有来源。Git 稀疏检出实现最小化下载。

### 🔒 版本锁定
使用 `skill@v1.0` 语法锁定特定版本。在 `ask.lock` 中跟踪精确提交，实现可复现安装。

### 📊 进度追踪
安装和更新过程中实时显示进度条。清晰反馈当前操作状态。

### 🔌 离线模式
使用 `--offline` 标志无需网络即可工作。搜索使用缓存结果，适用于无网络环境。

### 🌍 全局安装
使用 `--global` 标志全局安装技能，可在所有项目间共享。项目级安装优先于全局安装。

---

## 🚀 快速开始

**① 安装 ASK**

```bash
# macOS (Homebrew)
brew tap yeasy/ask
brew install ask

# 或从源码构建
git clone https://github.com/yeasy/ask.git && cd ask
make build && mv ask /usr/local/bin/
```

**② 初始化项目**

```bash
ask init    # 在当前目录创建 ask.yaml
```

**③ 搜索并安装技能**

```bash
ask skill search browser      # 从所有来源搜索
ask skill install browser-use           # 安装技能
ask skill install skill1 skill2 skill3  # 批量安装多个技能
ask skill list                # 查看已安装技能
```

---

## 📋 命令列表

| 命令 | 说明 |
|------|------|
| `ask init` | 初始化项目，创建 `ask.yaml` |
| **技能管理** | |
| `ask skill search <关键词>` | 从所有来源搜索技能 |
| `ask skill install <技能...>` | 安装技能到 `.agent/skills/` |
| `ask skill install skill@v1.0` | 安装指定版本 |
| `ask skill uninstall <技能>` | 移除技能 |
| `ask skill list` | 列出已安装技能 |
| `ask skill info <技能>` | 显示技能详情 |
| `ask skill update [技能]` | 更新技能到最新版 |
| `ask skill outdated` | 检查可更新技能 |
| `ask skill create <名称>` | 创建新技能模板 |
| **仓库管理** | |
| `ask repo list` | 列出技能来源 |
| `ask repo add <url>` | 添加自定义来源 |
| `ask repo remove <名称>` | 移除来源 |
| **工具命令** | |
| `ask benchmark` | 运行性能基准测试 |
| `ask completion <shell>` | 生成 shell 补全脚本 |
| `--offline` | 全局标志：无网络模式 |
| `--global, -g` | 全局标志：使用全局安装 (~/.ask/skills) |

---

## 🌐 技能来源

ASK 默认搜索以下来源：

| 来源 | 类型 | 说明 |
|------|------|------|
| **Community** | topic | 带有 `agent-skill` 主题的 GitHub 仓库 |
| **Anthropics** | dir | 官方 [anthropics/skills](https://github.com/anthropics/skills) |
| **Scientific** | dir | [K-Dense-AI](https://github.com/K-Dense-AI/claude-scientific-skills) 科研技能 |
| **Superpowers** | dir | [obra/superpowers](https://github.com/obra/superpowers) 核心技能 |
| **OpenAI** | dir | [openai/skills](https://github.com/openai/skills) Codex 技能 |
| **MATLAB** | dir | 官方 [matlab/skills](https://github.com/matlab/skills) |
| **Composio** | dir | [Composio/awesome-claude-skills](https://github.com/Composio/awesome-claude-skills) Awesome Claude Skills |

### 🔍 发现更多技能

想要寻找更多技能？请访问 [SkillsMP](https://skillsmp.com)，这是最大的开源 AI Agent 技能市场。您可以搜索数以千计的社区技能，并通过 `ask repo add` 添加它们，或直接从其仓库安装。

### 添加自定义来源

```yaml
# ask.yaml
repos:
  - name: my-team
    type: dir
    url: my-org/agent-skills/skills
```

---

## 📂 项目结构

使用 ASK 后，您的项目结构如下：

```
my-agent/
├── ask.yaml          # 清单文件
├── ask.lock          # 版本锁定文件
├── main.py           # 您的智能体代码
└── .agent/
    └── skills/       # 项目级技能（默认）

### 安装路径

技能会根据不同的 Agent 安装到特定目录：

| Agent | Flag 参数 | 路径 |
|-------|------|------|
| **通用/默认** | (无) | `.agent/skills/` |
| **Claude** | `--agent claude` | `.claude/skills/` |
| **Cursor** | `--agent cursor` | `.cursor/skills/` |
| **Codex** | `--agent codex` | `.codex/skills/` |
| **OpenCode** | `--agent opencode` | `.opencode/skills/` |
| **全局** | `--global` | `~/.ask/skills/` |目录
        ├── browser-use/
        └── web-surfer/

# 全局技能（跨项目共享）
~/.ask/
├── config.yaml       # 全局配置
├── ask.lock          # 全局版本锁定
└── skills/           # 全局技能
    └── shared-skill/
```

### 安装范围

```bash
# 项目级（默认）- 存储在 .agent/skills/
ask skill install browser-use

# 全局 - 存储在 ~/.ask/skills/，所有项目可用
ask skill install --global shared-skill
ask skill install -g shared-skill

# 列出全部
ask skill list --all
```

---

## 🛠 开发

```bash
make build     # 构建二进制文件
make test      # 运行测试
make fmt       # 格式化代码
make vet       # 检查代码
make install   # 安装到 $GOPATH/bin
```

---

## 📚 文档

查看 [docs/](docs/README.md) 目录获取详细文档：

- [安装指南](docs/installation.md)
- [命令参考](docs/commands.md)
- [技能来源](docs/skill-sources.md)
- [SKILL.md 格式](docs/skill-format.md)
- [配置说明](docs/configuration.md)

---

## 🤝 贡献

我们欢迎贡献！请查看 [贡献指南](CONTRIBUTING.md)。

---

## 📄 许可证

MIT

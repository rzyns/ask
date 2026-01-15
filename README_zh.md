# ASK: Agent Skills Kit

<p align="center">
  <img src="assets/logo.png" alt="ASK Logo" width="150"/>
</p>

<p align="center">
  <strong>只需询问，您的智能体即可获得。</strong>
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> •
  <a href="#安装">安装</a> •
  <a href="#使用方法">使用方法</a> •
  <a href="#贡献">贡献</a>
</p>

---

**ASK**（Agent Skills Kit，智能体技能工具包）是一个命令行工具，专门用于管理 AI 智能体的技能包。类似于 `brew` 管理 macOS 软件包或 `npm` 管理 Node.js 依赖，`ask` 帮助您发现、安装和管理智能体的能力扩展。

## 功能特性

- **📦 包管理**：轻松安装、卸载和列出技能。
- **🔍 多源搜索**：从多个来源搜索技能（基于主题和目录）。
- **🌐 多数据源**：内置支持 `agent-skill` 主题和 [Anthropics Skills](https://github.com/anthropics/skills)。
- **⚡️ 快速原生**：使用 Go 构建，编译为单个静态二进制文件，无运行时依赖。
- **🛠 项目本地化**：通过 `ask.yaml` 管理依赖，保持项目可复现性。

## 安装

### macOS (Homebrew)
```bash
brew tap yeasy/ask
brew install ask
```

### 手动安装
从 [Releases](https://github.com/yeasy/ask/releases) 页面下载最新版本，或从源码构建：

```bash
git clone https://github.com/yeasy/ask.git
cd ask
make build
mv ask /usr/local/bin/
```

## 使用方法

### 1. 初始化项目
在您的智能体根目录中初始化 `ask`：
```bash
ask init
```
这将创建一个 `ask.yaml` 文件来跟踪您的技能。

### 2. 搜索技能
查找符合需求的技能：
```bash
ask search browser
# 返回如: browser-use, web-surfer 等技能
```

### 3. 安装技能
直接从 GitHub 安装技能：
```bash
ask install browser-use/browser-use
```
这将：
- 将仓库克隆到 `./skills/browser-use`
- 将依赖添加到 `ask.yaml`

### 4. 卸载技能
移除不再需要的技能：
```bash
ask uninstall browser-use
```
这将：
- 删除 `skills/browser-use` 目录
- 从 `ask.yaml` 中移除依赖

### 5. 列出已安装技能
查看您的智能体已装备的技能：
```bash
ask list
```

## 目录结构
使用 `ask` 后，您的项目结构如下：

```text
my-agent/
├── ask.yaml          # 清单文件
├── main.py           # 您的智能体代码
└── skills/           # 托管技能目录
    ├── browser-use/
    └── web-surfer/
```

## 技能来源
默认情况下，`ask` 搜索以下来源：

1. **Community**（`topic`）：带有 `agent-skill` 主题的 GitHub 仓库。
2. **Anthropics**（`dir`）：来自 [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills) 的技能。

您可以在 `ask.yaml` 中自定义来源：

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

## 开发
本项目包含 `Makefile` 以简化常见开发任务：

- `make build`：编译二进制文件为 `ask`。
- `make test`：运行单元测试。
- `make clean`：删除二进制文件并运行 `go clean`。
- `make run`：运行 `go run main.go`。
- `make deps`：下载依赖（`go mod download`）。
- `make fmt`：格式化代码（`go fmt ./...`）。
- `make vet`：检查代码（`go vet ./...`）。
- `make install`：将二进制文件安装到 `$GOPATH/bin`（`go install`）。

## 贡献
我们欢迎贡献！请查看我们的 [贡献指南](CONTRIBUTING.md)。

## 许可证
MIT

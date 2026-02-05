# ASK 架构设计

本文档详细描述了 ASK (Agent Skills Kit) 的技术架构。

## 系统概览

ASK 被设计为一个用 Go 语言编写的轻量级、快速的命令行工具，其管理 AI Agent 技能的方式类似于 Homebrew 或 npm 管理依赖项。

```mermaid
graph LR
    subgraph "User Interface"
        direction TB
        CLI[Terminal / CLI]
        GUI[Web UI / Desktop]
    end

    subgraph "ASK Core"
        direction TB
        Mgr[Skill Manager]
        Sec[Security Audit]
        Config[Config ask.yaml]
        Lock[Lock ask.lock]
    end

    subgraph "Cloud Ecosystem"
        GitHub[GitHub / Community]
        Official[Official Repos]
    end

    subgraph "Agent Environment"
        direction TB
        Project[.agent/skills/]
        Global[~/.ask/skills/]
        Agents{Agents}
    end

    CLI --> Mgr
    GUI --> Mgr
    
    Mgr <-->|Discover & Pull| GitHub
    Mgr <-->|Discover & Pull| Official
    
    Mgr -->|Scan| Sec
    Mgr <-->|Read/Write| Config
    Mgr -->|Write| Lock
    
    Mgr -->|Install| Project
    Mgr -->|Install| Global
    
    Project -.->|Load| Agents
    Global -.->|Load| Agents

    style Mgr fill:#4a9eff,color:white
    style Sec fill:#ff6b6b,color:white
    style Agents fill:#90ee90,color:black
```

## 核心组件

### 1. CLI 层 (`cmd/`)

命令层使用 [Cobra](https://github.com/spf13/cobra) 作为 CLI 框架。

**目录结构**:
```
cmd/
├── root.go          # 根命令与配置
├── init.go          # 项目初始化
├── skill.go         # 技能父命令
├── search.go        # 技能搜索
├── install.go       # 技能安装
├── uninstall.go     # 技能卸载
├── update.go        # 技能更新
├── outdated.go      # 检查过期技能
├── list.go          # 列出已安装技能
├── info.go          # 技能详情
├── create.go        # 创建技能模板
├── repo.go          # 仓库管理
└── completion.go    # Shell 补全
```

**命令流程**:
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Config
    participant Internal

    User->>CLI: ask skill install browser-use
    CLI->>Config: Load ask.yaml
    Config-->>CLI: Configuration loaded
    CLI->>Internal: Execute install logic
    Internal->>GitHub: Clone repository
    GitHub-->>Internal: Repository downloaded
    Internal->>Config: Update ask.yaml & ask.lock
    Config-->>CLI: Config saved
    CLI-->>User: Installation complete
```

### 2. 内部包 (`internal/`)

#### 配置管理 (`internal/config/`)

处理配置文件和锁定文件。

**关键文件**:
- `config.go`: 主配置逻辑
- `lock.go`: 版本锁定机制

**配置流程**:
```mermaid
graph LR
    A[Load ask.yaml] --> B{Exists?}
    B -->|Yes| C[Parse YAML]
    B -->|No| D[Create Default]
    C --> E[Merge Default Repos]
    D --> E
    E --> F[Return Config]
```

#### GitHub 集成 (`internal/github/`)

处理 GitHub API 交互以进行技能发现。

**特性**:
- 基于主题的搜索 (GitHub topics)
- 基于目录的搜索 (仓库子目录)
- 结果缓存以提高性能

#### Git 操作 (`internal/git/`)

处理所有 Git 相关操作。

**关键功能**:
- `Clone()`: 标准 git clone
- `SparseClone()`: 高效的子目录克隆
- `InstallSubdir()`: 从仓库子目录安装
- `GetLatestTag()`: 获取最新版本标签

**稀疏检出优化的理由**:
- **速度**: 仅下载所需文件
- **磁盘空间**: 占用更小
- **带宽**: 减少网络使用

对于像 `anthropics/skills` 这样的 monorepo，这比完整克隆快 10-100 倍。

#### 技能解析 (`internal/skill/`)

解析 `SKILL.md` 文件以获取元数据。

**SKILL.md 格式**:
```yaml
---
name: browser-use
description: Browser automation for AI agents
version: 1.0.0
author: browser-use
tags:
  - browser
  - automation
dependencies:
  - playwright
---

# Browser Use

技能详细说明...
```

#### 依赖解析 (`internal/deps/`)

按拓扑顺序解析技能依赖。

#### UI 组件 (`internal/ui/`)

提供进度条和加载动画。

#### 缓存 (`internal/cache/`)

基于时间过期的搜索结果缓存。

- **TTL**: 1小时 (可配置)
- **存储**: 内存 (可持久化)

#### 服务器 (`internal/server/`)

用于 Web UI 和桌面应用的嵌入式 HTTP 服务器。

**结构**:
- `server.go`: 服务器生命周期和路由
- `handlers_skill.go`: 技能管理 API
- `handlers_repo.go`: 仓库管理 API
- `handlers_system.go`: 系统配置 API

#### 服务管理 (`internal/service/`)

服务器的后台进程管理。

**特性**:
- PID 文件管理
- 进程状态检查
- 服务生命周期控制

## 数据流

### 技能安装流程

```mermaid
sequenceDiagram
    participant U as User
    participant C as CLI
    participant G as GitHub
    participant Git as Git Ops
    participant FS as File System
    participant Cfg as Config

    U->>C: ask skill install browser-use
    C->>Cfg: Load config
    C->>G: Resolve skill source
    G-->>C: Repository URL
    C->>Git: Clone repository
    Git->>FS: Download to .agent/skills/
    Git-->>C: Clone complete
    C->>FS: Parse SKILL.md
    FS-->>C: Skill metadata
    C->>Cfg: Update ask.yaml
    C->>Cfg: Add to ask.lock
    Cfg-->>C: Saved
    C-->>U: Installation complete
```

## 文件结构

### 项目布局

```
my-agent-project/
├── ask.yaml              # 项目配置
├── ask.lock              # 版本锁定文件
├── main.py               # 你的 Agent 代码
└── .agent/
    └── skills/           # 已安装技能
        ├── browser-use/
        │   ├── SKILL.md
        │   ├── scripts/
        │   └── references/
        └── web-surfer/
            ├── SKILL.md
            └── ...
```

**Agent 专属路径:**
- **Claude**: `.claude/skills/`
- **Cursor**: `.cursor/skills/`
- **Codex**: `.codex/skills/`

### ASK 安装路径

```
/usr/local/bin/
└── ask                   # 单一二进制文件 (Go 编译)

~/.cache/ask/             # 可选缓存目录
└── search-cache.db       # 搜索结果缓存
```

## 性能优化

1. **并行搜索**: 使用 goroutines 并发扫描多个来源
2. **稀疏检出**: 仅下载所需的子目录
3. **缓存**: 搜索结果缓存 1 小时
4. **单一二进制**: 无运行时依赖，启动快

## 安全考量

### 信任模型

```mermaid
graph TB
    A[User Trusts] --> B[Repository Source]
    B --> C{Verified?}
    C -->|Official| D[anthropics, openai, MCP]
    C -->|Community| E[GitHub Topics]
    D --> F[Higher Trust]
    E --> G[User Verification Needed]
```

**安全实践**:
1. 安装前阅读 `SKILL.md`
2. 审查 `scripts/` 目录内容
3. 检查仓库的 Star 数和活跃度
4. 使用版本锁定确保可复现性
5. 审计依赖项

## 扩展性

### 自定义仓库

用户可以添加自定义源：

```yaml
repos:
  - name: my-team
    type: dir
    url: my-org/internal-skills/skills
```

---

更多详细信息，请参阅：
- [配置指南](configuration.md)
- [SKILL.md 格式规范](skill-format.md)

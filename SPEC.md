# ASK (Agent Skills Kit) - 产品规格说明书

## 项目概述

ASK 是一个用于管理 AI Agent 技能的命令行工具，类似于 Homebrew 管理 macOS 软件包。

### 核心价值
- 让 AI Agent 开发者能够快速发现、安装和管理技能
- 提供标准化的技能格式和元数据规范
- 支持多数据源的技能搜索

---

## 功能需求

### 已实现功能 ✅

| 命令 | 功能 | 状态 |
|------|------|------|
| `ask init` | 初始化项目，创建 `ask.yaml` | ✅ |
| `ask search <keyword>` | 搜索技能（并行多源） | ✅ |
| `ask install <skill>` | 安装技能 | ✅ |
| `ask uninstall <skill>` | 卸载技能 | ✅ |
| `ask list` | 列出已安装技能 | ✅ |
| `ask info <skill>` | 显示技能详情 | ✅ |

### 待实现功能 ⏳

| 命令 | 功能 | 优先级 |
|------|------|--------|
| `ask update [skill]` | 更新技能到最新版本 | P1 |
| `ask install skill@v1.0` | 版本锁定安装 | P1 |
| `ask create <name>` | 创建技能模板 | P2 |

---

## 技能来源规范

### 支持的来源类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `topic` | GitHub 主题搜索 | `agent-skill` |
| `dir` | GitHub 仓库目录 | `anthropics/skills/skills` |

### 默认来源

```yaml
sources:
  - name: community
    type: topic
    url: agent-skill
  - name: anthropics
    type: dir
    url: anthropics/skills/skills
  - name: mcp-servers
    type: dir
    url: modelcontextprotocol/servers/src
```

---

## SKILL.md 规范

每个技能应包含 `SKILL.md` 文件，支持 YAML frontmatter：

```yaml
---
name: browser-use
description: Browser automation for AI agents
version: 1.0.0
author: browser-use
dependencies:
  - playwright
tags:
  - browser
  - automation
---

# Browser Use

技能的详细说明...
```

---

## ask.yaml 规范

项目配置文件格式：

```yaml
version: "1.0"
skills:
  - browser-use      # 已安装技能列表
skills_info:         # 技能元数据
  - name: browser-use
    description: Browser automation
    url: https://github.com/browser-use/browser-use
sources:             # 可选：自定义来源
  - name: custom
    type: dir
    url: owner/repo/path
```

---

## 技术架构

### 目录结构

```
ask/
├── cmd/                  # 命令实现
│   ├── root.go
│   ├── init.go
│   ├── search.go
│   ├── install.go
│   ├── uninstall.go
│   ├── list.go
│   └── info.go
├── internal/
│   ├── config/           # 配置管理
│   ├── github/           # GitHub API 客户端
│   ├── git/              # Git 操作
│   ├── skill/            # SKILL.md 解析
│   └── deps/             # 依赖解析
├── assets/               # 静态资源（logo等）
├── .github/workflows/    # CI/CD
├── Makefile
├── README.md
└── README_zh.md
```

### 性能优化

1. **并行搜索**: 使用 goroutines 并行扫描多个来源
2. **子目录安装**: 克隆到临时目录，复制子目录
3. **默认来源合并**: `LoadConfig` 自动合并新增默认来源

---

## 发布规范

### 版本号
遵循 [Semantic Versioning](https://semver.org/):
- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能新增
- PATCH: 向后兼容的问题修复

### 发布流程
1. 更新 CHANGELOG.md
2. 创建 tag: `git tag v0.1.0`
3. 推送: `git push --tags`
4. GitHub Actions 自动运行 goreleaser

### Homebrew 安装
需要创建 `yeasy/homebrew-tap` 仓库，goreleaser 会自动生成 formula。

---

## 代码规范

### Go 代码风格
- 使用 `gofmt` 格式化
- 使用 `go vet` 检查
- 公共函数需要注释

### 测试要求
- 所有 internal 包需要测试文件
- 测试命令: `make test`
- CI 自动运行测试

### 提交规范
```
<type>: <description>

[optional body]
```

类型: feat, fix, docs, refactor, test, chore

---

## 待办事项

### 高优先级 (P1)
- [ ] `ask update` 命令
- [ ] 版本锁定 (`skill@v1.0`)
- [ ] Homebrew tap 仓库创建

### 中优先级 (P2)
- [ ] Git sparse checkout 优化
- [ ] 搜索结果缓存
- [ ] 进度条显示
- [ ] `ask create` 命令

### 低优先级 (P3)
- [ ] 插件系统
- [ ] 自定义注册表 API
- [ ] 离线模式

---

## 更新日志

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-01-15 | 0.1.0 | 初始版本，基本功能实现 |

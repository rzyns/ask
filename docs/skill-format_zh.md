# SKILL.md 格式

每个技能都应该包含一个 `SKILL.md` 文件，用于描述其用途和配置。本文档解释了该文件的格式。

---

## 基本结构

`SKILL.md` 文件包含两个部分：
1. **YAML Frontmatter** - 结构化元数据
2. **Markdown Body** - 详细说明

```markdown
---
name: my-skill
description: A short description of what this skill does
version: 1.0.0
author: your-name
---

# My Skill

Detailed instructions and documentation go here...
```

---

## Frontmatter 字段

### 必填字段

| 字段 | 类型 | 描述 |
|-------|------|-------------|
| `name` | string | 唯一的技能标识符 |
| `description` | string | 简短描述 (显示在搜索结果中) |

### 可选字段

| 字段 | 类型 | 描述 |
|-------|------|-------------|
| `version` | string | 语义化版本 (例如 `1.0.0`) |
| `author` | string | 技能作者或组织 |
| `license` | string | 许可证标识符 (例如 `MIT`) |
| `tags` | list | 可搜索的标签 |
| `dependencies` | list | 所需的包或工具 |
| `requires` | list | 该技能依赖的其他技能 |

---

## 完整示例

```markdown
---
name: browser-use
description: Browser automation for AI agents using Playwright
version: 1.2.0
author: browser-use
license: MIT
tags:
  - browser
  - automation
  - playwright
  - web
dependencies:
  - playwright
  - python>=3.10
requires:
  - base-agent
---

# Browser Use

A skill that enables AI agents to interact with web browsers.

## Installation

After installing with ASK, run:

\`\`\`bash
pip install playwright
playwright install chromium
\`\`\`

## Usage

\`\`\`python
from browser_use import BrowserAgent

agent = BrowserAgent()
result = await agent.run("Search for weather in Tokyo")
\`\`\`

## Configuration

Set environment variables:

- `BROWSER_HEADLESS`: Run in headless mode (default: true)
- `BROWSER_TIMEOUT`: Page timeout in ms (default: 30000)

## API Reference

### BrowserAgent

Main agent class for browser interactions.

**Methods:**
- `run(task: str)` - Execute a browser task
- `screenshot()` - Capture current page
- `close()` - Close browser instance
```

---

## 创建新技能

使用 `ask create` 生成技能模板：

```bash
ask skill create my-new-skill
```

这将创建：

```
.agent/skills/my-new-skill/
├── SKILL.md           # 技能元数据和文档
├── __init__.py        # Python 模块初始化
└── main.py            # 主要实现
```

---

## 最佳实践

1. **清晰的描述**: 编写简洁、易于搜索的描述
2. **语义化版本控制**: 使用正确的语义化版本
3. **记录依赖项**: 列出所有必需的包
4. **使用示例**: 包含可工作的代码示例
5. **配置文档**: 记录所有环境变量

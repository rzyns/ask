# 技能源

ASK 可以从多个源搜索和安装技能。本文档解释了源的工作原理以及如何配置它们。

---

## 默认源

ASK 附带了六个预配置的源：

| 名称 | 类型 | URL | 描述 |
|------|------|-----|-------------|
| `community` | topic | `agent-skill` | 带有 `agent-skill` 主题的 GitHub 仓库 |
| `anthropics` | dir | `https://github.com/anthropics/skills/tree/main/skills` | Anthropic 官方技能 |
| `scientific` | dir | `https://github.com/K-Dense-AI/claude-scientific-skills/tree/main/scientific-skills` | 科研技能 |
| `superpowers` | dir | `https://github.com/obra/superpowers/tree/main/skills` | 核心技能库 |
| **OpenAI** | Official | `openai` | OpenAI 官方技能 |
| **MATLAB** | Official | `matlab` | MATLAB 官方技能 |
| **Composio** | Community | `composio` | Awesome Claude Skills |
| **SkillHub** | Service | `skills` | SkillHub.club 搜索服务 |
| **Vercel** | AI SDK | `vercel-labs/agent-skills` | Vercel AI SDK 技能 |

---

## 源类型

### 主题源 (`topic`)

主题源在 GitHub 上搜索具有特定主题标签的仓库。

```yaml
repos:
  - name: community
    type: topic
    url: agent-skill   # 要搜索的 GitHub 主题
```

**优点：**
- 自动发现社区维护的技能
- 发布后立即显示新技能

**缺点：**
- 搜索结果取决于 GitHub API 限制
- 社区项目的质量参差不齐

### 目录源 (`dir`)

目录源指向 GitHub 仓库中的特定路径。

```yaml
repos:
  - name: anthropics
    type: dir
    url: anthropics/skills/skills   # owner/repo/path
```

**优点：**
- 精选、一致的质量
- 搜索速度更快（无需 API 查询）
- 适用于组织管理的技能

**缺点：**
- 新技能需要更新仓库

---

## 添加自定义源

编辑您的 `ask.yaml` 以添加自定义源：

```yaml
version: "1.0"
skills:
  - browser-use
repos:
  # 在这里添加您的自定义源
  - name: my-team
    type: dir
    url: my-org/agent-skills/skills
  
  - name: awesome-skills
    type: topic
    url: awesome-agent-skill
```

```bash
ask repo add my-org/agent-skills/skills
```

---

## 查看源

列出所有配置的源：

```bash
ask repo list
```

列出源中可用的技能：

```bash
ask repo list <source-name>
```

---

## 移除源

按名称移除源：

```bash
ask repo remove my-team
```

> **注意：** 默认源始终可用，并将在下次调用 `LoadConfig()` 时重新添加。

---

## 搜索优先级

当您运行 `ask search` 时，ASK 会并行查询所有源并合并结果。如果同一技能存在于多个源中，则会显示这两个条目并指明其来源。

---

## GitHub API 速率限制

基于主题的搜索使用 GitHub Search API，该 API 有速率限制：
- **未验证**：10 次请求/分钟
- **已验证**：30 次请求/分钟

要增加限制，请设置 GitHub 令牌：

```bash
export GITHUB_TOKEN=your_token_here
```

基于目录的源使用 Contents API，其限制更高。

# SKILL.md Format

Every skill should include a `SKILL.md` file that describes its purpose and configuration. This document explains the format.

---

## Basic Structure

A `SKILL.md` file has two parts:
1. **YAML Frontmatter** - Structured metadata
2. **Markdown Body** - Detailed instructions

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

## Frontmatter Fields

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique skill identifier |
| `description` | string | Short description (shown in search) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Semantic version (e.g., `1.0.0`) |
| `author` | string | Skill author or organization |
| `license` | string | License identifier (e.g., `MIT`) |
| `tags` | list | Searchable tags |
| `dependencies` | list | Required packages or tools |
| `requires` | list | Other skills this depends on |

---

## Complete Example

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

## Creating a New Skill

Use `ask create` to generate a skill template:

```bash
ask skill create my-new-skill
```

This creates:

```
.agent/skills/my-new-skill/
├── SKILL.md           # Skill metadata and docs
├── __init__.py        # Python module init
└── main.py            # Main implementation
```

---

## Best Practices

1. **Clear descriptions**: Write concise, searchable descriptions
2. **Semantic versioning**: Use proper semver for versions
3. **Document dependencies**: List all required packages
4. **Usage examples**: Include working code examples
5. **Configuration docs**: Document all environment variables

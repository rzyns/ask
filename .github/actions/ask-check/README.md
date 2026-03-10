# ASK Security Check Action

Scan agent skills for security issues using [ASK](https://github.com/yeasy/ask) in your GitHub Actions workflows.

Detects hardcoded secrets, dangerous commands, network activity, and other security risks in AI agent skill files. Results can be uploaded to GitHub Code Scanning for inline annotations on pull requests.

## Quick Start

```yaml
- uses: yeasy/ask/.github/actions/ask-check@main
```

This will scan the repository root, output SARIF, and upload results to GitHub Code Scanning.

## Usage Examples

### Basic — scan on every push

```yaml
name: Skill Security
on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    permissions:
      security-events: write  # Required for Code Scanning upload
    steps:
      - uses: actions/checkout@v4
      - uses: yeasy/ask/.github/actions/ask-check@main
```

### Scan a specific directory

```yaml
- uses: yeasy/ask/.github/actions/ask-check@main
  with:
    path: './skills'
    severity: 'critical'
```

### JSON report without Code Scanning

```yaml
- uses: yeasy/ask/.github/actions/ask-check@main
  with:
    format: 'json'
    output: 'security-report.json'
    upload-sarif: 'false'
```

### Pin to a specific ASK version

```yaml
- uses: yeasy/ask/.github/actions/ask-check@main
  with:
    version: 'v1.7.5'
```

### Use findings count in later steps

```yaml
- uses: yeasy/ask/.github/actions/ask-check@main
  id: scan
  with:
    fail-on-findings: 'false'

- run: echo "Found ${{ steps.scan.outputs.findings-count }} issues"
```

## Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `path` | Path to skill or project directory | `.` |
| `severity` | Minimum severity: `info`, `warning`, `critical` | `warning` |
| `format` | Output format: `console`, `sarif`, `json`, `html`, `markdown` | `sarif` |
| `output` | Report file path | `ask-check-results.sarif` |
| `version` | ASK version to install | `latest` |
| `upload-sarif` | Upload SARIF to GitHub Code Scanning | `true` |
| `fail-on-findings` | Fail if findings at or above severity | `true` |

## Outputs

| Output | Description |
|--------|-------------|
| `findings-count` | Total number of findings |
| `critical-count` | Number of critical findings |
| `report-path` | Path to the generated report |

## Permissions

To upload SARIF results to GitHub Code Scanning, add:

```yaml
permissions:
  security-events: write
```

## License

Same as [ASK](https://github.com/yeasy/ask) — Apache 2.0.

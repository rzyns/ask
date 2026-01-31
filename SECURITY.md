# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security issues in ASK seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do NOT Open a Public Issue

Security vulnerabilities should not be disclosed publicly until they have been addressed.

### 2. Send a Private Report

Please report security vulnerabilities by emailing:

- **Email**: [Create a security advisory](https://github.com/yeasy/ask/security/advisories/new) on GitHub

Or use GitHub's private vulnerability reporting feature:
1. Go to the [Security tab](https://github.com/yeasy/ask/security) of the repository
2. Click "Report a vulnerability"
3. Fill in the details

### 3. Include Details

Please include the following in your report:

- **Description**: A clear description of the vulnerability
- **Steps to Reproduce**: Detailed steps to reproduce the issue
- **Impact**: What could an attacker achieve with this vulnerability?
- **Affected Versions**: Which versions are affected?
- **Suggested Fix**: If you have one

### 4. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: 24-72 hours
  - High: 1-2 weeks
  - Medium: 2-4 weeks
  - Low: Next release cycle

## Security Best Practices

When using ASK:

1. **Verify Skill Sources**: Only install skills from trusted repositories
2. **Use Security Scanning**: Run `ask check` before installing new skills
3. **Review Skill Code**: Inspect SKILL.md and associated files before installation
4. **Keep Updated**: Regularly run `ask update` to get security patches
5. **Use Lock Files**: Commit `ask.lock` to ensure reproducible installs

## Security Features

ASK includes built-in security features:

- **Entropy Analysis**: Detects potential secrets and API keys
- **Dangerous Command Detection**: Identifies risky shell commands
- **Binary File Scanning**: Flags suspicious executable files
- **HTML Security Reports**: Generate detailed audit reports with `ask check -o report.html`

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who report valid vulnerabilities.

Thank you for helping keep ASK secure! 🛡️

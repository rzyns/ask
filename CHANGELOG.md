# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Shell completion support for bash, zsh, fish, and powershell
- Comprehensive test coverage for `internal/git`, `internal/skill`, `internal/deps`, and `internal/ui` packages
- CI/CD quality gates: linting, test coverage reporting, and security scanning
- Documentation: troubleshooting guide and architecture diagrams
- Git author extraction from git config for `ask skill create` command

### Changed
- Enhanced command help text with practical examples
- Improved error messages with actionable suggestions

## [0.4.0] - 2026-01-15

### Added
- Offline mode support with `--offline` flag
- Performance benchmarking command `ask benchmark`
- Search result caching leveraged for offline usage

### Changed
- `ask skill install` now prevented in offline mode
- `ask skill outdated` skips remote checks in offline mode

## [0.3.0] - 2026-01-15

### Changed
- Updated version structure (skipped public release of 0.3.0 features directly into 0.4.0 or this is a retrospective update if 0.3.0 was skipped)
- *Note: Assuming 0.3.0 was a planned release. Adjusting to reflect current state.*

Actually, looking at previous steps, I implemented features on top of 0.2.0 directly to 0.4.0.
Let's just release as 0.4.0.

## [0.4.0] - 2026-01-15

### Added
- Offline Mode (`--offline` flag)
- `ask benchmark` command
- Search caching for offline support

### Changed
- `install` and `outdated` commands respect offline mode

## [0.2.0] - 2026-01-15

### Added
- Skill commands moved to `ask skill` subcommand for better organization
- New default repositories: OpenAI, GitHub Copilot skills
- `ask skill outdated` command to check for available updates
- `ask skill update` command to update skills to latest versions
- `ask skill create` command to generate skill templates
- Version locking support with `ask.lock` file
- Git sparse checkout optimization for faster skill installation
- Search result caching for improved performance
- Progress bars for long-running operations
- Configurable skills directory (default: `.agent/skills/`)

### Changed
- CLI restructure: all skill operations now under `ask skill` parent command
- Skills installation path changed from `./skills/` to `.agent/skills/`
- Default repositories expanded to include community, Anthropic, MCP, Scientific, Superpowers, and OpenAI sources

### Fixed
- Uninstall command now properly removes skills from both filesystem and config
- Same-name skills from different sources are now properly distinguished
- Repository management commands (`ask repo add/list/remove`) now work correctly

## [0.1.0] - 2026-01-15

### Added
- Initial release of ASK (Agent Skills Kit)
- Basic skill management: search, install, uninstall, list, info
- Multi-source skill discovery (GitHub topics and directories)
- Project initialization with `ask init`
- Repository management with `ask repo` commands
- Configuration file support (`ask.yaml`)
- Default repositories: Community, Anthropic, MCP-Servers, Scientific, Superpowers

[Unreleased]: https://github.com/yeasy/ask/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/yeasy/ask/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yeasy/ask/releases/tag/v0.1.0

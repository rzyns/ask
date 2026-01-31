# Changelog

All notable changes to this project will be documented in this file.

## [1.3.2] - 2026-01-30

 ### Fixed
 - **Build**: Updated GoReleaser config to use `brews` instead of `homebrew_casks` for correct Formula generation.
 - **CI**: Updated `release.yml` to use `libwebkit2gtk-4.1-dev` for compatibility with Ubuntu 24.04 (Noble).

## [1.3.1] - 2026-01-30
 
 ### Fixed
 - **Web UI**: Fixed missing icons in Repositories view by correctly prioritizing GitHub URLs for avatar generation.
 - **Server**: Fixed unused variable lint error in server code.
 
 ## [1.1.3] - 2026-01-26

### Added
- **Security Report Improvements**:
    - **Report Completeness**: HTML reports now list all scanned modules, including "safe" ones, providing a complete audit trail.
    - **Enhanced Visualization**: Clean modules are styled with a light green background and "Safe" badge for quick identification.
    - **Optimized Sorting**: Modules are now sorted by risk level (Critical -> Warning -> Info -> Clean) to prioritize attention on issues.

## [1.1.2] - 2026-01-26

### Added
- **Repo Sync**: Added `--sync` flag to `ask repo add` command for immediate synchronization.
- **Fuzzy Matching**: Enhanced `ask repo list` to support fuzzy matching by URL or `owner/repo` pattern.

### Changed
- **Case Sensitivity**: Fixed an issue where `cmd_test.go` failed due to case sensitivity of agent names.
- **Documentation**: Updated `README.md` and `README_zh.md` to clarify repository sync behavior.

## [1.1.0] - 2026-01-26

### Added
- **Enhanced Security Reports**: 
    - Generated HTML reports now include collapsible Module/Severity sections for better navigation.
    - Added comprehensive Overview dashboard with severity distribution charts.
    - Improved report header with "by ASK" attribution and precise timestamp.
    - Path display optimization: Reports now intelligently show paths (e.g., `~/Projects/...`) instead of just `.` or full absolute paths.
- **Batch Scanning**: Support for batch security scanning of multiple repositories.
- **Documentation**: Added "Security Reports" section to `README.md` and `README_zh.md` with links to live samples.

### Changed
- Refined HTML report CSS for better readability and interactivity.
- Defaulted findings groups to collapsed state for cleaner initial view.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-25

### Added
- **Security Checks**: New `ask check` command to scan skills for secrets, dangerous commands, and suspicious files.
- **Values Reports**: Generate detailed security reports in Markdown, HTML, or JSON with `ask check -o <file>`.
- **Entropy Analysis**: Smart secret detection using Shannon entropy to reduce false positives.

## [1.0.0-rc2] - 2026-01-24

### Added
- **Docker-Style Aliases**: New top-level commands `ask install`, `ask search`, `ask list` for faster access.
- **Install Aliases**: `ask add` (and `ask i`) supported as aliases for `install`.
- **Smart Sync**: `ask search` now automatically initializes the local cache if empty (Lazy Init) and updates it in the background if stale (> 3 days).
- **Auto-Caching on Install**: Installation of a skill from an uncached repository now triggers a background sync.
- **Local Install Optimization**: Installing a skill that exists in the local cache now performs a fast file copy instead of a git clone.
- **Testing**: Added comprehensive unit tests for CLI commands.

### Fixed
- **Panic on Single-Word Install**: Fixed critical panic when using `ask install <name>` with a single word argument.
- **Uninstall Alias**: Added missing top-level `ask uninstall` alias (previously only `ask skill uninstall` worked).
- **Documentation**: Removed invalid `skillhub/skills` repository example and clarified `mcp-builder` installation.
- **Input Validation**: Added input length limits and stricter validation to prevent empty skill name installations from malformed inputs.
- **Robustness**: Improved re-installation check safety.

### Changed
- **Repository Naming**: Local cache directories now use the user-configured repository name (e.g. `anthropics`).
- **Improved UX**: Reduced verbosity of installation commands.
- **Search UI**: Removed `local:` prefix from search results.
- **Documentation**: Updated English and Chinese READMEs with new alias usage.

### Fixed
- **Robust Installation**: Fixed issue where `ask skill install Source/Skill` would fail if the local cache was empty/missing.
- **Index Reliability**: Fixed a bug where repository URLs were not being persisted to `index.json`.

## [0.9.0] - 2026-01-21

### Changed
- **Config Tests**: Fixed and updated `internal/config` tests to align with default repository configuration.
- **Code Quality**: Resolved linter warnings including `io/ioutil` deprecation and unhandled errors.
- **Stability**: general improvements and test coverage updates.

## [0.8.0] - 2026-01-19

### Added
- **SkillHub Integration**: Added support for searching and installing skills from [SkillHub.club](https://www.skillhub.club).
- **Slug Resolution**: intelligent resolution of SkillHub slugs to GitHub install URLs.

## [0.7.6] - 2026-01-19

### Changed
- **Skill Discovery**: Now uses `git clone` for skill discovery from configured repositories, eliminating the need for GitHub tokens for public repositories and avoiding API rate limits.

## [0.7.5] - 2026-01-19

### Added
- **Vercel Skills**: Added `vercel-labs/agent-skills` to default repositories
- **Documentation**: Added CLI demo screenshot, sorted skill sources alphabetically

## [0.7.4] - 2026-01-19

### Added
- **Composio Skills**: Added `ComposioHQ/awesome-claude-skills` to default repositories as `composio`
- **Documentation**: Updated skill sources documentation to include Composio

## [0.7.3] - 2026-01-19

### Added
- **Bulk Skill Installation**: `ask skill install <repo>` now installs all skills from the repository (e.g. `ask skill install superpowers`)
- **MATLAB Skills**: Added official `matlab` repository to default skill sources
- **Documentation**: Added detailed install path documentation for different agents (`.claude`, `.cursor`, etc.) in `README.md` and `README_zh.md`

### Fixed
- Fixed specific config URL handling in `ask skill install` matching logic

## [0.7.2] - 2026-01-17

### Added
- `ask skills` command alias
- GitHub token authentication support (`GITHUB_TOKEN` or `GH_TOKEN`) to avoid rate limit errors

## [0.7.1] - 2026-01-17

### Fixed
- Fixed GoReleaser config to place Homebrew formula in `Formula/` directory

## [0.7.0] - 2026-01-17

### Fixed
- Homebrew formula now uses pre-compiled binaries instead of source compilation
- Users no longer need Go installed to install via Homebrew

## [0.6.1] - 2026-01-16

### Fixed
- Fixed GitHub Actions release workflow to use `GH_PAT` for Homebrew tap updates

## [0.6.0] - 2026-01-16

### Added
- `ask repo list [name]` command to inspect repository skills
- Shell completion support for bash, zsh, fish, and powershell
- Comprehensive test coverage for `internal/git`, `internal/skill`, `internal/deps`, and `internal/ui` packages
- CI/CD quality gates: linting, test coverage reporting, and security scanning
- Documentation: troubleshooting guide and architecture diagrams
- Git author extraction from git config for `ask skill create` command

### Changed
- `ask repo list` now supports optional arguments to list skills in a specific repository
- Enhanced command help text with practical examples
- Improved error messages with actionable suggestions

## [0.5.0] - 2026-01-17

### Added
- **Multi-Tool Support** (`--agent` / `-a` flag)
  - Automatically detects and installs skills for: Claude Code, Cursor, OpenAI Codex, OpenCode
  - Supports installing to multiple agents simultaneously
- **Global Installation Support** (`--global` / `-g` flag)
  - Install skills globally to `~/.ask/skills/` for sharing across all projects
  - Global configuration stored in `~/.ask/config.yaml`
  - Global lock file at `~/.ask/ask.lock`
  - `ask skill list --all` to show both project and global skills
- All skill commands now support `--global` flag: `install`, `uninstall`, `list`, `update`, `outdated`, `info`

### Changed
- Skill commands now display installation scope (project/global) in output messages

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

[Unreleased]: https://github.com/yeasy/ask/compare/v1.3.2...HEAD
[1.3.2]: https://github.com/yeasy/ask/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/yeasy/ask/compare/v1.1.3...v1.3.1
[1.1.3]: https://github.com/yeasy/ask/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/yeasy/ask/compare/v1.1.0...v1.1.2
[1.1.0]: https://github.com/yeasy/ask/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/yeasy/ask/compare/v1.0.0-rc2...v1.0.0
[1.0.0-rc2]: https://github.com/yeasy/ask/compare/v0.9.0...v1.0.0-rc2
[0.9.0]: https://github.com/yeasy/ask/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/yeasy/ask/compare/v0.7.6...v0.8.0
[0.7.6]: https://github.com/yeasy/ask/compare/v0.7.5...v0.7.6
[0.7.5]: https://github.com/yeasy/ask/compare/v0.7.4...v0.7.5
[0.7.4]: https://github.com/yeasy/ask/compare/v0.7.3...v0.7.4
[0.7.3]: https://github.com/yeasy/ask/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/yeasy/ask/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/yeasy/ask/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/yeasy/ask/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/yeasy/ask/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/yeasy/ask/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/yeasy/ask/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/yeasy/ask/compare/v0.2.0...v0.4.0
[0.2.0]: https://github.com/yeasy/ask/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yeasy/ask/releases/tag/v0.1.0

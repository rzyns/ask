# Task 9 Hermes Source Policy Docs/Config Cleanup Plan

> **For Hermes:** Continue strict TDD for config-generation behavior; docs edits are verified by review plus full tests.

**Goal:** Align ASK's generated/recommended Hermes source policy with the lifecycle implementation: user-installable Hermes skills come from the canonical `hermes-index`; bundled Hermes skills under `NousResearch/hermes-agent/skills` are not ASK-managed.

**Architecture:** Treat `hermes-index` as the generated default/recommended Hermes repository. Keep direct `optional-skills` support available for explicit user configuration, but do not recommend or default the bundled `skills` tree. Documentation should explain install/update/uninstall ownership rules briefly enough for users to choose safe commands.

**Tech Stack:** Go config defaults/tests, Markdown docs, existing ASK/Hermes repo-source classifiers.

---

### Task 9.1: Characterize generated default repos

**Objective:** Add tests proving generated defaults include canonical `hermes-index` and exclude bundled Hermes sources.

**Files:**
- Modify: `internal/config/config_test.go`

**Steps:**
1. Write failing assertions in `TestDefaultConfig` / `TestDefaultReposConfiguration`:
   - default repo count increases to include `hermes-index`,
   - `hermes-index` has `type: hermes`,
   - URL is `https://hermes-agent.nousresearch.com/docs/api/skills-index.json`,
   - no default repo URL contains `NousResearch/hermes-agent/skills`.
2. Run `go test ./internal/config -run 'TestDefaultConfig|TestDefaultReposConfiguration' -count=1`; expect failure before implementation.

### Task 9.2: Add canonical default repo

**Objective:** Update generated defaults to recommend/use `hermes-index` without adding bundled Hermes sources.

**Files:**
- Modify: `internal/config/config.go`

**Steps:**
1. Add default repo `{Name: "hermes-index", Type: RepoTypeHermes, URL: "https://hermes-agent.nousresearch.com/docs/api/skills-index.json"}`.
2. Run targeted config tests; expect pass.

### Task 9.3: Update user docs

**Objective:** Explain Hermes source policy and lifecycle safety in docs.

**Files:**
- Modify: `README.md`, `README_zh.md`, `docs/commands.md`, `docs/commands_zh.md`, `docs/configuration.md`, `docs/configuration_zh.md`

**Steps:**
1. Expand Hermes notes: `hermes-index` is canonical; bundled `NousResearch/hermes-agent/skills` is refused/not managed; direct `optional-skills` is only for explicit advanced use.
2. Document Hermes uninstall flags `--forget` and `--delete-files` and safe update behavior.
3. Add Hermes Index to default sources/config examples.

### Task 9.4: Clean live global recommendation config

**Objective:** Remove the unsafe bundled source from the user's live global ASK config while preserving explicitly configured `hermes-index` and `hermes-optional`.

**Files:**
- Modify: `/home/openclaw/.ask/config.yaml` after creating a timestamped backup beside it.

**Steps:**
1. Backup `/home/openclaw/.ask/config.yaml`.
2. Remove only the `hermes-bundled` repo pointing at `NousResearch/hermes-agent/skills`.
3. Preserve `hermes-index` and `hermes-optional`.

### Task 9.5: Review, verify, and commit

**Objective:** Ensure docs/config match policy and tests pass.

**Steps:**
1. Run `gofmt` on Go files.
2. Run targeted config tests and full `go test ./...`.
3. Delegate final docs/config policy review.
4. Update `memory/2026-04-29.md` with debrief.
5. Commit tracked changes with `docs: clarify hermes skill source policy`.

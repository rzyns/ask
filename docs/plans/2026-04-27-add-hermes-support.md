# Add Hermes Support Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Add `hermes` as a first-class ASK target agent so users can install, list, initialize, and restore skills for Hermes via the same ASK workflows used for Claude, Cursor, Codex, OpenClaw, and other supported agents.

**Architecture:** ASK centralizes agent support in `internal/config/agents.go`; installer, init, list, completion, and detection behavior derive from that registry. Hermes support should therefore be implemented as a registry addition plus targeted tests and docs. The only domain-specific caveat is that Hermes loads `$HERMES_HOME/skills` by default, while ASK's project-local convention would install to `.hermes/skills`; docs must explain that project-local Hermes skills require configuring Hermes `skills.external_dirs` or equivalent runtime setup.

**Tech Stack:** Go 1.26.2 available at `/home/linuxbrew/.linuxbrew/bin/go` in this environment; repo declares `go 1.25.0` in `go.mod`. Cobra CLI, YAML config, SKILL.md markdown/frontmatter conventions.

---

## Current discovery state

- Repository: `/home/openclaw/dev/ask`
- GitNexus: usable; `npx gitnexus analyze --skip-agents-md` indexed the repo successfully during planning.
- GitNexus cleanup: `.gitignore` was restored and `.gitnexus` / `.claude` artifacts were removed after discovery.
- Git status before writing this plan was clean.
- Go is installed, but not on the default Hermes shell `PATH`; use `/home/linuxbrew/.linuxbrew/bin/go` or prepend `/home/linuxbrew/.linuxbrew/bin`.

## Recommended Hermes target layout

```text
Agent key:   hermes
Display:     Hermes
Alias:       hermes-agent
Project dir: .hermes/skills
Global dir:  ~/.hermes/skills
```

### Important caveat

Hermes' built-in local skills directory is `$HERMES_HOME/skills`, defaulting to `~/.hermes/skills`. Hermes also supports external skill directories via config:

```yaml
skills:
  external_dirs:
    - /absolute/path/to/project/.hermes/skills
```

Therefore:

- `ask install <skill> --agent hermes --global` should immediately install to the default Hermes skills directory for the current user/profile root.
- `ask install <skill> --agent hermes` should install to `.hermes/skills` consistently with ASK's project-local agent model, but docs should state that Hermes must be configured to scan that project-local directory before those skills are automatically loaded.

---

## Task 0: Verify baseline toolchain and tests

**Objective:** Establish that the repo can be tested before code changes.

**Files:** none

**Step 1: Verify Go**

Run:

```bash
cd /home/openclaw/dev/ask
/home/linuxbrew/.linuxbrew/bin/go version
```

Expected: prints Go version, currently observed as `go1.26.2 linux/amd64`.

**Step 2: Run targeted baseline tests**

Run:

```bash
cd /home/openclaw/dev/ask
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config ./cmd
```

Expected: pass before implementation. If not, diagnose and record baseline failures before making Hermes changes.

---

## Task 1: Add Hermes to the centralized agent registry

**Objective:** Make `hermes` and alias `hermes-agent` valid supported agents.

**Files:**
- Modify: `internal/config/agents.go`

**Step 1: Add the AgentType constant**

Add near the existing `AgentType` constants:

```go
// AgentHermes represents the Hermes Agent
AgentHermes AgentType = "hermes"
```

**Step 2: Add SupportedAgents entry**

Add to `SupportedAgents`:

```go
AgentHermes: {
    Name:       "Hermes",
    ProjectDir: ".hermes/skills",
    GlobalDir:  ".hermes/skills",
    Aliases:    []string{"hermes-agent"},
},
```

**Step 3: Format**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/gofmt -w internal/config/agents.go
```

**Step 4: Verify targeted config tests still compile**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config
```

Expected: existing tests pass or only fail because new Hermes-specific assertions have not been added yet.

---

## Task 2: Add Hermes registry tests

**Objective:** Prove Hermes appears in supported names, validates directly and via alias, resolves to `AgentHermes`, and maps to correct directories.

**Files:**
- Modify: `internal/config/agents_test.go`

**Step 1: Update known agents list**

Include `hermes` in `TestGetSupportedAgentNames`:

```go
knownAgents := []string{"claude", "cursor", "codex", "gemini", "copilot", "hermes"}
```

**Step 2: Add validation cases**

In `TestIsValidAgent`, add:

```go
{"direct agent hermes", "hermes", true},
{"alias hermes-agent", "hermes-agent", true},
```

**Step 3: Add resolution cases**

In `TestResolveAgentType`, add:

```go
{"direct name hermes", "hermes", AgentHermes, true},
{"alias hermes-agent", "hermes-agent", AgentHermes, true},
```

**Step 4: Add directory cases**

In `TestGetAgentSkillsDir`, add:

```go
t.Run("project dir for hermes", func(t *testing.T) {
    dir, err := GetAgentSkillsDir(AgentHermes, false)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if dir != ".hermes/skills" {
        t.Errorf("Expected '.hermes/skills', got %q", dir)
    }
})

t.Run("global dir for hermes", func(t *testing.T) {
    dir, err := GetAgentSkillsDir(AgentHermes, true)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    expected := filepath.Join(home, ".hermes/skills")
    if dir != expected {
        t.Errorf("Expected %q, got %q", expected, dir)
    }
})
```

**Step 5: Optionally assert all-dir inclusion**

Extend `TestGetAllAgentSkillsDirs` to assert `.hermes/skills` and `~/.hermes/skills` are present.

**Step 6: Format and test**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/gofmt -w internal/config/agents_test.go
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config
```

Expected: pass.

---

## Task 3: Add Hermes detection coverage

**Objective:** Prove ASK detects Hermes in a project when `.hermes` exists.

**Files:**
- Modify: `internal/config/config_test.go` or `internal/config/agents_test.go` depending on existing test organization.

**Step 1: Add detection test**

```go
func TestDetectExistingToolDirsDetectsHermes(t *testing.T) {
    tmp := t.TempDir()
    if err := os.MkdirAll(filepath.Join(tmp, ".hermes"), 0755); err != nil {
        t.Fatalf("failed to create .hermes dir: %v", err)
    }

    detected := DetectExistingToolDirs(tmp)

    found := false
    for _, target := range detected {
        if target.Name == "hermes" {
            found = true
            if target.SkillsDir != ".hermes/skills" {
                t.Errorf("Hermes skills dir = %q, want .hermes/skills", target.SkillsDir)
            }
            if !target.Enabled {
                t.Error("Hermes target should be enabled")
            }
        }
    }

    if !found {
        t.Fatal("expected Hermes to be detected")
    }
}
```

**Step 2: Format and test**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/gofmt -w internal/config/config_test.go internal/config/agents_test.go
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config
```

Expected: pass.

---

## Task 4: Verify CLI validation/completion path

**Objective:** Ensure CLI-facing paths accept Hermes without bespoke installer changes.

**Files:**
- Modify: `cmd/install_test.go` or another existing command test file if appropriate.

**Step 1: Add a direct validation/target test**

```go
func TestHermesAgentTargetIsValid(t *testing.T) {
    if !config.IsValidAgent("hermes") {
        t.Fatal("hermes should be a valid agent")
    }

    agentType, ok := config.ResolveAgentType("hermes")
    if !ok {
        t.Fatal("hermes should resolve")
    }

    dir, err := config.GetAgentSkillsDir(agentType, false)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if dir != ".hermes/skills" {
        t.Fatalf("Hermes project dir = %q, want .hermes/skills", dir)
    }
}
```

If an existing test already checks `completeAgentNames`, add `hermes` to the expected suggestions there instead.

**Step 2: Test command package**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/gofmt -w cmd/install_test.go
/home/linuxbrew/.linuxbrew/bin/go test ./cmd
```

Expected: pass.

---

## Task 5: Update CLI help and examples

**Objective:** Make Hermes discoverable in help text.

**Files:**
- Modify: `cmd/install.go`
- Maybe modify: `cmd/list.go`, `cmd/lock_install.go`, `cmd/quickstart.go` if grep finds manual agent examples.

**Step 1: Update install command long help**

Change:

```text
Use --agent (-a) to specify target agents (e.g., claude, cursor, codex).
```

To:

```text
Use --agent (-a) to specify target agents (e.g., claude, cursor, codex, hermes).
```

**Step 2: Add Hermes examples**

Add examples such as:

```text
ask skill install pdf --agent hermes
ask skill install pdf --agent hermes --global
```

**Step 3: Update flag description**

Change:

```go
cmd.Flags().StringSliceP("agent", "a", []string{}, "Target agent(s) to install for (e.g. claude, cursor)")
```

To:

```go
cmd.Flags().StringSliceP("agent", "a", []string{}, "Target agent(s) to install for (e.g. claude, cursor, hermes)")
```

**Step 4: Format and test**

Run:

```bash
/home/linuxbrew/.linuxbrew/bin/gofmt -w cmd/install.go cmd/list.go cmd/lock_install.go cmd/quickstart.go
/home/linuxbrew/.linuxbrew/bin/go test ./cmd
```

Expected: pass.

---

## Task 6: Update README and docs

**Objective:** User-facing docs should advertise Hermes and accurately explain the project-local caveat.

**Files:**
- Modify: `README.md`
- Modify: `README_zh.md`
- Search and modify relevant `docs/*.md`

**Step 1: Update supported-agent count**

Current docs mention `19 Agents`; after Hermes this should be `20 Agents`.

Search:

```bash
grep -R "19 Agents\|19 agents\|Claude, Cursor\|OpenClaw" -n README.md README_zh.md docs || true
```

Use `search_files` rather than shell grep if working through Hermes tools.

**Step 2: Update English README wording**

Change the supported list to include Hermes, for example:

```text
Install once — works with Claude, Cursor, Codex, Copilot, Windsurf, Gemini, Hermes, OpenClaw, and 12 more.
```

Change:

```text
19 Agents, One CLI
```

To:

```text
20 Agents, One CLI
```

**Step 3: Update Chinese README equivalently**

Mirror the count/list changes in `README_zh.md`.

**Step 4: Add directory docs if a suitable section exists**

Add:

```text
Hermes
  Project: .hermes/skills
  Global:  ~/.hermes/skills
  Alias:   hermes-agent
```

Add caveat:

```text
Hermes loads skills from $HERMES_HOME/skills by default. Project-local ASK installs to .hermes/skills may require adding that path to Hermes `skills.external_dirs`.
```

**Step 5: Verify docs-only changes do not affect tests**

Run at least:

```bash
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config ./cmd
```

Expected: pass.

---

## Task 7: Full verification

**Objective:** Validate the implementation end-to-end.

**Step 1: Targeted tests**

```bash
cd /home/openclaw/dev/ask
/home/linuxbrew/.linuxbrew/bin/go test ./internal/config ./cmd
/home/linuxbrew/.linuxbrew/bin/go test ./internal/installer ./internal/skill ./internal/cache
```

Expected: pass or record any unrelated pre-existing failures.

**Step 2: Full test suite**

```bash
/home/linuxbrew/.linuxbrew/bin/go test ./...
```

Expected: pass. If network/live tests fail, record exact package/test names and reason.

**Step 3: Manual smoke test in temp dir**

```bash
tmpdir="$(mktemp -d)"
cd "$tmpdir"
mkdir -p .hermes
/home/openclaw/dev/ask/ask init --yes
/home/openclaw/dev/ask/ask list --agent hermes
/home/openclaw/dev/ask/ask list --agent hermes --global
```

Expected:

- `ask init --yes` reports Hermes detected if `.hermes` exists.
- `ask list --agent hermes` checks `.hermes/skills`.
- `ask list --agent hermes --global` checks `~/.hermes/skills`.

If no built binary exists or is stale, build first:

```bash
cd /home/openclaw/dev/ask
/home/linuxbrew/.linuxbrew/bin/go build -o ask .
```

---

## Task 8: Final review before commit

**Objective:** Ensure the patch is minimal and correct.

**Step 1: Review diff**

```bash
git diff --stat
git diff -- internal/config/agents.go internal/config/agents_test.go cmd README.md README_zh.md docs
```

**Step 2: Check git status**

```bash
git status --short
```

**Step 3: Suggested commit**

```bash
git add internal/config/agents.go internal/config/*_test.go cmd README.md README_zh.md docs/plans/2026-04-27-add-hermes-support.md docs
git commit -m "feat: add Hermes agent support"
```

Only commit after tests and manual smoke verification are complete.

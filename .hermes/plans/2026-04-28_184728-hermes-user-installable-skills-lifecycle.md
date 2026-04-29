# Hermes User-Installable Skills Lifecycle Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Extend ASK so it can safely list, install, import/adopt, update, and uninstall user-installable Hermes skills while explicitly refusing to manage Hermes bundled/built-in skills.

**Architecture:** Treat ASK as the package/lifecycle manager and Hermes as the runtime. ASK should manage only user-installable Hermes skills with explicit ownership/provenance, record that ownership in config/lock metadata, and leave bundled/runtime-owned Hermes skills out of scope. Prefer the Hermes aggregate index (`type: hermes`) as the canonical official source; remove direct support/recommendations for `NousResearch/hermes-agent/skills`, and avoid direct `optional-skills` duplication unless deliberately kept as a compatibility path.

**Tech Stack:** Go, Cobra CLI, ASK config/lock YAML, existing `internal/installer`, `internal/repository`, `internal/config`, `internal/skill` packages, filesystem symlink/copy helpers, Go tests.

---

## Design Decisions / Non-Negotiables

1. **ASK manages user-installable skills only.**
   - In scope: ASK-managed installs, imported/adopted user skills, official optional Hermes skills, third-party GitHub-backed skills.
   - Out of scope: bundled/built-in Hermes runtime skills.

2. **Do not manage `NousResearch/hermes-agent/skills`.**
   - This path represents Hermes bundled/core runtime distribution.
   - ASK should not search/install/update/uninstall/import those as packages.
   - If users configure this repo, warn or skip rather than silently treating it as user-installable.

3. **Prefer `hermes-index` over direct Hermes repo directory sources.**
   - Canonical source:
     ```yaml
     repos:
       - name: hermes-index
         type: hermes
         url: https://hermes-agent.nousresearch.com/docs/api/skills-index.json
     ```
   - `optional-skills` may remain supported as a compatibility source, but `hermes-index` should be the recommended path to preserve source identity and avoid duplicates.

4. **Provenance controls lifecycle power.**
   - Strong provenance: install/update/uninstall allowed.
   - Weak provenance: import/adopt allowed as local-only, but update unavailable.
   - ASK must not infer same-name skills are equivalent without strong evidence.

5. **Uninstall must respect ownership.**
   - ASK-owned symlink/cache installs can be removed normally.
   - Imported in-place Hermes-native directories should default to non-destructive `--forget` behavior unless the user explicitly requests file deletion.

6. **Do not implement Hermes runtime operations.**
   - No enable/disable, platform activation, session loading, or Hermes runtime config.
   - Those remain Hermes commands, e.g. `hermes skills config`, `/skill`, etc.

---

## Current Code Context

Relevant existing files inspected:

- `cmd/skill.go`
  - Parent `ask skill` command.
  - Currently includes `check` and `prompt` directly; lifecycle subcommands are registered from their own files.

- `cmd/list.go`
  - `ask skill list --agent hermes` currently scans Hermes skill directories via `config.GetAgentSkillsDir(config.AgentHermes, global)`.
  - It lists directories and parses `SKILL.md`, but does not distinguish ASK-managed vs Hermes-native vs bundled vs imported.

- `cmd/install.go`
  - `ask skill install ... --agent hermes --global` already installs into ASK central storage then symlinks/copies into Hermes skills dir.
  - Current config loading still uses `config.LoadConfig()` in places and should be reviewed for `--global` / `--config` consistency.

- `cmd/update.go`
  - Current update is generic and git-pull-based; it does not understand Hermes ownership/provenance or ASK-managed symlink installs well enough.

- `cmd/uninstall.go`
  - Current uninstall removes target dirs and optionally source with `--all`; needs safer ownership semantics for Hermes imports/native installs.

- `internal/config/lock.go`
  - `LockEntry` currently has `Name`, `Source`, `URL`, `Commit`, `Version`, `InstalledAt`.
  - Needs metadata for agent, ownership, install path, target path(s), source identifier, checksum, and update strategy.

- `internal/config/agents.go`
  - Hermes global skills dir honors `HERMES_HOME/skills` when `HERMES_HOME` is set, else `~/.hermes/skills`.
  - This behavior is good and should remain.

- `internal/installer/installer.go`
  - Installs central copy under ASK storage and links/copies to agent target dirs.
  - Already updates config and lockfile, but lockfile lacks enough lifecycle metadata.

- `internal/repository/hermes_index.go`
  - Parses Hermes aggregate index and maps official entries to `NousResearch/hermes-agent/optional-skills/...`.
  - Needs explicit bundled/core rejection and possibly richer metadata propagation into install/lock.

---

## Target UX

### Search / repositories

Recommended global config should include only the canonical Hermes aggregate index:

```yaml
repos:
  - name: hermes-index
    type: hermes
    url: https://hermes-agent.nousresearch.com/docs/api/skills-index.json
```

ASK should reject or warn for direct bundled source:

```yaml
repos:
  - name: hermes-bundled
    type: dir
    url: NousResearch/hermes-agent/skills
```

Expected warning text:

```text
Repository NousResearch/hermes-agent/skills contains Hermes bundled/runtime skills.
ASK does not manage bundled Hermes skills. Use hermes-index for user-installable Hermes skills.
```

### List

```bash
ask skill list --agent hermes --global
```

Should show installed Hermes skills with ownership/status:

```text
NAME                  STATUS       MANAGED BY  SOURCE        UPDATE
----                  ------       ----------  ------        ------
gitnexus-explorer     installed    ask         hermes-index  current
discrawl              installed    ask         github        current
local-debug-skill     installed    hermes      local         unavailable
```

Optional later flag:

```bash
ask skill list --agent hermes --global --include-bundled
```

If implemented, bundled rows must be read-only:

```text
honcho                bundled      hermes      builtin       n/a
```

For MVP, omit bundled entirely and possibly include a summary count:

```text
Skipped bundled Hermes skills: 12
```

### Import/adopt

```bash
ask skill import --agent hermes --global --dry-run
ask skill import --agent hermes --global --all
ask skill import --agent hermes --global gitnexus-explorer
```

Dry run should classify:

```text
NAME                  PATH                                      CLASSIFICATION     ACTION
gitnexus-explorer     ~/.hermes/skills/gitnexus-explorer         already-managed    skip
local-debug-skill     ~/.hermes/skills/local-debug-skill         local-only         import as local
bundled/foo           <bundled source path>                      bundled            skip
```

Default import should record local-only skills as managed-by-ASK metadata without claiming updateability.

### Install

```bash
ask skill install gitnexus-explorer --agent hermes --global
ask skill install hermes-index/gitnexus-explorer --agent hermes --global
```

Should:

1. Resolve through ASK cache/Hermes index.
2. Reject bundled/core source paths.
3. Install into ASK-owned storage.
4. Symlink/copy into `$HERMES_HOME/skills`.
5. Record provenance and ownership in lockfile.

### Update

```bash
ask skill update --agent hermes --global
ask skill update gitnexus-explorer --agent hermes --global
```

Should:

- update ASK-owned installable skills with known source,
- skip local-only imports with a clear reason,
- skip/refuse bundled/core skills,
- detect dirty local modifications before overwriting.

### Uninstall

ASK-owned install:

```bash
ask skill uninstall gitnexus-explorer --agent hermes --global
```

Should remove Hermes target symlink/copy, ASK central copy, config entry, and lock entry.

Imported in-place local skill:

```bash
ask skill uninstall local-debug-skill --agent hermes --global
```

Should refuse destructive deletion by default:

```text
local-debug-skill was imported in-place and was not installed by ASK.
Use --forget to remove it from ASK tracking, or --delete-files to remove the Hermes directory.
```

---

## Lockfile / Metadata Design

### Extend `config.LockEntry`

Modify `internal/config/lock.go`.

Current:

```go
type LockEntry struct {
    Name        string    `yaml:"name"`
    Source      string    `yaml:"source,omitempty"`
    URL         string    `yaml:"url"`
    Commit      string    `yaml:"commit,omitempty"`
    Version     string    `yaml:"version,omitempty"`
    InstalledAt time.Time `yaml:"installed_at"`
}
```

Proposed additive fields:

```go
type LockEntry struct {
    Name             string    `yaml:"name"`
    Agent            string    `yaml:"agent,omitempty"`
    Source           string    `yaml:"source,omitempty"`
    SourceIdentifier string    `yaml:"source_identifier,omitempty"`
    URL              string    `yaml:"url"`
    Commit           string    `yaml:"commit,omitempty"`
    Version          string    `yaml:"version,omitempty"`
    InstalledAt      time.Time `yaml:"installed_at"`

    Ownership        string    `yaml:"ownership,omitempty"`        // ask, imported, hermes, bundled
    InstallMode      string    `yaml:"install_mode,omitempty"`      // ask-cache, in-place, symlink, copy
    UpdateStrategy   string    `yaml:"update_strategy,omitempty"`   // git, hermes-index, none
    TargetPath       string    `yaml:"target_path,omitempty"`
    SourcePath       string    `yaml:"source_path,omitempty"`
    Checksum         string    `yaml:"checksum,omitempty"`
}
```

Keep fields additive for backward compatibility with existing lockfiles.

### Identity rules

- Physical identity: `agent + target_path`.
- Package identity when available: `agent + source_identifier`.
- Display name: `name`.
- Do not rely on `name` alone for destructive operations when multiple paths can contain the same name.

---

## Implementation Tasks

### Task 1: Add Hermes source classification helpers

**Objective:** Centralize the bundled-vs-user-installable policy.

**Files:**
- Create: `internal/hermes/skills.go` or `internal/config/hermes_skills.go`.
- Test: `internal/hermes/skills_test.go` or adjacent package test.

**Step 1: Write failing tests**

Test cases:

```go
func TestClassifyHermesSourceRejectsBundledSkillsRepo(t *testing.T) {
    got := ClassifyHermesSource("NousResearch/hermes-agent/skills")
    require.Equal(t, HermesSourceBundled, got.Kind)
    require.False(t, got.Manageable)
}

func TestClassifyHermesSourceAllowsOptionalSkillsRepo(t *testing.T) {
    got := ClassifyHermesSource("NousResearch/hermes-agent/optional-skills")
    require.Equal(t, HermesSourceOfficialOptional, got.Kind)
    require.True(t, got.Manageable)
}

func TestClassifyHermesSourceAllowsHermesIndex(t *testing.T) {
    got := ClassifyHermesSource("https://hermes-agent.nousresearch.com/docs/api/skills-index.json")
    require.Equal(t, HermesSourceIndex, got.Kind)
    require.True(t, got.Manageable)
}
```

Also cover HTTPS GitHub URLs:

```text
https://github.com/NousResearch/hermes-agent/tree/main/skills/foo          => bundled/unmanageable
https://github.com/NousResearch/hermes-agent/tree/main/optional-skills/foo => optional/manageable
```

**Step 2: Run test to verify failure**

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/hermes ./internal/config ./internal/repository
```

Expected: fail because helper package/functions do not exist.

**Step 3: Implement helper**

Add types conceptually like:

```go
type HermesSourceKind string

const (
    HermesSourceUnknown          HermesSourceKind = "unknown"
    HermesSourceIndex            HermesSourceKind = "index"
    HermesSourceOfficialOptional HermesSourceKind = "official-optional"
    HermesSourceBundled          HermesSourceKind = "bundled"
)

type HermesSourceClassification struct {
    Kind       HermesSourceKind
    Manageable bool
    Reason     string
}
```

Implement normalization for:

- bare `NousResearch/hermes-agent/skills...`,
- bare `NousResearch/hermes-agent/optional-skills...`,
- GitHub browser URLs,
- `.git` suffixes,
- `https://github.com/NousResearch/hermes-agent/tree/<branch>/...`.

**Step 4: Verify**

```bash
go test ./internal/hermes ./internal/config ./internal/repository
```

Expected: pass.

**Step 5: Commit**

```bash
git add internal/hermes internal/config internal/repository
git commit -m "refactor: classify hermes skill sources"
```

---

### Task 2: Reject bundled Hermes repo sources in repo add/search/fetch paths

**Objective:** Prevent `NousResearch/hermes-agent/skills` from being treated as a normal user-installable skill repository.

**Files:**
- Modify: `cmd/repo.go`
- Modify: `internal/repository/source.go` or relevant fetch dispatcher
- Modify: `internal/repository/hermes_index.go`
- Tests: `cmd/repo_test.go` or new `cmd/repo_hermes_test.go`; `internal/repository/hermes_index_test.go`

**Step 1: Write failing tests**

Test expected behavior:

1. `repo add hermes-bundled NousResearch/hermes-agent/skills --type dir` warns/refuses.
2. Fetching/searching a configured bundled source returns a clear unsupported-source error or zero candidates with warning.
3. Hermes index entries resolving to `NousResearch/hermes-agent/skills/...` are skipped.

Example assertions:

```go
require.Error(t, err)
require.Contains(t, err.Error(), "bundled Hermes skills")
```

For index conversion:

```go
skills := []hermesIndexSkill{{
    Name: "core-skill",
    ResolvedGitHubID: "NousResearch/hermes-agent/skills/core-skill",
}}
candidates := hermesIndexSkillsToCandidates(skills, "")
require.Empty(t, candidates)
```

**Step 2: Implement bundled rejection**

Use the helper from Task 1.

- In repo add: refuse or warn-and-skip. Prefer refusal for explicit add.
- In repository fetch/search: treat bundled source as unsupported.
- In Hermes index mapping: if resolved path points under `hermes-agent/skills`, skip.

**Step 3: Verify**

```bash
go test ./cmd ./internal/repository ./internal/hermes
```

Expected: pass.

**Step 4: Commit**

```bash
git add cmd internal/repository internal/hermes
git commit -m "fix: exclude bundled hermes skills from ask sources"
```

---

### Task 3: Add Hermes installed-skill scanner

**Objective:** Discover user-visible Hermes skill directories and classify them without mutating anything.

**Files:**
- Create: `internal/hermes/installed.go`
- Test: `internal/hermes/installed_test.go`

**Step 1: Write failing tests**

Construct temp Hermes skills tree:

```text
skills/
  gitnexus-explorer/SKILL.md
  research/gitnexus-explorer/SKILL.md
  .hub/taps.json
  category-without-skill/child/SKILL.md
```

Test scanner:

- finds nested `SKILL.md`,
- ignores `.hub`, hidden dirs, and non-skill directories,
- handles symlinks safely,
- returns name, path, relative path, metadata description/version,
- classifies lockfile-backed entries as ASK-managed if matching lock metadata exists,
- marks unknown entries as Hermes-native/local-only.

Example shape:

```go
type InstalledHermesSkill struct {
    Name        string
    Description string
    Version     string
    Path        string
    RelativePath string
    Ownership   string // ask, imported, hermes-native, bundled
    Managed     bool
    UpdateStrategy string
}
```

**Step 2: Implement scanner**

Use `skill.FindSkillMD` / `skill.ParseSkillMD` where possible.

Important safety rules:

- Do not follow symlinks outside expected path unless only reading target metadata safely.
- Ignore dot directories.
- Limit recursion depth reasonably, e.g. 4 or 5.
- Do not infer updateability from name alone.

**Step 3: Verify**

```bash
go test ./internal/hermes ./internal/skill
```

Expected: pass.

**Step 4: Commit**

```bash
git add internal/hermes
git commit -m "feat: scan installed hermes skills"
```

---

### Task 4: Upgrade `ask skill list --agent hermes` output

**Objective:** Show native Hermes installed skills with ownership/provenance status while omitting bundled skills by default.

**Files:**
- Modify: `cmd/list.go`
- Test: `cmd/list_test.go` or new `cmd/list_hermes_test.go`

**Step 1: Write failing tests**

Test human and JSON output for Hermes agent:

```bash
ask skill list --agent hermes --global
```

Expected columns include ownership/status. JSON should include fields like:

```json
{
  "name": "gitnexus-explorer",
  "agent": "hermes",
  "managed_by": "ask",
  "status": "installed",
  "source": "hermes-index",
  "update": "current",
  "path": "..."
}
```

Test bundled omission:

- scanner returns bundled item,
- default list omits it,
- optional `--include-bundled` includes it as read-only if that flag is implemented in this task.

**Step 2: Implement Hermes-specific list path**

In `runList`, when `--agent hermes` is specified, route through the Hermes scanner instead of the generic one-level `showAgentSkills` function.

Add fields to `SkillListItem` additively:

```go
ManagedBy string `json:"managed_by,omitempty"`
Status    string `json:"status,omitempty"`
Source    string `json:"source,omitempty"`
Update    string `json:"update,omitempty"`
```

Do not break non-Hermes list behavior.

**Step 3: Verify**

```bash
go test ./cmd ./internal/hermes
```

Manual check:

```bash
go build -o /tmp/ask-hermes-lifecycle .
/tmp/ask-hermes-lifecycle skill list --agent hermes --global
/tmp/ask-hermes-lifecycle skill list --agent hermes --global --json | python3 -m json.tool >/dev/null
```

**Step 4: Commit**

```bash
git add cmd/list.go cmd/list_test.go internal/hermes
git commit -m "feat: show hermes skill ownership in list"
```

---

### Task 5: Add `ask skill import --agent hermes`

**Objective:** Allow ASK to adopt existing user-installed Hermes skills into its lockfile without claiming unsafe updateability.

**Files:**
- Create: `cmd/import.go` or `cmd/skill_import.go`
- Create: `internal/hermes/import.go`
- Modify: `cmd/skill.go` to register import command if needed
- Modify: `internal/config/lock.go` for additive metadata helpers
- Tests: `cmd/import_test.go`, `internal/hermes/import_test.go`, `internal/config/lock_test.go`

**Step 1: Write failing tests**

Cases:

1. `--dry-run` reports importable local skill and does not write lockfile.
2. `--all` records local-only imported skill with:
   - `Agent: hermes`,
   - `Ownership: imported`,
   - `InstallMode: in-place`,
   - `UpdateStrategy: none`,
   - `TargetPath`,
   - checksum.
3. ASK-managed skill already in lockfile is skipped.
4. Bundled skill is skipped and not written.
5. Named import imports only requested skill.

**Step 2: Implement command flags**

```go
importCmd.Flags().StringSliceP("agent", "a", []string{}, "target agent(s)")
importCmd.Flags().Bool("dry-run", false, "show what would be imported without writing")
importCmd.Flags().Bool("all", false, "import all eligible skills")
importCmd.Flags().Bool("global", false, "use global agent skill directory") // root persistent global may already exist
```

Decision: for MVP, require `--agent hermes`; if omitted, print a helpful error.

**Step 3: Add checksum helper**

Add deterministic tree checksum for `SKILL.md` directory contents.

Rules:

- sort paths,
- skip `.git`, `.env`, hidden cache dirs if appropriate,
- include file mode? probably no for MVP,
- hash relative path + content.

**Step 4: Implement lockfile metadata helpers**

Add helper methods such as:

```go
func (l *LockFile) GetEntryForAgent(name, agent string) *LockEntry
func (l *LockFile) AddOrUpdateEntry(entry LockEntry)
```

Preserve existing `GetEntry(name)` for backward compatibility.

**Step 5: Verify**

```bash
go test ./cmd ./internal/hermes ./internal/config
```

Manual dry-run:

```bash
/tmp/ask-hermes-lifecycle skill import --agent hermes --global --dry-run
```

**Step 6: Commit**

```bash
git add cmd internal/hermes internal/config
git commit -m "feat: import installed hermes skills"
```

---

### Task 6: Record Hermes-specific ownership/provenance during install

**Objective:** Ensure `ask skill install ... --agent hermes` writes enough metadata for future update/uninstall decisions.

**Files:**
- Modify: `internal/installer/installer.go`
- Modify: `internal/repository/source.go` / `SkillCandidate` types if needed
- Modify: `internal/repository/hermes_index.go`
- Tests: `internal/installer/installer_test.go`, `internal/repository/hermes_index_test.go`

**Step 1: Write failing tests**

Test installing for Hermes writes lock entry with:

```yaml
agent: hermes
ownership: ask
install_mode: ask-cache
update_strategy: hermes-index # or git, depending on source
source_identifier: official/research/gitnexus-explorer
source_path: ~/.ask/skills/gitnexus-explorer
 target_path: ~/.hermes/skills/gitnexus-explorer
checksum: ...
```

Exact paths in tests should use temp dirs / `HERMES_HOME`.

**Step 2: Propagate source metadata**

Current `Install(input, opts)` mostly knows URL/subDir/name but not Hermes index identifier. Options:

- Add source metadata to `InstallOptions` only when command-level resolver has it.
- Or encode canonical install input such that installer can infer source.
- Prefer additive struct field:

```go
type InstallSourceMetadata struct {
    Source           string
    SourceIdentifier string
    UpdateStrategy   string
}

type InstallOptions struct {
    ...
    SourceMetadata *InstallSourceMetadata
}
```

If no metadata exists, fall back to existing behavior.

**Step 3: Reject bundled source before install**

Before clone/copy, classify resolved source URL/path. If bundled, return an error.

**Step 4: Save lock metadata**

When `opts.Agents` includes Hermes, set agent-specific metadata.

Caveat: current installer can install to multiple agents. If one install targets multiple agents, a single lock entry may not be enough. For MVP either:

- create separate lock entries per agent by extending identity, or
- record `Agent: hermes` only when exactly one target agent is Hermes.

Preferred: extend lock identity and `AddEntry` semantics to allow same skill name for different agents.

**Step 5: Verify**

```bash
go test ./internal/installer ./internal/config ./internal/repository ./cmd
```

Manual install check with temp `HERMES_HOME`:

```bash
HERMES_HOME=$(mktemp -d) /tmp/ask-hermes-lifecycle skill install gitnexus-explorer --agent hermes --global --skip-score
```

Do not use real `~/.hermes` in tests.

**Step 6: Commit**

```bash
git add internal/installer internal/config internal/repository cmd
git commit -m "feat: record hermes skill install provenance"
```

---

### Task 7: Implement safe Hermes update behavior

**Objective:** Update ASK-managed Hermes skills and skip/refuse unsafe ones based on provenance and dirty state.

**Files:**
- Modify: `cmd/update.go`
- Create: `internal/hermes/update.go`
- Tests: `cmd/update_test.go`, `internal/hermes/update_test.go`

**Step 1: Write failing tests**

Cases:

1. ASK-managed Hermes skill with known source updates.
2. Imported local-only skill is skipped with `update unavailable`.
3. Bundled skill is skipped/refused.
4. Dirty checksum mismatch refuses update unless `--force`.
5. Named update only updates requested skill.

**Step 2: Add flags**

```bash
ask skill update [name] --agent hermes --global
ask skill update [name] --agent hermes --global --force
```

Do not change non-Hermes update behavior in this task except where necessary.

**Step 3: Implement update strategy switch**

Conceptual behavior:

```go
switch entry.UpdateStrategy {
case "hermes-index", "git":
    // reinstall/update from recorded URL/source identifier into ASK cache, then relink
case "none", "":
    skip
}
```

For MVP, if reinstalling is easier and safe:

- resolve source to temp,
- checksum current installed tree,
- if clean, replace ASK cache atomically,
- refresh symlink/copy target,
- update checksum and commit.

**Step 4: Verify**

```bash
go test ./cmd ./internal/hermes ./internal/installer
```

Manual dry update once implemented:

```bash
/tmp/ask-hermes-lifecycle skill update --agent hermes --global
```

**Step 5: Commit**

```bash
git add cmd/update.go internal/hermes internal/installer internal/config
git commit -m "feat: update managed hermes skills safely"
```

---

### Task 8: Implement safe Hermes uninstall behavior

**Objective:** Make `ask skill uninstall ... --agent hermes` ownership-aware and non-destructive by default for imported native skills.

**Files:**
- Modify: `cmd/uninstall.go`
- Create: `internal/hermes/uninstall.go`
- Tests: `cmd/uninstall_test.go`, `internal/hermes/uninstall_test.go`

**Step 1: Write failing tests**

Cases:

1. ASK-owned symlinked Hermes install removes target symlink, ASK source, config entry, lock entry.
2. Imported in-place skill refuses deletion without `--forget` or `--delete-files`.
3. `--forget` removes only lock/config tracking.
4. `--delete-files` removes in-place Hermes directory after explicit flag.
5. Bundled/core skills cannot be uninstalled by ASK.

**Step 2: Add flags**

```bash
ask skill uninstall foo --agent hermes --global --forget
ask skill uninstall foo --agent hermes --global --delete-files
```

Clarify existing `--all` behavior for non-Hermes. For Hermes, either map `--all` to ASK-owned full uninstall only or reject with a clearer message if ambiguous.

**Step 3: Implement Hermes-specific path**

When `--agent hermes` is present, route through Hermes uninstall logic instead of generic directory removal.

Rules:

- `Ownership == ask`: remove target and ASK source.
- `Ownership == imported && InstallMode == in-place`: refuse unless `--forget` or `--delete-files`.
- `Ownership == bundled`: always refuse.
- Missing lock entry but directory exists: treat as Hermes-native unmanaged; refuse destructive deletion and suggest import or manual removal.

**Step 4: Verify**

```bash
go test ./cmd ./internal/hermes ./internal/config
```

**Step 5: Commit**

```bash
git add cmd/uninstall.go internal/hermes internal/config
git commit -m "feat: uninstall hermes skills with ownership checks"
```

---

### Task 9: Remove duplicate/recommended bundled source from local/global ASK config/docs

**Objective:** Align documented and local recommended config with policy: use only `hermes-index` unless direct optional source is explicitly retained.

**Files:**
- Modify docs/README/config examples as applicable.
- Possibly modify default/recommended repo setup code if ASK has one.
- Do not mutate user config in code without explicit user action.

**Step 1: Search for Hermes bundled docs/config references**

```bash
search_files("hermes-bundled|hermes-agent/skills|optional-skills|hermes-index", path="/home/openclaw/dev/ask")
```

**Step 2: Write tests if there is config generation logic**

If repo defaults are generated, add tests asserting bundled source is absent and `hermes-index` is present.

**Step 3: Update docs/help text**

Mention:

```text
ASK does not manage Hermes bundled skills. Use hermes-index for user-installable Hermes skills.
```

**Step 4: Verify**

```bash
go test ./...
```

**Step 5: Commit**

```bash
git add README.md docs cmd internal
git commit -m "docs: clarify hermes skill source policy"
```

---

### Task 10: End-to-end verification and final cleanup

**Objective:** Prove the full lifecycle works without touching real Hermes data except in an explicitly controlled manual check.

**Files:**
- Tests only unless fixes are needed.

**Step 1: Run full test suite**

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./...
```

Expected: all pass.

**Step 2: Build test binary**

```bash
go build -o /tmp/ask-hermes-lifecycle .
```

**Step 3: Run temp-HERMES_HOME smoke test**

```bash
TMP_HOME=$(mktemp -d)
TMP_ASK=$(mktemp -d)
HERMES_HOME="$TMP_HOME/hermes" HOME="$TMP_ASK" /tmp/ask-hermes-lifecycle skill list --agent hermes --global
```

Then test import/list/uninstall with a fake local skill:

```bash
mkdir -p "$TMP_HOME/hermes/skills/local-debug-skill"
printf '%s\n' '---' 'name: local-debug-skill' 'description: Local debug skill' '---' > "$TMP_HOME/hermes/skills/local-debug-skill/SKILL.md"
HERMES_HOME="$TMP_HOME/hermes" HOME="$TMP_ASK" /tmp/ask-hermes-lifecycle skill import --agent hermes --global --dry-run
HERMES_HOME="$TMP_HOME/hermes" HOME="$TMP_ASK" /tmp/ask-hermes-lifecycle skill import --agent hermes --global --all
HERMES_HOME="$TMP_HOME/hermes" HOME="$TMP_ASK" /tmp/ask-hermes-lifecycle skill list --agent hermes --global
HERMES_HOME="$TMP_HOME/hermes" HOME="$TMP_ASK" /tmp/ask-hermes-lifecycle skill uninstall local-debug-skill --agent hermes --global --forget
```

Expected:

- dry-run writes nothing,
- import writes lock metadata,
- list shows imported/local-only,
- uninstall without `--forget` refuses destructive deletion,
- `--forget` removes tracking but leaves files.

**Step 4: Optional real config check**

Only after explicit user approval, inspect real config and recommend removing `hermes-bundled` if present. Do not mutate real `~/.ask/config.yaml` silently.

Read-only commands:

```bash
ask -g repo list
ask -g search nexus
```

Expected: `gitnexus-explorer` remains discoverable through `hermes-index`.

**Step 5: Final commit if cleanup occurred**

```bash
git status --short
git diff --check
git add <changed-files>
git commit -m "test: cover hermes skill lifecycle"
```

---

## Risks and Tradeoffs

### Risk: lockfile identity migration

Existing `LockFile.AddEntry` keys only on name. Adding agent-aware entries can change behavior when the same skill name is installed for multiple agents.

Mitigation:

- Keep old methods for compatibility.
- Add new agent-aware methods.
- Only use new identity semantics in Hermes-specific paths initially.
- Add regression tests for old lock behavior.

### Risk: optional-skills source duplication

Keeping both `hermes-index` and direct `optional-skills` can produce duplicate search results.

Mitigation:

- Recommend only `hermes-index`.
- Later deduplicate search results by canonical source identifier / resolved GitHub path.

### Risk: provenance overclaiming

ASK may be tempted to match local skill name to index result and mark it updateable.

Mitigation:

- Import should default to `update_strategy: none` unless provenance is explicit and strong.
- Do not same-name-match by default.

### Risk: destructive uninstall

Existing uninstall has broad remove behavior. Hermes-native skills need stricter handling.

Mitigation:

- Route Hermes uninstall through ownership-aware logic.
- Default to non-destructive refusal for unmanaged/imported in-place content.
- Require explicit `--delete-files` for file deletion.

### Risk: bundled skills physically appear in `$HERMES_HOME/skills`

If Hermes stores built-ins in the same directory as user skills, scanner classification may be hard.

Mitigation:

- Prefer source/lock metadata when available.
- For unclassified local directories, treat as local-only user skill, not bundled.
- Do not manage known Hermes repo bundled source paths.
- Add `--include-bundled` later only if bundled detection is reliable.

---

## Open Questions Before Implementation

1. Should direct `NousResearch/hermes-agent/optional-skills` remain supported as a `dir` repo, or should we remove it from recommended config and rely exclusively on `hermes-index`?

   Recommendation: rely exclusively on `hermes-index` for official Hermes skills; keep direct optional repo support only as compatibility.

2. Should `ask skill list --agent hermes` show unmanaged local-only skills by default?

   Recommendation: yes, because Hermes sees them and the user asked for natively installed skill management. Mark them as `managed_by: hermes` / `update: unavailable`.

3. Should import default to `--all` behavior or require explicit skill names / `--all`?

   Recommendation: require explicit name or `--all`; default with no args should show dry-run-style summary and instructions.

4. Should `--delete-files` require an interactive confirmation?

   Recommendation: yes when TTY is interactive; require `--yes` for noninteractive destructive deletion.

5. Should imported in-place skills ever become updateable without migration?

   Recommendation: not in MVP. Add `ask skill import --agent hermes --migrate` later to move/copy into ASK-owned storage and enable full lifecycle.

---

## Acceptance Criteria

- `ask skill list --agent hermes --global` distinguishes ASK-managed, imported/local-only, and unmanaged Hermes-visible skills.
- ASK does not search/install/update/uninstall `NousResearch/hermes-agent/skills` bundled skills.
- `ask skill import --agent hermes --global --dry-run` safely reports what would be imported.
- `ask skill import --agent hermes --global --all` records eligible user-installed skills with local-only provenance and no update strategy.
- `ask skill install ... --agent hermes --global` records agent/provenance/ownership/checksum metadata.
- `ask skill update ... --agent hermes --global` updates only safe ASK-managed skills and skips/refuses unsafe cases clearly.
- `ask skill uninstall ... --agent hermes --global` is non-destructive by default for imported/native skills.
- Full `go test ./...` passes.
- Manual smoke tests use temp `HERMES_HOME` before any real-user-path test.

---

## Suggested Implementation Order

```text
1. Source classification / bundled policy
2. Bundled source rejection in repo/search/install paths
3. Hermes installed-skill scanner
4. Hermes-aware list output
5. Import/adopt command
6. Install provenance metadata
7. Safe update
8. Safe uninstall
9. Docs/config recommendation cleanup
10. End-to-end verification
```

This order keeps the riskiest policy boundary first, then builds read-only discovery, then import/provenance, then lifecycle mutations.

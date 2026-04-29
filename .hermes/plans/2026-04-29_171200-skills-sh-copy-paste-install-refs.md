# skills.sh Copy-Paste Install Refs Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Make supported `skills.sh` search results actionable without an API key by exposing a copy-pasteable native ASK install reference and, where safe, allowing install by selecting from public search results.

**Architecture:** Keep `skills.sh` as discovery/metadata only. Public search continues to use unauthenticated `/api/search`; ASK derives native GitHub install refs by resolving the result's `source` + `skillId/name` to an actual skill directory in the GitHub repository. Search output should show that native ref. Install-by-name should use public search only when explicitly scoped to `skills.sh`, and should reject ambiguous or unsupported results rather than guessing.

**Tech Stack:** Go, cobra CLI, existing `internal/repository` source abstraction, GitHub API/tree discovery or existing git-based discovery helpers, TDD with package-level command/repository tests.

---

## Current Problem

A user can run:

```bash
ask skill search grill --remote --global
```

and see `skills.sh` results such as `grill-me`, but the CLI output does not include enough information to install the result. `ask install -g grill-me` fails because `grill-me` is only a display name and there may be duplicates. `ask install -g --repo skills-sh grill-me` currently uses full catalog fetch/list, which requires an API key.

The UX violates the catalog integration rule: users must be able to go from `search` to `install` without guessing hidden source paths.

## Non-Goals

- Do not call `npx skills add`.
- Do not use `skills.sh` blob/download endpoints as a general install source.
- Do not fake list-all through empty/wildcard public search.
- Do not silently pick one result when multiple supported matches have the same display name.
- Do not implement `.well-known/agent-skills` native artifact install in this slice.

## Desired UX

Search text output should include a copy-pasteable native install ref for supported results, for example:

```text
NAME       SOURCE     INSTALL REF                                      STARS  DESCRIPTION
grill-me   skills.sh  mattpocock/skills/skills/productivity/grill-me   49729  
```

JSON output should include the same field, for example:

```json
{
  "name": "grill-me",
  "source": "skills.sh",
  "install_ref": "mattpocock/skills/skills/productivity/grill-me",
  "supported": true
}
```

Optional install-by-public-search behavior should work only when explicitly scoped:

```bash
ask install -g --repo skills-sh grill-me
```

Expected semantics:
- if exactly one supported public search match has name `grill-me`, install its native ref;
- if multiple supported matches have name `grill-me`, fail with a list of disambiguating install refs;
- if only unsupported matches exist, fail with unsupported reasons;
- if no matches exist, report not found;
- no API key required for this explicit search/select path.

---

## Task 1: Preserve a Native Install Ref in Search Results

**Objective:** Ensure repository candidates converted to `github.Repository` retain a user-visible native install ref distinct from display name and source label.

**Files:**
- Modify: `internal/github/github.go` or existing repository DTO file if `Repository` lives elsewhere
- Modify: `internal/repository/source_types.go`
- Test: `internal/repository/source_types_task4_test.go` or new focused test

**Step 1: Write failing tests**

Add tests proving that a `SkillCandidate` with `Install.Value = "https://github.com/owner/repo/tree/main/path"` converts to `github.Repository` with a field that can be displayed as an install ref.

The test should assert both:
- URL/path is still used by installer internals;
- a user-facing install ref is preserved or derived.

**Step 2: Run targeted test and verify RED**

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository -run 'InstallRef|SkillsSH' -count=1
```

Expected: fail because no displayable install ref field exists yet, or JSON/search display does not expose it.

**Step 3: Implement minimal DTO support**

Add a field such as:

```go
InstallRef string
```

to the internal/public repository result shape, populated from `candidate.Install.Value`.

Keep backward compatibility: do not remove or repurpose `HTMLURL`.

**Step 4: Run targeted tests and verify GREEN**

```bash
go test ./internal/repository -run 'InstallRef|SkillsSH' -count=1
```

---

## Task 2: Resolve Public skills.sh Search Results to Actual GitHub Skill Paths

**Objective:** For public `/api/search` results, derive the exact native GitHub path when possible, instead of only resolving repo root.

**Files:**
- Modify: `internal/repository/skills_sh.go`
- Test: `internal/repository/skills_sh_test.go`

**Step 1: Write failing tests**

Use fake HTTP/GitHub discovery seams rather than live network. Add tests for:

1. Public search result:

```json
{"source":"mattpocock/skills","skillId":"grill-me","name":"grill-me"}
```

with fake repo tree containing:

```text
skills/productivity/grill-me/SKILL.md
```

Expected install ref:

```text
https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me
```

2. Duplicate directory names in one repo should be unsupported/ambiguous unless exact disambiguation is possible.
3. Missing skill directory should remain visible but unsupported with a clear reason.
4. Domain-backed/non-GitHub entries remain unsupported.

**Step 2: Run targeted tests and verify RED**

```bash
go test ./internal/repository -run 'SkillsSH.*Path|SkillsSH.*Legacy|SkillsSH.*Public' -count=1
```

**Step 3: Implement minimal resolver**

Add a narrow resolver seam, for example:

```go
var resolveSkillsSHGitHubSkillPathFunc = resolveSkillsSHGitHubSkillPath
```

Implementation approach:
- trust bare `owner/repo` only in public skills.sh search path;
- fetch or reuse a repo tree/listing;
- find directories containing `SKILL.md` whose basename equals `skillId` or `name`;
- if exactly one match, return GitHub tree URL for that directory;
- if zero or multiple matches, return unsupported/ambiguous reason.

Prefer existing GitHub/tree helpers if available; otherwise use a small package-level resolver that can be faked in tests.

**Step 4: Run targeted tests and verify GREEN**

```bash
go test ./internal/repository -run SkillsSH -count=1
```

---

## Task 3: Display Install Ref in Search Output

**Objective:** Make search results copy-pasteable in both text and JSON output.

**Files:**
- Modify: `cmd/search.go`
- Test: `cmd/search_task4_test.go` or new `cmd/search_install_ref_test.go`

**Step 1: Write failing tests**

Add command/output tests proving that supported `skills.sh` search results include an install ref in:

- table/text output;
- JSON output.

JSON expected field:

```json
"install_ref": "mattpocock/skills/skills/productivity/grill-me"
```

or full GitHub URL if that is the existing installer-preferred format. Prefer whatever is copy-pasteable by `ask install`.

Unsupported results should either omit `install_ref` or set it to empty while preserving `unsupported_reason`.

**Step 2: Run targeted tests and verify RED**

```bash
go test ./cmd -run 'Search.*InstallRef|SkillsSH' -count=1
```

**Step 3: Implement output change**

Text output: add an `INSTALL REF` column, or print it in a second line only for results that have it if table width becomes ugly.

JSON output: add `install_ref,omitempty`.

**Step 4: Run targeted tests and verify GREEN**

```bash
go test ./cmd -run 'Search.*InstallRef|SkillsSH' -count=1
```

---

## Task 4: Explicit `--repo skills-sh <name>` Public Search Selection

**Objective:** Let users install from skills.sh by name without an API key when explicitly scoped to that repo, without requiring full catalog/list.

**Files:**
- Modify: `cmd/install.go`
- Test: `cmd/install_task4_test.go` or new `cmd/install_skills_sh_public_search_test.go`

**Step 1: Write failing tests**

Add command-level tests for `--repo skills-sh grill-me` using fake repository search/fetch seams:

1. Exactly one supported public search match by name queues its install ref.
2. Multiple supported matches with same name fail with disambiguation refs and do not install anything.
3. Unsupported-only matches fail with unsupported reason.
4. No matches fail with not found.
5. `--repo skills-sh` with no skill args still requires authenticated full list/fetch and should not call public search with an empty query.

**Step 2: Run targeted tests and verify RED**

```bash
go test ./cmd -run 'Repo.*SkillsSH|Install.*SkillsSH' -count=1
```

**Step 3: Implement selection path**

In the `--repo` install branch:
- if target repo type is `skills.sh` and skill args are provided, call `repository.SearchSkills(ctx, repo, wanted)` per wanted name rather than `FetchSkills`;
- filter exact `Name == wanted`;
- require exactly one supported result;
- use existing `appendInstallableRepoSkill` to queue install and preserve metadata;
- if ambiguous, print candidates with install refs.

Do not change `--repo skills-sh` with no args: it remains catalog/list and requires a token.

**Step 4: Run targeted tests and verify GREEN**

```bash
go test ./cmd -run 'Repo.*SkillsSH|Install.*SkillsSH' -count=1
```

---

## Task 5: Final Verification, Review, and Commit

**Objective:** Prove the feature works, update debrief memory, and commit.

**Files:**
- Modify: `memory/2026-04-29.md`

**Step 1: Run gates**

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
git diff --check
go test ./internal/repository ./internal/config ./cmd -count=1
go test ./...
```

**Step 2: Independent review**

Dispatch review focused on:
- no `npx` or skills.sh special installer;
- no API key required for explicit public search install path;
- no list-all faking;
- duplicate/ambiguous name handling;
- unsupported entries blocked before installer;
- install refs are copy-pasteable and tested.

**Step 3: Optional live smoke**

If network access is acceptable:

```bash
go run . skill search grill --remote --global
go run . install -g --repo skills-sh grill-me --skip-score
```

If there are multiple `grill-me` matches, expected behavior is a clean ambiguity error listing install refs. Then test one copy-pasteable ref from search output directly.

**Step 4: Update debrief**

Append a short debrief to `memory/2026-04-29.md`.

**Step 5: Commit**

```bash
git add internal/repository cmd memory/2026-04-29.md .hermes/plans/2026-04-29_171200-skills-sh-copy-paste-install-refs.md
git commit -m "feat: expose skills.sh install refs"
```

---

## Acceptance Criteria

- `ask skill search grill --remote --global` exposes copy-pasteable native install refs for supported `skills.sh` results.
- `ask skill search grill --remote --global --json` includes `install_ref` for supported results.
- `ask install -g --repo skills-sh <name>` can install an unambiguous supported public search result without a skills.sh API key.
- Ambiguous duplicate names fail with useful disambiguation instead of guessing.
- Unsupported/domain/artifact entries remain visible but cannot reach the installer.
- `--repo skills-sh` with no args does not use empty public search as list-all.
- Full test suite passes and independent review approves.

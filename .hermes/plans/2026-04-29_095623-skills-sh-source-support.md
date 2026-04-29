# skills.sh Source Support Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Add `skills.sh` as an explicit ASK repository source for discovery/metadata while always resolving selected skills to a native ASK-installable source; do not invoke the skills.sh/npx installer.

**Architecture:** Implement `skills.sh` as an opt-in network-backed discovery source (`type: skills.sh`) using the existing `repositorySource`/`SkillCandidate` dispatcher. Add a resolver layer that maps catalog entries to native ASK install refs: GitHub tree/repo refs first, then public git hosts and well-known `skill-md`/`archive` artifacts only when ASK has native support for those ref kinds. Preserve skills.sh provenance for update/uninstall metadata, but keep installation mechanics inside ASK.

**Tech Stack:** Go, Cobra, internal repository source dispatcher, `net/http` with context/timeouts, table-driven tests with `httptest`, existing ASK lock/provenance metadata.

---

## API Findings

Docs: https://skills.sh/docs/api

Base URL: `https://skills.sh`, API namespace `/api/v1`, JSON responses.

Endpoints:

- `GET /api/v1/skills?view=all-time|trending|hot&page=0&per_page=100`
  - returns `{ data: V1Skill[], pagination: { page, perPage, total, hasMore } }`
  - `per_page` range: `1..500`
- `GET /api/v1/skills/search?q=<query>&limit=<n>`
  - returns `{ data: V1Skill[], query, searchType, count, durationMs }`
  - `limit` range: `1..200`
- `GET /api/v1/skills/curated`
  - returns official/curated owners and nested skills
- `GET /api/v1/skills/{source}/{skill}`
  - returns detail including `{ id, source, slug, installs, hash, files[] }`
- `GET /api/v1/skills/audit/{source}/{skill}`
  - returns security audit data; `404` means no audit yet, not necessarily invalid skill

`V1Skill` fields from docs:

```json
{
  "id": "vercel-labs/agent-skills/next-js-development",
  "slug": "next-js-development",
  "name": "Next.js Development",
  "source": "vercel-labs/agent-skills",
  "installs": 24531,
  "sourceType": "github",
  "installUrl": "https://github.com/vercel-labs/agent-skills",
  "url": "https://skills.sh/vercel-labs/agent-skills/next-js-development",
  "isDuplicate": true
}
```

Important live-doc mismatch observed on 2026-04-29: docs say unauthenticated access works with lower rate limits, but live unauthenticated `/api/v1/skills*` list/search/curated/detail returned `401 authentication_required`; audit returned unauthenticated `404` for an unaudited example. Implementation must support an API token and produce a clear `SKILLS_SH_API_KEY` setup error on `401`.

Also observed: the skills CLI currently uses a public legacy-ish endpoint `GET /api/search?q=<query>&limit=<n>`, which returned unauthenticated compact search results shaped as `{ query, searchType, skills: [{ id, skillId, name, installs, source }], count, duration_ms }`. This endpoint is useful evidence for discovery behavior, but ASK should prefer documented `/api/v1` when an API key is configured unless we deliberately choose a fallback mode.

Non-GitHub/well-known examples found via live search:

- `mintlify.com/mintlify`, `mintlify.com/mintlify-docs`, `mintlify.com/mintlify-api` have `source: "mintlify.com"` and no GitHub owner/repo shape. Direct skills.sh detail/download probes for these returned 404/invalid-page responses, so do not assume the skills.sh site itself is an artifact source for domain-backed skills.
- The Cloudflare Agent Skills Discovery RFC defines `.well-known/agent-skills/index.json` v0.2.0 entries with `type: "skill-md"` or `type: "archive"`, a relative/absolute `url`, and `digest`. Live examples include `https://docs.lovable.dev/.well-known/agent-skills/index.json` and `https://modelcontextprotocol.io/.well-known/agent-skills/index.json`, both `skill-md`-only at the time checked.
- The skills CLI also has an older well-known provider implementation expecting `index.json` entries with `{ name, description, files[] }` and fetches `{base}/.well-known/agent-skills/{skill}/SKILL.md` plus listed files. ASK should support the current RFC v0.2.0 shape first; the legacy `files[]` shape can be a compatibility follow-up if needed.

Error shape:

```json
{ "error": "error_code", "message": "Human-readable description." }
```

Rate-limit docs mention `X-RateLimit-*` and `Retry-After`, but live 401/audit responses did not include them. Code must not assume those headers exist.

---

## Design Decisions

1. **Source type/name**
   - Add `config.RepoTypeSkillsSH = "skills.sh"`.
   - Recommended configured repo:
     ```yaml
     repos:
       - name: skills-sh
         type: skills.sh
         url: https://skills.sh
         token: ${SKILLS_SH_API_KEY} # or store literal token only if existing ASK config supports that pattern
     ```
   - Prefer an environment-variable token path in docs/config guidance to avoid writing secrets into config. Existing `Repo.Token` can remain supported for private/local setups.

2. **Opt-in first, not default**
   - Do **not** add skills.sh to `DefaultConfig()` initially because live API currently requires an API key despite docs saying auth is optional.
   - Users should add/configure it explicitly.

3. **Install semantics: discovery-only + native ASK resolution**
   - Treat skills.sh as a catalog, not an installer.
   - Never shell out to `npx skills add`, never rely on skills.sh telemetry/install side effects, and never install by copying opaque skills.sh-generated output unless ASK owns and validates that installer path.
   - Add a resolver function that turns each catalog entry into one of ASK's native install refs. Suggested shape:
     ```go
     type ResolvedSkillInstall struct {
         Kind   InstallRefKind
         Value  string
         Digest string // optional, for direct artifact/well-known integrity
     }
     ```
   - First implementation slice should support only currently native-safe refs:
     - `sourceType == "github"` with `installUrl`/`source` resolving to `github.com/<owner>/<repo>`.
     - Prefer a GitHub tree/subpath URL when details expose an exact path. If only repo root + skill slug are available, resolve to repo root plus existing ASK filtering/discovery rather than guessing `/<slug>` for every repository layout.
   - Next native slices, if desired:
     - `gitlab`/generic public git URLs, if ASK's parser/installer can already clone them safely.
     - well-known `skill-md`: fetch the Markdown, verify `sha256` when `digest` is present, and install as a one-file skill through an ASK-owned direct-file installer.
     - well-known `archive`: fetch archive, verify `sha256`, safely extract with path traversal and symlink checks, then install through an ASK-owned archive installer.
   - The skills.sh detail endpoint `files[]` can inform exact path/digest later, but should not become the installation source unless ASK explicitly implements a validated snapshot installer.

4. **Unsupported entries**
   - Do not drop non-GitHub entries silently from discovery. Show/list them as discovered but `unsupported` or `not natively installable yet` when they cannot resolve to a native ASK install ref.
   - Do not coerce external URLs, domains, or bare IDs into GitHub paths.
   - For domain sources such as `mintlify.com`, try well-known discovery on the domain only if/when ASK implements well-known artifact support; otherwise report that the entry needs `skill-md`/`archive` support rather than pretending it is GitHub-backed.

5. **Provenance**
   - Populate:
     - `Source: config.RepoTypeSkillsSH`
     - `SourceIdentifier: skill.ID` (stable `source/slug`)
     - `UpdateStrategy: "skills.sh"`
   - Use `installs` as `StargazersCount` only as a pragmatic ranking column until ASK has a neutral popularity field; label docs/tests accordingly to avoid implying GitHub stars.

6. **HTTP behavior**
   - Use a package-level HTTP client/seam with timeout, context support, `User-Agent: ask-cli`.
   - Limit response bodies, e.g. 5 MiB, matching registry/repo patterns.
   - On `429`, include `Retry-After` when present.
   - On `401`, return a setup-focused error: `skills.sh API key required; set SKILLS_SH_API_KEY or repo token`.
   - On `503`, surface retry/backoff guidance; no hidden unbounded retries in CLI source code.

7. **Authentication**
   - Resolve token from `repo.Token` first only if present; otherwise from `SKILLS_SH_API_KEY`.
   - Never print token values in errors, debug logs, tests, or debriefs.

8. **Audits**
   - Do not block installation based on audit in initial source support unless user explicitly asks for policy enforcement.
   - Add source parser structs that can support audit later, but keep first source wiring to list/search/fetch.
   - Treat audit `riskLevel` as open-ended string: live docs/reports may include values beyond the documented enum, e.g. `SAFE`.

---

## Task 1: Add source type constant and dispatcher seam

**Objective:** Register an opt-in `skills.sh` repository source without implementing network behavior yet.

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/repository/source.go`
- Test: `internal/repository/source_test.go`

**Step 1: Write failing tests**

Add tests proving `sourceForRepo(config.Repo{Type: config.RepoTypeSkillsSH})` is recognized and that `FetchSkills`/`SearchSkills` dispatch into replaceable skills.sh source seams.

Expected RED: `config.RepoTypeSkillsSH` and source seam do not exist.

**Step 2: Implement minimal code**

Add:

```go
const RepoTypeSkillsSH = "skills.sh"
```

Add package-level seams similar to existing source functions:

```go
var (
    searchSkillsSHFunc = searchSkillsSHSource
    fetchSkillsSHFunc  = fetchSkillsSHSource
)
```

Register in `sourceForRepo`.

**Step 3: Verify**

```bash
go test ./internal/repository -run 'Test.*SkillsSH|Test.*SourceDispatcher' -count=1
```

---

## Task 2: Implement skills.sh response parsing and native install resolution

**Objective:** Parse documented skills.sh payloads, preserve all discovered entries for UX, and mark whether each entry resolves to a native ASK install ref.

**Files:**
- Create: `internal/repository/skills_sh.go`
- Create: `internal/repository/skills_sh_test.go`
- Modify: `internal/repository/source_types.go` if a candidate needs an unsupported/installability flag

**Step 1: Write failing parser/resolver tests**

Cover:

- search/list payload maps `sourceType: github` to an installable candidate,
- GitHub host validation uses `net/url` and rejects lookalikes or non-HTTPS surprises,
- repo-root GitHub entries do not blindly append `slug` as a path unless exact path evidence is present,
- public legacy `/api/search` payload can be parsed into catalog records but has reduced metadata,
- non-GitHub/domain entries such as `source: "mintlify.com"` remain visible but are marked unsupported when no native resolver exists,
- well-known v0.2.0 records with `type: "skill-md"` or `type: "archive"` are classified as future native artifact refs, not GitHub refs,
- malformed JSON returns error,
- duplicate entries are either skipped or marked/de-prioritized; for first slice, skip `isDuplicate: true` by default unless user searches exact ID,
- candidate fields preserve name, description/page URL if available, installs-as-ranking, source identifier, update strategy.

Expected RED: parser and resolver missing.

**Step 2: Implement minimal parser/resolver**

Define structs for list/search payloads, legacy search payloads, and `skillsSHSkill` with open-ended string fields. Add strict URL normalization using `net/url`; do not accept prefix-only checks. Keep unsupported entries distinguishable from absent entries so list/search can explain why a result cannot be installed yet.

**Step 3: Verify**

```bash
go test ./internal/repository -run SkillsSH -count=1
```

---

## Task 3: Implement HTTP client for list/search/fetch

**Objective:** Call skills.sh APIs with auth, context, body limits, pagination, and clear errors.

**Files:**
- Modify: `internal/repository/skills_sh.go`
- Test: `internal/repository/skills_sh_test.go`

**Step 1: Write failing HTTP tests with `httptest.Server`**

Cover:

- `SearchSkills(ctx, repo, "react")` requests `/api/v1/skills/search?q=react&limit=...`,
- `FetchSkills(repo)` requests `/api/v1/skills?view=all-time&page=0&per_page=...`,
- `Authorization: Bearer <token>` is set from repo token or `SKILLS_SH_API_KEY`,
- `401` returns setup-focused API-key error without leaking token,
- `429` includes `Retry-After` if present,
- malformed JSON and non-2xx return useful errors,
- body limit is enforced.

Expected RED: HTTP source missing.

**Step 2: Implement minimal HTTP behavior**

Use repo URL as base URL, defaulting to `https://skills.sh` when empty. Use context-aware requests for search. For fetch, either use `context.Background()` or introduce a helper accepting context internally without public API churn.

**Step 3: Verify**

```bash
go test ./internal/repository -run SkillsSH -count=1
```

---

## Task 4: Wire install provenance and unsupported-result UX

**Objective:** Ensure `ask repo list skills-sh`, `ask skill search --repo skills-sh`, and `ask skill install --repo skills-sh <skill>` preserve skills.sh metadata while installing only through native ASK refs.

**Files:**
- Modify: `internal/repository/source_types.go` if needed
- Modify: `cmd/install_test.go` or add focused command tests if source metadata is dropped
- Test: existing repository/cmd tests

**Step 1: Write failing adapter/CLI tests**

Cover that candidates converted to `github.Repository` retain:

- `Source == config.RepoTypeSkillsSH`,
- `SourceIdentifier == skill.ID`,
- `UpdateStrategy == "skills.sh"`,
- install value suitable for existing installer when the result is natively resolvable,
- unsupported/domain/well-known entries display a clear unsupported reason and are not passed to the installer.

If install command tests use repo fetch seams, assert `recordInstallSourceMetadata` records `skills.sh` provenance for supported installs.

**Step 2: Implement minimal adapter/CLI changes**

Likely no adapter change is needed if `SkillCandidate.Install.Value` is already assigned to `HTMLURL`. Add only what's necessary.

**Step 3: Verify**

```bash
go test ./internal/repository ./cmd -run 'SkillsSH|Install.*SourceMetadata' -count=1
```

---

## Task 5: Add repo-add/config UX for skills.sh

**Objective:** Make skills.sh easy to configure while avoiding accidental default enablement or token leaks.

**Files:**
- Modify: `cmd/repo.go`
- Modify: docs/config/help files as appropriate
- Test: repo command tests

**Step 1: Write failing CLI/config tests**

Choose one UX path:

Option A — explicit command support:

```bash
ask repo add skills-sh --type skills.sh --url https://skills.sh --token-env SKILLS_SH_API_KEY
```

Option B — document manual config only for first implementation.

If implementing Option A, tests should prove:

- `repo add` can create a `type: skills.sh` source without GitHub repo validation,
- token values are not printed,
- duplicate names are handled,
- `repo list` displays `https://skills.sh` cleanly.

**Step 2: Implement minimal UX**

Prefer a small `--type` switch rather than overloading GitHub validation. If this is too invasive, defer to docs and skip code in this task.

**Step 3: Verify**

```bash
go test ./cmd -run 'Repo.*SkillsSH|RepoAdd' -count=1
```

---

## Task 6: Documentation and manual smoke test

**Objective:** Document safe setup/usage and verify the explicit source against an isolated config.

**Files:**
- Modify: README/docs/config as appropriate
- Optional: add sample config snippet

**Step 1: Write docs**

Include:

```yaml
repos:
  - name: skills-sh
    type: skills.sh
    url: https://skills.sh
```

and:

```bash
export SKILLS_SH_API_KEY=...
ask repo list skills-sh --global
ask skill search react --repo skills-sh --global
ask skill install <skill-name> --repo skills-sh --global --skip-score
```

Mention that live API may require an API key, even though public docs describe unauthenticated limits.

**Step 2: Run final verification**

```bash
git diff --check
go test ./internal/repository ./cmd ./internal/config -count=1
go test ./...
```

**Step 3: Manual smoke with temp home**

Use isolated `HOME`/`ASK_CONFIG`/`HERMES_HOME` as applicable. If no API key is available, smoke the expected `401` setup error and do not treat it as a failure; if `SKILLS_SH_API_KEY` is available, verify list/search and one install path.

---

## Task 7: Native well-known/artifact installer follow-up (optional but likely valuable)

**Objective:** Support non-GitHub skills.sh discoveries without using the skills.sh installer by adding ASK-owned direct artifact install refs.

**Files:**
- Modify: `internal/repository/source_types.go`
- Create/modify: installer package files for direct `skill-md` and archive install support
- Test: repository resolver tests and installer extraction tests

**Step 1: Write failing tests**

Cover:

- `.well-known/agent-skills/index.json` v0.2.0 with `type: "skill-md"` resolves to a direct file ref,
- relative artifact URLs resolve against the well-known index URL per RFC 3986,
- `digest: sha256:<hex>` is verified before install,
- `archive` extraction rejects `../`, absolute paths, unsafe symlinks, and missing `SKILL.md`,
- `skill-md` installs as a valid skill folder containing `SKILL.md`,
- legacy well-known `files[]` shape is either rejected with a clear message or supported in a separate compatibility test.

**Step 2: Implement minimal native artifact support**

Add new `InstallRefKind` values only when the installer can consume them safely. Keep network fetches bounded by context, size limits, content-type/extension sanity checks, and digest verification.

**Step 3: Verify**

```bash
go test ./internal/repository ./internal/installer ./cmd -run 'WellKnown|Artifact|SkillsSH' -count=1
go test ./...
```

---

## Open Questions Before Implementation

1. Which native install ref kinds should land in the first PR? I recommend **GitHub/public-git only first**, then a separate well-known artifact PR with digest/extraction safety tests.
2. Should `skills.sh` be a default repo once auth behavior is clarified? I recommend **not default** until unauthenticated access works or ASK has onboarding for the API key.
3. How should duplicates be handled? I recommend skipping `isDuplicate` by default and exposing them later behind an explicit flag if needed.
4. Should audit status gate install? I recommend surfacing audit data in a later task; do not block initial source support on it.
5. Should ASK use public `/api/search` fallback when `/api/v1` requires auth? I recommend treating it as an optional fallback for search-only discovery, not as a full source, because it lacks `sourceType`, `installUrl`, detail metadata, and documented stability.

---

## Final Review Checklist

- [ ] No tokens printed or committed.
- [ ] No live network calls in tests.
- [ ] `skills.sh` is opt-in, not default.
- [ ] Non-GitHub/well-known entries are never coerced into GitHub paths; they are either resolved to a native ASK artifact ref or displayed with a clear unsupported reason.
- [ ] Source provenance survives search/fetch/install.
- [ ] Installation path never shells out to `npx skills add` or another external installer.
- [ ] Direct artifact support, if added, verifies digest and prevents archive path traversal/symlink escapes.
- [ ] Full suite passes.

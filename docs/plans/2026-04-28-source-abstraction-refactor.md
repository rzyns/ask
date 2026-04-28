# Source Abstraction Refactor Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Refactor ASK's existing skill-source handling behind a small internal abstraction before adding Hermes skill sources.

**Architecture:** Keep the persisted `config.Repo` model and existing repo type strings unchanged. First centralize remote search dispatch in `internal/repository`, then introduce typed source/result models and resolver/fetch seams in later slices. Preserve behavior at every step; Hermes integration is intentionally out of scope for this plan.

**Tech Stack:** Go, Cobra CLI, existing `internal/config`, `internal/repository`, `internal/github`, `internal/skillhub`, and Go unit tests.

---

## Guardrails

- Use strict TDD for production-code changes: write a failing test, watch it fail, implement, watch it pass.
- Preserve existing CLI behavior in every slice.
- Do not add Hermes source integration in this refactor branch/slice.
- Do not rename persisted repo type strings: `topic`, `dir`, `registry`, `skillhub`.
- Do not refactor installer side effects until characterization tests exist.
- Go commands require:
  ```bash
  export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
  ```

---

## Target End-State

ASK should have one internal source layer that existing code can call instead of manually switching on `config.Repo.Type` in multiple packages.

Initial shape:

```go
package repository

func SearchSkills(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error)
```

Later shape after behavior is characterized:

```go
type Source interface {
    Search(ctx context.Context, keyword string) ([]SkillResult, error)
    Resolve(ctx context.Context, ref string) (*InstallRef, error)
    Fetch(ctx context.Context, ref InstallRef) (*SkillBundle, error)
}
```

Do not jump to the later shape until existing search/fetch/install behavior is covered well enough.

---

### Task 1: Centralize remote search dispatch in `internal/repository`

**Objective:** Move the existing `cmd/search.go` remote-source switch into `repository.SearchSkills` with no behavior change.

**Files:**
- Create: `internal/repository/source.go`
- Create: `internal/repository/source_test.go`
- Modify: `cmd/search.go`

**Step 1: Write failing tests**

Create `internal/repository/source_test.go` with tests for the new dispatcher:

```go
package repository

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/yeasy/ask/internal/config"
)

func TestSearchSkillsUnknownTypeReturnsError(t *testing.T) {
    _, err := SearchSkills(context.Background(), config.Repo{Type: "bogus"}, "")
    if err == nil {
        t.Fatal("expected error for unknown repository type")
    }
    if got := err.Error(); !contains(got, "unknown repository type: bogus") {
        t.Fatalf("expected unknown type error, got %q", got)
    }
}

func TestSearchSkillsRegistryUsesRegistryFetcher(t *testing.T) {
    config.SetOffline(false)

    index := validRegistryIndex()
    data, err := json.Marshal(index)
    if err != nil {
        t.Fatalf("failed to marshal test index: %v", err)
    }

    _, cleanup := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(data)
    })
    defer cleanup()

    results, err := SearchSkills(context.Background(), config.Repo{
        Type: "registry",
        URL:  "owner/repo/registry/index.json",
    }, "docker")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0].Name != "docker-helper" {
        t.Fatalf("expected docker-helper, got %q", results[0].Name)
    }
}

func TestSearchSkillsDirInvalidURLPreservesNoopBehavior(t *testing.T) {
    results, err := SearchSkills(context.Background(), config.Repo{
        Type: "dir",
        URL:  "owneronly",
    }, "anything")
    if err != nil {
        t.Fatalf("expected nil error to preserve existing search behavior, got %v", err)
    }
    if len(results) != 0 {
        t.Fatalf("expected no results, got %d", len(results))
    }
}
```

Adjust imports as needed; include `net/http` because the second test uses `http.ResponseWriter`.

**Step 2: Run test to verify failure**

Run:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository -run 'TestSearchSkills' -count=1
```

Expected: FAIL because `SearchSkills` does not exist yet.

**Step 3: Implement minimal dispatcher**

Create `internal/repository/source.go`:

```go
package repository

import (
    "context"
    "fmt"
    "strings"

    "github.com/yeasy/ask/internal/config"
    "github.com/yeasy/ask/internal/github"
)

func SearchSkills(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    switch repo.Type {
    case "topic":
        return github.SearchTopic(repo.URL, keyword)
    case "dir":
        parts := strings.Split(repo.URL, "/")
        if len(parts) < 2 {
            return nil, nil
        }
        owner := parts[0]
        repoName := parts[1]
        path := ""
        if len(parts) > 2 {
            path = strings.Join(parts[2:], "/")
        }
        repos, err := github.SearchDir(owner, repoName, path)
        if err != nil || keyword == "" {
            return repos, err
        }
        var filtered []github.Repository
        lowerKeyword := strings.ToLower(keyword)
        for _, rp := range repos {
            if strings.Contains(strings.ToLower(rp.Name), lowerKeyword) {
                filtered = append(filtered, rp)
            }
        }
        return filtered, nil
    case "registry":
        return FetchSkillsFromRegistry(repo.URL, keyword)
    case "skillhub":
        return FetchSkillsFromSkillHub(keyword, "")
    default:
        return nil, fmt.Errorf("unknown repository type: %s", repo.Type)
    }
}
```

**Step 4: Run test to verify pass**

Run:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository -run 'TestSearchSkills' -count=1
```

Expected: PASS.

**Step 5: Refactor `cmd/search.go`**

Replace the remote-search switch in `cmd/search.go` with:

```go
repos, err = repository.SearchSkills(searchCtx, r, keyword)
```

Remove now-unused imports if any. `strings` is still needed by `runSearch`, so it should remain.

**Step 6: Verify command package and full suite**

Run:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository ./cmd -count=1
go test ./... 
```

Expected: all tests pass.

**Step 7: Commit**

```bash
git add internal/repository/source.go internal/repository/source_test.go cmd/search.go docs/plans/2026-04-28-source-abstraction-refactor.md
git commit -m "refactor: centralize source search dispatch"
```

---

### Task 2: Add repo type constants for persisted source types

**Objective:** Replace repeated raw repo type literals with shared constants without changing YAML/config behavior.

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/repository/source.go`
- Modify: `internal/repository/fetch.go`
- Modify tests as needed

**Step 1: Write failing/characterization test**

Add or extend config tests to assert constants equal persisted values:

```go
func TestRepoTypeConstantsMatchPersistedValues(t *testing.T) {
    cases := map[string]string{
        "topic":    config.RepoTypeTopic,
        "dir":      config.RepoTypeDir,
        "registry": config.RepoTypeRegistry,
        "skillhub": config.RepoTypeSkillHub,
    }
    for want, got := range cases {
        if got != want {
            t.Fatalf("expected %q, got %q", want, got)
        }
    }
}
```

Expected first run: FAIL because constants do not exist.

**Step 2: Add constants**

In `internal/config/config.go`, near `Repo`:

```go
const (
    RepoTypeTopic    = "topic"
    RepoTypeDir      = "dir"
    RepoTypeRegistry = "registry"
    RepoTypeSkillHub = "skillhub"
)
```

Update comments:

```go
Type string `yaml:"type"` // one of RepoTypeTopic, RepoTypeDir, RepoTypeRegistry, RepoTypeSkillHub
```

**Step 3: Replace source-related literals**

Use constants in:

```text
internal/repository/source.go
internal/repository/fetch.go
```

Do not do broad whole-repo churn unless tests make it safe.

**Step 4: Verify**

Run:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/config ./internal/repository -count=1
go test ./...
```

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go internal/repository/source.go internal/repository/fetch.go
git commit -m "refactor: define repo type constants"
```

---

### Task 3: Characterize `FetchSkills` source dispatch before deeper unification

**Objective:** Add tests that lock down existing fetch/list behavior, especially the important difference between search-time `dir` behavior and fetch-time git-first behavior.

**Files:**
- Modify: `internal/repository/fetch_test.go`
- Modify: `internal/repository/fetch.go` only if a tiny no-behavior refactor is safe

**Step 1: Add tests for existing behavior**

Add tests for:

- unknown repo type returns `unknown repository type`
- invalid `dir` URL returns `invalid repository URL format`
- non-dir `FetchSkillsViaGit` returns `git fetch only supports 'dir' type repos`
- registry dispatch still returns registry entries

Use existing helpers where possible.

**Step 2: Run tests**

Run:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository -run 'TestFetchSkills|TestFetchSkillsViaGit' -count=1
```

Expected: tests should pass against current behavior unless test names/imports are wrong.

**Step 3: Optional tiny refactor**

Only after characterization passes, replace raw literals with config constants if Task 2 did not already cover them.

**Step 4: Verify and commit**

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./internal/repository -count=1
go test ./...
git add internal/repository/fetch_test.go internal/repository/fetch.go
git commit -m "test: characterize source fetch dispatch"
```

---

### Task 4: Introduce internal source result/install reference types without wiring them yet

**Objective:** Add neutral types that can represent GitHub, registry, SkillHub, and future Hermes results, but do not change CLI behavior yet.

**Files:**
- Create: `internal/source/types.go` or `internal/repository/types.go`
- Create: corresponding test file

**Preferred package:** `internal/source` if we want a clean conceptual boundary; `internal/repository` if we want minimal churn. Choose one after Task 1/2 results.

**Suggested types:**

```go
type InstallKind string

const (
    InstallKindGitHub      InstallKind = "github"
    InstallKindURL         InstallKind = "url"
    InstallKindLocal       InstallKind = "local"
    InstallKindGenerated   InstallKind = "generated"
    InstallKindUnsupported InstallKind = "unsupported"
)

type InstallRef struct {
    Kind       InstallKind
    Reference  string
    Source     string
    Identifier string
    TrustLevel string
}

type SkillResult struct {
    Name        string
    Description string
    Source      string
    Identifier  string
    Installable bool
    TrustLevel  string
    Tags        []string
    InstallRef  *InstallRef
}
```

**Testing:** simple zero-value/constant tests only. No behavioral wiring yet.

**Commit:**

```bash
git commit -m "refactor: add source result types"
```

---

### Task 5: Characterize installer input resolution before extraction

**Objective:** Add tests around existing install target resolution behavior before extracting source-aware resolver logic.

**Files:**
- Modify: `internal/installer/installer_test.go`
- Modify production code only if testability seams are needed and behavior remains unchanged

**Behaviors to characterize:**

- GitHub `owner/repo` input parses as repo install.
- GitHub `owner/repo/path/to/skill` input parses as subdir install.
- GitHub tree URL parses branch/path correctly.
- Bare skill name resolution via cache behaves as expected.
- Ambiguous bare skill names return the existing ambiguity error.
- SkillHub slug fallback is called only after local cache miss.

**Important:** This is likely larger/riskier than Tasks 1-4. Do not start unless enough time/context remains to finish and verify.

---

## Final Verification

After each committed slice:

```bash
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
go test ./...
git status --short --branch
```

Before starting Hermes integration, verify:

- existing tests pass
- search behavior is unchanged
- fetch/list behavior is characterized
- installer resolution has tests or remains untouched
- source/result types can represent `resolved_github_id`, trust level, and unsupported entries

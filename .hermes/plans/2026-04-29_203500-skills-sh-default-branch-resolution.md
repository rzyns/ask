# skills.sh Default Branch Resolution Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Make `skills.sh` public-search GitHub skill path resolution use the repository default branch instead of assuming `main`.

**Architecture:** Keep `skills.sh` as catalog/discovery metadata only. For trusted public legacy search results with bare `owner/repo` sources, resolve the GitHub repository default branch through the GitHub REST repo endpoint, then fetch the recursive git tree for that branch and derive the native ASK install ref from the matching `SKILL.md` directory. Preserve existing unsupported/ambiguous behavior when GitHub lookup fails.

**Tech Stack:** Go, `net/http`, existing repository package tests, `go test`.

---

## Acceptance Criteria

- `resolveSkillsSHGitHubSkillPath` calls `GET https://api.github.com/repos/{owner}/{repo}` before tree lookup.
- The repo endpoint `default_branch` value is used for the subsequent tree request and resulting install ref.
- A repo with `default_branch: "trunk"` can resolve to `https://github.com/owner/repo/tree/trunk/path/to/skill`.
- Empty/missing default branch remains unsupported with existing clear reason.
- HTTP failures remain non-fatal to search and produce unsupported `skills.sh` candidates instead of panics.
- Existing `skills.sh` behavior and tests continue passing.

## Task 1: Add failing default-branch resolver test

**Objective:** Prove the resolver uses GitHub's repository default branch instead of hardcoded `main`.

**Files:**
- Modify: `internal/repository/skills_sh_test.go`

**Steps:**
1. Add a test that temporarily rewires HTTP requests through an `httptest.Server` by swapping `skillsSHGitHubAPIBaseURL`.
2. Make the fake API respond to `/repos/owner/repo` with `{"default_branch":"trunk"}`.
3. Make the fake API respond to `/repos/owner/repo/git/trees/trunk?recursive=1` with a `SKILL.md` path.
4. Assert the returned install ref uses `/tree/trunk/` and that both endpoint paths were requested.
5. Run:
   ```bash
   go test ./internal/repository -run TestResolveSkillsSHGitHubSkillPathUsesDefaultBranch -count=1
   ```
   Expected before implementation: fail because the test seam/functionality does not exist yet or tree lookup still uses `main`.

## Task 2: Implement default branch API seam and lookup

**Objective:** Replace hardcoded `main` with a default-branch lookup while keeping behavior small and testable.

**Files:**
- Modify: `internal/repository/skills_sh.go`

**Steps:**
1. Add package variable:
   ```go
   skillsSHGitHubAPIBaseURL = "https://api.github.com"
   ```
2. Add a small DTO:
   ```go
   type skillsSHGitHubRepo struct {
       DefaultBranch string `json:"default_branch"`
   }
   ```
3. Extract GitHub GET helper or add focused helper functions:
   - `fetchSkillsSHGitHubDefaultBranch(ownerRepo string) (string, bool)`
   - `fetchSkillsSHGitHubTreePaths(ownerRepo, branch string) ([]string, bool)`
4. Use URL construction that safely appends to the configured base URL.
5. Preserve `Accept: application/vnd.github.v3+json` and `User-Agent: ask-cli`.
6. If repo lookup fails or default branch is empty, return existing unsupported reason.
7. If tree lookup fails, return existing unsupported reason.

## Task 3: Add failure-mode tests

**Objective:** Lock down unsupported behavior for missing default branch and failed GitHub responses.

**Files:**
- Modify: `internal/repository/skills_sh_test.go`

**Steps:**
1. Add test for repo endpoint returning `{}`: expected no ref and reason contains `no GitHub skill path found`.
2. Add test for repo endpoint 404 or tree endpoint 404: expected no ref and same non-fatal reason.
3. Run:
   ```bash
   go test ./internal/repository -run 'TestResolveSkillsSHGitHubSkillPath.*DefaultBranch|TestResolveSkillsSHGitHubSkillPath.*GitHubFailure' -count=1
   ```

## Task 4: Run verification gates and smoke check

**Objective:** Verify the slice did not regress search/install behavior.

**Commands:**
```bash
gofmt -w internal/repository/skills_sh.go internal/repository/skills_sh_test.go
git diff --check
go test ./internal/repository -run SkillsSH -count=1
go test ./cmd -run 'TestDisplaySearchResultsShowsInstallRef|TestSkillsSHRepoSearchSelection|TestRepoInstallSelectionRecordsSkillsSHProvenanceForNativeRef|TestRepoNameInstallExpansionSkipsUnsupportedSkillsSHEntries' -count=1
go test ./internal/repository ./internal/config ./cmd -count=1
go test ./...
```

Optional live smoke, if network is available:
```bash
go run . --config /tmp/ask-skills-sh-smoke.yaml skill search grill-with-docs --remote --json
```

## Task 5: Review, debrief, commit

**Objective:** Finish with independent review, durable debrief, and clean commit.

**Steps:**
1. Run independent review focused on:
   - default branch correctness;
   - no credential leakage;
   - no regression to `npx`/`skills.sh` installer behavior;
   - unsupported behavior remains explicit.
2. Update `memory/2026-04-29.md` with a short debrief.
3. Commit after all gates pass:
   ```bash
   git add .hermes/plans/2026-04-29_203500-skills-sh-default-branch-resolution.md internal/repository/skills_sh.go internal/repository/skills_sh_test.go memory/2026-04-29.md
   git commit -m "fix: resolve skills.sh refs from default branch"
   ```

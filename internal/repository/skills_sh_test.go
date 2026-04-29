package repository

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestSkillsSHSearchUsesPublicLegacyEndpointWithoutAuth(t *testing.T) {
	t.Setenv("SKILLS_SH_API_KEY", "env-token")
	var gotPath, gotQuery, gotUA, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotUA = r.Header.Get("User-Agent")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"skills":[{"id":"vercel-labs/agent-skills/vercel-react-best-practices","skillId":"vercel-react-best-practices","name":"vercel-react-best-practices","installs":358797,"source":"vercel-labs/agent-skills"}],"count":1}`))
	}))
	defer server.Close()

	candidates, err := searchSkillsSHSource(context.Background(), config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "repo-token"}, "react")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if gotPath != "/api/search" || !strings.Contains(gotQuery, "q=react") || !strings.Contains(gotQuery, "limit=10") {
		t.Fatalf("request path/query = %q?%q", gotPath, gotQuery)
	}
	if gotUA != "ask-cli" {
		t.Fatalf("user-agent=%q", gotUA)
	}
	if gotAuth != "" {
		t.Fatalf("public search should not send Authorization header, got %q", gotAuth)
	}
	if len(candidates) != 1 || candidates[0].Name != "vercel-react-best-practices" {
		t.Fatalf("candidates=%#v", candidates)
	}
	if !candidates[0].Supported || candidates[0].Install.Value != "https://github.com/vercel-labs/agent-skills" {
		t.Fatalf("public search GitHub source was not natively installable: %#v", candidates[0])
	}
}

func TestSkillsSHFetchRequiresTokenForFullCatalog(t *testing.T) {
	t.Setenv("SKILLS_SH_API_KEY", "")
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	_, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL})
	if err == nil || !strings.Contains(err.Error(), "full catalog listing requires") || !strings.Contains(err.Error(), "SKILLS_SH_API_KEY") {
		t.Fatalf("expected clear full-catalog auth error, got %v", err)
	}
	if called {
		t.Fatal("full catalog fetch without token should fail before making HTTP request")
	}
}

func TestSkillsSHFetchRequestsAuthenticatedV1Endpoint(t *testing.T) {
	t.Setenv("SKILLS_SH_API_KEY", "")
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	if _, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "repo-token"}); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotPath != "/api/v1/skills" || !strings.Contains(gotQuery, "view=all-time") || !strings.Contains(gotQuery, "page=0") || !strings.Contains(gotQuery, "per_page=") {
		t.Fatalf("request path/query = %q?%q", gotPath, gotQuery)
	}
}

func TestSkillsSHAuthPrefersRepoTokenThenEnv(t *testing.T) {
	t.Setenv("SKILLS_SH_API_KEY", "env-token")
	var auths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auths = append(auths, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	if _, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "repo-token"}); err != nil {
		t.Fatal(err)
	}
	if _, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL}); err != nil {
		t.Fatal(err)
	}
	if auths[0] != "Bearer repo-token" || auths[1] != "Bearer env-token" {
		t.Fatalf("auths=%#v", auths)
	}
}

func TestSkillsSHHTTPErrorMessages(t *testing.T) {
	t.Setenv("SKILLS_SH_API_KEY", "sentinel-secret-token")
	for _, tc := range []struct {
		status     int
		retryAfter string
		want       string
	}{
		{http.StatusUnauthorized, "", "skills.sh API key required"},
		{http.StatusTooManyRequests, "12", "Retry-After: 12"},
		{http.StatusInternalServerError, "", "skills.sh API error: 500"},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tc.retryAfter != "" {
				w.Header().Set("Retry-After", tc.retryAfter)
			}
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`oops`))
		}))
		_, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL})
		server.Close()
		if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "sentinel-secret-token") {
			t.Fatalf("status %d err=%v", tc.status, err)
		}
	}
}

func TestSkillsSHHTTPErrorRedactsTokenFromResponseSnippet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`upstream echoed repo-secret-token`))
	}))
	defer server.Close()

	_, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "repo-secret-token"})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "repo-secret-token") || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("token was not redacted: %v", err)
	}
}

func TestSkillsSHLargeHTTPErrorStillReportsStatus(t *testing.T) {
	old := skillsSHMaxBodyBytes
	skillsSHMaxBodyBytes = 8
	t.Cleanup(func() { skillsSHMaxBodyBytes = old })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`response body larger than tiny test limit`))
	}))
	defer server.Close()

	_, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "token"})
	if err == nil || !strings.Contains(err.Error(), "Retry-After: 30") || strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected rate-limit error despite large body, got %v", err)
	}
}

func TestSkillsSHMalformedJSONFromHTTPReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"data":[`)) }))
	defer server.Close()
	if _, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "token"}); err == nil || !strings.Contains(err.Error(), "parse skills.sh") {
		t.Fatalf("err=%v", err)
	}
}

func TestSkillsSHBodyLimitEnforced(t *testing.T) {
	old := skillsSHMaxBodyBytes
	skillsSHMaxBodyBytes = 8
	t.Cleanup(func() { skillsSHMaxBodyBytes = old })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"data":[]} extra`)) }))
	defer server.Close()
	if _, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "token"}); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("err=%v", err)
	}
}

func TestSkillsSHSearchContextCancellationRespected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := searchSkillsSHSource(ctx, config.Repo{Type: config.RepoTypeSkillsSH, URL: "https://example.invalid", Token: "token"}, "react")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestSkillsSHHTTPClientHasTimeout(t *testing.T) {
	if skillsSHHTTPClient.Timeout <= 0 {
		t.Fatal("skills.sh HTTP client must have a timeout")
	}
}

func TestSkillsSHServiceUnavailableSuggestsRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`temporarily unavailable`))
	}))
	defer server.Close()

	_, err := fetchSkillsSHSource(config.Repo{Type: config.RepoTypeSkillsSH, URL: server.URL, Token: "token"})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "retry") {
		t.Fatalf("expected retry guidance for 503, got %v", err)
	}
}

func TestSkillsSHV1GitHubCandidateIsInstallable(t *testing.T) {
	body := []byte(`{"data":[{"id":"vercel-labs/agent-skills/next-js-development","slug":"next-js-development","name":"Next.js Development","description":"Build Next.js apps","source":"vercel-labs/agent-skills","installs":24531,"sourceType":"github","installUrl":"https://github.com/vercel-labs/agent-skills/tree/main/next-js-development","url":"https://skills.sh/vercel-labs/agent-skills/next-js-development"}],"count":1}`)
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("len=%d", len(candidates))
	}
	c := candidates[0]
	if !c.Supported {
		t.Fatalf("candidate unsupported: %q", c.UnsupportedReason)
	}
	if c.Install.Kind != InstallRefGitHubPath || c.Install.Value != "https://github.com/vercel-labs/agent-skills/tree/main/next-js-development" {
		t.Fatalf("install=%#v", c.Install)
	}
	if c.Name != "Next.js Development" || c.Description != "Build Next.js apps" || c.PageURL != "https://skills.sh/vercel-labs/agent-skills/next-js-development" {
		t.Fatalf("metadata=%#v", c)
	}
	if c.Stars != 24531 {
		t.Fatalf("stars=%d", c.Stars)
	}
	if c.Source != config.RepoTypeSkillsSH || c.SourceIdentifier != "vercel-labs/agent-skills/next-js-development" || c.UpdateStrategy != "skills.sh" {
		t.Fatalf("provenance=%#v", c)
	}
}

func TestSkillsSHGitHubURLValidationRejectsLookalikesAndHTTP(t *testing.T) {
	for _, installURL := range []string{"https://github.com.evil/owner/repo", "http://github.com/owner/repo", "https://github.com/owner"} {
		body := []byte(`{"data":[{"id":"bad","name":"Bad","sourceType":"github","installUrl":"` + installURL + `","source":"owner/repo"}]}`)
		candidates, err := parseSkillsSHV1Candidates(body)
		if err != nil {
			t.Fatalf("parse %s: %v", installURL, err)
		}
		if len(candidates) != 1 || candidates[0].Supported {
			t.Fatalf("%s unexpectedly supported: %#v", installURL, candidates)
		}
	}
}

func TestSkillsSHRepoRootDoesNotAppendSlug(t *testing.T) {
	body := []byte(`{"data":[{"id":"vercel-labs/agent-skills/next-js-development","slug":"next-js-development","name":"Next.js Development","source":"vercel-labs/agent-skills","sourceType":"github","installUrl":"https://github.com/vercel-labs/agent-skills"}]}`)
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := candidates[0].Install.Value; got != "https://github.com/vercel-labs/agent-skills" {
		t.Fatalf("install value=%q", got)
	}
}

func TestSkillsSHLegacySearchParsesReducedMetadata(t *testing.T) {
	body := []byte(`{"skills":[{"id":"gh","skillId":"agent-skills","name":"GitHub Skill","installs":7,"source":"github.com/vercel-labs/agent-skills"},{"id":"mintlify.com/mintlify","skillId":"mintlify","name":"Mintlify","installs":5,"source":"mintlify.com"}],"count":2}`)
	candidates, err := parseSkillsSHLegacySearchCandidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("len=%d", len(candidates))
	}
	if !candidates[0].Supported || candidates[0].Install.Value != "https://github.com/vercel-labs/agent-skills" {
		t.Fatalf("github legacy=%#v", candidates[0])
	}
	if candidates[1].Supported || candidates[1].UnsupportedReason == "" {
		t.Fatalf("domain legacy=%#v", candidates[1])
	}
}

func TestSkillsSHLegacySearchDoesNotCoerceBareSourceToGitHub(t *testing.T) {
	body := []byte(`{"skills":[{"id":"ambiguous","skillId":"bar","name":"Ambiguous","source":"foo/bar"}],"count":1}`)
	candidates, err := parseSkillsSHLegacySearchCandidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("len=%d", len(candidates))
	}
	if candidates[0].Supported || candidates[0].Install.Kind != InstallRefUnsupported || candidates[0].UnsupportedReason == "" {
		t.Fatalf("ambiguous legacy source unexpectedly installable: %#v", candidates[0])
	}
}

func TestSkillsSHDomainEntriesRemainVisibleUnsupported(t *testing.T) {
	body := []byte(`{"data":[{"id":"mintlify.com/mintlify","slug":"mintlify","name":"Mintlify","source":"mintlify.com","sourceType":"website","installs":5}]}`)
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(candidates) != 1 || candidates[0].Supported || candidates[0].UnsupportedReason == "" {
		t.Fatalf("candidate=%#v", candidates)
	}
}

func TestSkillsSHWellKnownSkillMDAndArchiveUnsupported(t *testing.T) {
	body := []byte(`{"data":[{"id":"one","name":"One","type":"skill-md","source":"https://github.com/owner/repo"},{"id":"two","name":"Two","type":"archive","source":"https://github.com/owner/repo"}]}`)
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, c := range candidates {
		if c.Supported || c.UnsupportedReason == "" {
			t.Fatalf("future artifact supported: %#v", c)
		}
	}
}

func TestSkillsSHMalformedJSONReturnsError(t *testing.T) {
	if _, err := parseSkillsSHV1Candidates([]byte(`{"data":[`)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := parseSkillsSHLegacySearchCandidates([]byte(`{"skills":[`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestSkillsSHDuplicatesSkippedByDefault(t *testing.T) {
	body := []byte(`{"data":[{"id":"dupe","name":"Dupe","sourceType":"github","installUrl":"https://github.com/owner/repo","isDuplicate":true},{"id":"keep","name":"Keep","sourceType":"github","installUrl":"https://github.com/owner/repo2"}]}`)
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(candidates) != 1 || candidates[0].SourceIdentifier != "keep" {
		t.Fatalf("candidates=%#v", candidates)
	}
}

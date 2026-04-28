package repository

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestFetchSkillsHermesSourceFetchesIndexAndIgnoresUnsupportedEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"skills": [
				{"name":"Alpha","description":"GitHub skill","repo":"owner/repo","path":"skills/alpha"},
				{"name":"Ignored","description":"External skill","url":"https://example.com/owner/repo"}
			]
		}`))
	}))
	defer server.Close()

	results, err := FetchSkills(config.Repo{Type: config.RepoTypeHermes, URL: server.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 safely GitHub-resolvable result, got %d: %#v", len(results), results)
	}
	if results[0].Name != "Alpha" || results[0].FullName != "owner/repo/skills/alpha" {
		t.Fatalf("unexpected result: %#v", results[0])
	}
	if results[0].HTMLURL != "owner/repo/skills/alpha" || results[0].Source != config.RepoTypeHermes {
		t.Fatalf("unexpected adapted repository fields: %#v", results[0])
	}
}

func TestSearchSkillsHermesSourceAppliesKeywordFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"skills": [
				{"name":"Alpha","description":"First skill","resolved_github_id":"owner/repo/skills/alpha"},
				{"name":"Beta","description":"Marketing automation","resolved_github_id":"owner/repo/skills/beta"}
			]
		}`))
	}))
	defer server.Close()

	results, err := SearchSkills(context.Background(), config.Repo{Type: config.RepoTypeHermes, URL: server.URL}, "market")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Beta" {
		t.Fatalf("expected only Beta keyword match, got %#v", results)
	}
}

func TestSearchSkillsHermesSourceReturnsMalformedJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"skills": [`))
	}))
	defer server.Close()

	_, err := SearchSkills(context.Background(), config.Repo{Type: config.RepoTypeHermes, URL: server.URL}, "")
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func TestFetchSkillsHermesSourceReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchSkills(config.Repo{Type: config.RepoTypeHermes, URL: server.URL})
	if err == nil {
		t.Fatal("expected HTTP error")
	}
}

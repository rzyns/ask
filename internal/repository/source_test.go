package repository

import (
	"context"
	"encoding/json"
	"net/http"
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

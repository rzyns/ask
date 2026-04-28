package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

func TestSourceDispatcherFetchTopicUsesEmptyKeyword(t *testing.T) {
	origSearchTopic := searchTopicFunc
	t.Cleanup(func() { searchTopicFunc = origSearchTopic })

	var gotTopic, gotKeyword string
	searchTopicFunc = func(topic, keyword string) ([]github.Repository, error) {
		gotTopic = topic
		gotKeyword = keyword
		return []github.Repository{{Name: "topic-skill"}}, nil
	}

	source, err := sourceForRepo(config.Repo{Type: config.RepoTypeTopic})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := source.Fetch(config.Repo{Type: config.RepoTypeTopic, URL: "agent-skill"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Name != "topic-skill" {
		t.Fatalf("unexpected results: %#v", results)
	}
	if gotTopic != "agent-skill" || gotKeyword != "" {
		t.Fatalf("expected topic fetch with empty keyword, got topic=%q keyword=%q", gotTopic, gotKeyword)
	}
}

func TestFetchSkillsRoutesThroughSourceDispatcher(t *testing.T) {
	origSearchTopic := searchTopicFunc
	t.Cleanup(func() { searchTopicFunc = origSearchTopic })

	called := false
	searchTopicFunc = func(topic, keyword string) ([]github.Repository, error) {
		called = true
		if topic != "agent-skill" || keyword != "" {
			t.Fatalf("expected FetchSkills topic dispatch with empty keyword, got topic=%q keyword=%q", topic, keyword)
		}
		return []github.Repository{{Name: "fetch-topic-skill"}}, nil
	}

	results, err := FetchSkills(config.Repo{Type: config.RepoTypeTopic, URL: "agent-skill"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected FetchSkills to call dispatcher source")
	}
	if len(results) != 1 || results[0].Name != "fetch-topic-skill" {
		t.Fatalf("unexpected results: %#v", results)
	}
}

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

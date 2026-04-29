package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

func TestSourceForRepoRecognizesSkillsSH(t *testing.T) {
	_, err := sourceForRepo(config.Repo{Type: config.RepoTypeSkillsSH})
	if err != nil {
		t.Fatalf("expected skills.sh source to be recognized, got %v", err)
	}
}

func TestSearchSkillsSkillsSHDispatchesThroughSeam(t *testing.T) {
	origSearch := searchSkillsSHFunc
	t.Cleanup(func() { searchSkillsSHFunc = origSearch })

	called := false
	searchSkillsSHFunc = func(ctx context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
		called = true
		if repo.Type != config.RepoTypeSkillsSH || repo.URL != "skills.sh" || keyword != "docker" {
			t.Fatalf("unexpected skills.sh search args: repo=%#v keyword=%q", repo, keyword)
		}
		return []SkillCandidate{{Name: "skills-sh-search", Install: InstallRef{Kind: InstallRefSlug, Value: "skills-sh-search"}}}, nil
	}

	results, err := SearchSkills(context.Background(), config.Repo{Type: config.RepoTypeSkillsSH, URL: "skills.sh"}, "docker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected SearchSkills to dispatch through skills.sh seam")
	}
	if len(results) != 1 || results[0].Name != "skills-sh-search" {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestFetchSkillsSkillsSHDispatchesThroughSeam(t *testing.T) {
	origFetch := fetchSkillsSHFunc
	t.Cleanup(func() { fetchSkillsSHFunc = origFetch })

	called := false
	fetchSkillsSHFunc = func(repo config.Repo) ([]SkillCandidate, error) {
		called = true
		if repo.Type != config.RepoTypeSkillsSH || repo.URL != "skills.sh" {
			t.Fatalf("unexpected skills.sh fetch args: repo=%#v", repo)
		}
		return []SkillCandidate{{Name: "skills-sh-fetch", Install: InstallRef{Kind: InstallRefSlug, Value: "skills-sh-fetch"}}}, nil
	}

	results, err := FetchSkills(config.Repo{Type: config.RepoTypeSkillsSH, URL: "skills.sh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected FetchSkills to dispatch through skills.sh seam")
	}
	if len(results) != 1 || results[0].Name != "skills-sh-fetch" {
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

func TestSearchSkillsDirRejectsBundledHermesSkills(t *testing.T) {
	origSearchDir := searchDirFunc
	t.Cleanup(func() { searchDirFunc = origSearchDir })
	searchDirFunc = func(_, _, _ string) ([]github.Repository, error) {
		t.Fatal("bundled Hermes dir source should be rejected before GitHub search")
		return nil, nil
	}

	_, err := SearchSkills(context.Background(), config.Repo{
		Type: config.RepoTypeDir,
		URL:  "NousResearch/hermes-agent/skills",
	}, "foo")
	if err == nil {
		t.Fatal("expected bundled Hermes skills error")
	}
	if got := err.Error(); !strings.Contains(got, "bundled Hermes skills") {
		t.Fatalf("expected bundled Hermes skills error, got %q", got)
	}
}

func TestSearchSkillsDirRejectsBundledHermesSkillChildren(t *testing.T) {
	origSearchDir := searchDirFunc
	t.Cleanup(func() { searchDirFunc = origSearchDir })
	searchDirFunc = func(_, _, _ string) ([]github.Repository, error) {
		t.Fatal("bundled Hermes dir source should be rejected before GitHub search")
		return nil, nil
	}

	_, err := SearchSkills(context.Background(), config.Repo{
		Type: config.RepoTypeDir,
		URL:  "NousResearch/hermes-agent/skills/core-skill",
	}, "foo")
	if err == nil {
		t.Fatal("expected bundled Hermes skills error")
	}
	if got := err.Error(); !strings.Contains(got, "bundled Hermes skills") {
		t.Fatalf("expected bundled Hermes skills error, got %q", got)
	}
}

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yeasy/ask/internal/config"
)

func TestShowHermesAgentSkillsIncludesOwnershipFields(t *testing.T) {
	root := t.TempDir()
	writeHermesListSkill(t, root, "managed-skill", "managed-skill", "Managed description", "")
	writeHermesListSkill(t, root, filepath.Join("research", "native-skill"), "native-skill", "Native description", "2.0.0")

	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "managed-skill",
		Source:      "hermes-index",
		URL:         "https://github.com/example/managed-skill.git",
		Version:     "1.2.3",
		InstalledAt: time.Unix(0, 0),
	}}}

	items := listHermesItemsForTest(t, root, "Project", lock)
	byName := hermesListItemsByName(items)
	if len(byName) != 2 {
		t.Fatalf("got %d items %#v, want 2", len(items), items)
	}

	managed := byName["managed-skill"]
	if managed.Name != "managed-skill" || managed.Agent != "hermes" || managed.Scope != "Project" {
		t.Fatalf("managed identity fields = %#v", managed)
	}
	if managed.ManagedBy != "ask" || managed.Status != "installed" || managed.Source != "hermes-index" || managed.Update != "current" || managed.Version != "1.2.3" {
		t.Fatalf("managed ownership fields = %#v, want ask/installed/hermes-index/current/1.2.3", managed)
	}

	native := byName["native-skill"]
	if native.Name != "native-skill" || native.ManagedBy != "hermes-native" || native.Status != "installed" || native.Source != "local" || native.Update != "none" || native.Version != "2.0.0" {
		t.Fatalf("native fields = %#v, want hermes-native installed local none 2.0.0", native)
	}
}

func TestShowHermesAgentSkillsJSONFields(t *testing.T) {
	root := t.TempDir()
	writeHermesListSkill(t, root, "managed-skill", "managed-skill", "Managed description", "")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "managed-skill", Source: "hermes-index", URL: "https://example.test/repo.git", Version: "1.2.3"}}}

	items := listHermesItemsForTest(t, root, "Global", lock)
	data, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	jsonText := string(data)
	for _, want := range []string{`"agent":"hermes"`, `"managed_by":"ask"`, `"status":"installed"`, `"source":"hermes-index"`, `"update":"current"`, `"path":"`} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("JSON %s missing %s", jsonText, want)
		}
	}
}

func TestBuildHermesSkillListItemsReturnsScanErrors(t *testing.T) {
	root := t.TempDir()
	writeHermesListSkill(t, root, "bad", "bad", "Bad", "")
	if err := os.WriteFile(filepath.Join(root, "bad", "SKILL.md"), []byte("---\nname: [unterminated\n---\n"), 0o644); err != nil {
		t.Fatalf("write malformed SKILL.md: %v", err)
	}

	_, err := buildHermesSkillListItems("hermes", root, "Project", nil)
	if err == nil {
		t.Fatal("buildHermesSkillListItems returned nil error for malformed SKILL.md")
	}
}

func TestShowHermesAgentSkillsMissingDirectoryIsEmpty(t *testing.T) {
	items := listHermesItemsForTest(t, filepath.Join(t.TempDir(), "missing"), "Project", nil)
	if len(items) != 0 {
		t.Fatalf("got %#v, want no items", items)
	}
}

func listHermesItemsForTest(t *testing.T, dir, scope string, lock *config.LockFile) []SkillListItem {
	t.Helper()
	items, err := buildHermesSkillListItems("hermes", dir, scope, lock)
	if err != nil {
		t.Fatalf("buildHermesSkillListItems returned error: %v", err)
	}
	return items
}

func hermesListItemsByName(items []SkillListItem) map[string]SkillListItem {
	byName := make(map[string]SkillListItem, len(items))
	for _, item := range items {
		byName[item.Name] = item
	}
	return byName
}

func writeHermesListSkill(t *testing.T, root, rel, name, description, version string) {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	frontmatter := "---\nname: " + name + "\ndescription: " + description + "\n"
	if version != "" {
		frontmatter += "version: " + version + "\n"
	}
	frontmatter += "---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(frontmatter), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

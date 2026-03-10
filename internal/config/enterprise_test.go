package config

import (
	"os"
	"testing"
)

func TestIsSourceAllowed_EmptyPatterns(t *testing.T) {
	if !IsSourceAllowed("anything", nil) {
		t.Error("empty patterns should allow all sources")
	}
	if !IsSourceAllowed("anything", []string{}) {
		t.Error("empty patterns should allow all sources")
	}
}

func TestIsSourceAllowed_ExactMatch(t *testing.T) {
	patterns := []string{"anthropics/skills"}
	if !IsSourceAllowed("https://github.com/anthropics/skills.git", patterns) {
		t.Error("should match with full URL and .git suffix")
	}
	if !IsSourceAllowed("https://github.com/anthropics/skills", patterns) {
		t.Error("should match with full URL")
	}
	if !IsSourceAllowed("anthropics/skills", patterns) {
		t.Error("should match shorthand")
	}
}

func TestIsSourceAllowed_WildcardMatch(t *testing.T) {
	patterns := []string{"anthropics/*", "company-org/*"}

	if !IsSourceAllowed("https://github.com/anthropics/skills", patterns) {
		t.Error("should match anthropics/* pattern")
	}
	if !IsSourceAllowed("https://github.com/company-org/internal-skills", patterns) {
		t.Error("should match company-org/* pattern")
	}
	if IsSourceAllowed("https://github.com/evil-org/malware", patterns) {
		t.Error("should not match evil-org")
	}
}

func TestIsSourceAllowed_Blocked(t *testing.T) {
	patterns := []string{"trusted-org/*"}
	if IsSourceAllowed("https://github.com/untrusted/repo", patterns) {
		t.Error("should block untrusted source")
	}
}

func TestEnterpriseConfig_YAMLParsing(t *testing.T) {
	yamlContent := `
version: "1.2"
enterprise:
  allowed_sources:
    - "anthropics/*"
    - "company-org/*"
  require_check: true
  require_lock: true
  private_registry: "https://github.example.com/api/v3"
skills: []
repos: []
`
	dir := t.TempDir()
	writeTestFile(t, dir, "ask.yaml", yamlContent)

	cfg, err := loadConfigFromPath(dir + "/ask.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Enterprise == nil {
		t.Fatal("enterprise config should not be nil")
	}
	if len(cfg.Enterprise.AllowedSources) != 2 {
		t.Errorf("expected 2 allowed sources, got %d", len(cfg.Enterprise.AllowedSources))
	}
	if !cfg.Enterprise.RequireCheck {
		t.Error("require_check should be true")
	}
	if !cfg.Enterprise.RequireLock {
		t.Error("require_lock should be true")
	}
	if cfg.Enterprise.PrivateRegistry != "https://github.example.com/api/v3" {
		t.Errorf("unexpected private registry: %s", cfg.Enterprise.PrivateRegistry)
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := dir + "/" + name
	if err := writeFile(path, content); err != nil {
		t.Fatal(err)
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

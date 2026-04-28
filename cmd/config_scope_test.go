package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

func TestLoadConfigForCommandUsesGlobalScope(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	cwd := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("chdir temp cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := os.WriteFile("ask.yaml", []byte("version: \"1.2\"\nrepos:\n  - name: local-only\n    type: dir\n    url: local/repo\n"), 0o600); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	globalDir := filepath.Join(tmp, ".ask")
	if err := os.MkdirAll(globalDir, 0o700); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte("version: \"1.2\"\nrepos:\n  - name: hermes-index\n    type: hermes\n    url: https://example.test/skills-index.json\n"), 0o600); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("global", "g", true, "")

	cfg, err := loadConfigForCommand(cmd)
	if err != nil {
		t.Fatalf("loadConfigForCommand returned error: %v", err)
	}
	if len(cfg.Repos) == 0 || cfg.Repos[0].Name != "hermes-index" {
		t.Fatalf("expected global Hermes repo first, got %#v", cfg.Repos)
	}
}

func TestLoadConfigForCommandUsesExplicitConfigFile(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "custom.yaml")
	if err := os.WriteFile(customPath, []byte("version: \"1.2\"\nrepos:\n  - name: explicit-repo\n    type: hermes\n    url: https://example.test/skills-index.json\n"), 0o600); err != nil {
		t.Fatalf("write explicit config: %v", err)
	}

	oldCfgFile := cfgFile
	cfgFile = customPath
	t.Cleanup(func() { cfgFile = oldCfgFile })

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("global", "g", false, "")

	cfg, err := loadConfigForCommand(cmd)
	if err != nil {
		t.Fatalf("loadConfigForCommand returned error: %v", err)
	}
	if len(cfg.Repos) == 0 || cfg.Repos[0].Name != "explicit-repo" {
		t.Fatalf("expected explicit config repo first, got %#v", cfg.Repos)
	}
}

func TestSaveConfigForCommandUsesExplicitConfigFile(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "custom.yaml")
	oldCfgFile := cfgFile
	cfgFile = customPath
	t.Cleanup(func() { cfgFile = oldCfgFile })

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("global", "g", false, "")

	cfg := config.DefaultConfig()
	cfg.Repos = append(cfg.Repos, config.Repo{Name: "explicit-saved", Type: "hermes", URL: "https://example.test/index.json"})

	if err := saveConfigForCommand(cmd, &cfg); err != nil {
		t.Fatalf("saveConfigForCommand returned error: %v", err)
	}
	saved, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("read saved explicit config: %v", err)
	}
	if !strings.Contains(string(saved), "explicit-saved") {
		t.Fatalf("expected explicit config file to contain saved repo, got:\n%s", string(saved))
	}
}

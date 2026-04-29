package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallAliases(t *testing.T) {
	// Test skill install aliases
	foundAdd := false
	foundI := false
	for _, alias := range installCmd.Aliases {
		if alias == "add" {
			foundAdd = true
		}
		if alias == "i" {
			foundI = true
		}
	}
	if !foundAdd {
		t.Error("installCmd should have 'add' alias")
	}
	if !foundI {
		t.Error("installCmd should have 'i' alias")
	}

	// Test root install aliases
	foundAddRoot := false
	foundIRoot := false
	for _, alias := range installRootCmd.Aliases {
		if alias == "add" {
			foundAddRoot = true
		}
		if alias == "i" {
			foundIRoot = true
		}
	}
	if !foundAddRoot {
		t.Error("installRootCmd should have 'add' alias")
	}
	if !foundIRoot {
		t.Error("installRootCmd should have 'i' alias")
	}
}

func TestInstallFlags(t *testing.T) {
	// Need to import pflag? No, cmd.Flags() returns *pflag.FlagSet.
	// But I don't want to import pflag explicitly if not needed.
	// Just inline checks.

	if val := installCmd.Flags().Lookup("agent"); val == nil {
		t.Error("installCmd missing 'agent' flag")
	} else if val.Shorthand != "a" {
		t.Error("installCmd 'agent' flag shorthand should be 'a'")
	}

	// Global flag is persistent on root, so it applies to installRootCmd
	if val := rootCmd.PersistentFlags().Lookup("global"); val == nil {
		t.Error("rootCmd missing 'global' persistent flag")
	}
}

func TestInstallValidation(t *testing.T) {
	// Verify Args validation (Minimum 0 args)
	err := installCmd.Args(installCmd, []string{})
	if err != nil {
		t.Errorf("installCmd should accept 0 args, got error: %v", err)
	}

	err = installCmd.Args(installCmd, []string{"some-skill"})
	if err != nil {
		t.Errorf("installCmd should accept 1 arg, got error: %v", err)
	}
}

func TestLoadInstallConfigUsesCommandScope(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	globalDir := filepath.Join(tmp, ".ask")
	if err := os.MkdirAll(globalDir, 0o700); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte("version: \"1.2\"\nrepos:\n  - name: hermes-index\n    type: hermes\n    url: https://example.test/skills-index.json\n"), 0o600); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("global", "g", true, "")

	cfg := loadInstallConfig(cmd)
	if len(cfg.Repos) == 0 || cfg.Repos[0].Name != "hermes-index" {
		t.Fatalf("expected install config loader to honor --global scope, got %#v", cfg.Repos)
	}
}

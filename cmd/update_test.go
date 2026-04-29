package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/hermes"
	"github.com/yeasy/ask/internal/installer"
)

func TestUpdateFlagsIncludeHermesAgentAndForce(t *testing.T) {
	if flag := updateCmd.Flags().Lookup("agent"); flag == nil {
		t.Fatal("updateCmd missing --agent flag")
	}
	if flag := updateCmd.Flags().Lookup("force"); flag == nil {
		t.Fatal("updateCmd missing --force flag")
	}
}

func TestSkillUpdateHermesUsesSafePlanAndInstaller(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)

	source := writeCmdUpdateSkill(t, filepath.Join(home, ".ask", "skills"), "gitnexus-explorer", "old")
	writeCmdUpdateSkill(t, filepath.Join(hermesHome, "skills"), "gitnexus-explorer", "old")
	checksum, err := hermesDirectoryChecksumForUpdateTest(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "gitnexus-explorer",
		Source:           config.RepoTypeHermes,
		URL:              "https://github.com/NousResearch/hermes-agent",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Ownership:        string(hermes.HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/gitnexus-explorer",
		SourcePath:       source,
		TargetPath:       filepath.Join(hermesHome, "skills", "gitnexus-explorer"),
		Checksum:         checksum,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".ask", "config.yaml"), []byte("version: \"1.2\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var gotInput string
	var gotOpts installer.InstallOptions
	oldInstall := installHermesUpdate
	installHermesUpdate = func(input string, opts installer.InstallOptions) error {
		gotInput = input
		gotOpts = opts
		return nil
	}
	t.Cleanup(func() { installHermesUpdate = oldInstall })

	buf, err := executeUpdateCommandForTest("skill", "update", "gitnexus-explorer", "--agent", "hermes", "--global")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if gotInput != "official/gitnexus-explorer" {
		t.Fatalf("installer input = %q", gotInput)
	}
	if !gotOpts.Global || len(gotOpts.Agents) != 1 || gotOpts.Agents[0] != "hermes" || !gotOpts.ReplaceExisting || !gotOpts.SuppressGenericLockEntry || gotOpts.ReplaceExistingName != "gitnexus-explorer" || gotOpts.ReplaceExistingSource != source || gotOpts.ReplaceExistingTarget != filepath.Join(hermesHome, "skills", "gitnexus-explorer") {
		t.Fatalf("installer opts = %#v", gotOpts)
	}
	if gotOpts.SourceMetadata == nil || gotOpts.SourceMetadata.UpdateStrategy != "hermes-index" {
		t.Fatalf("source metadata = %#v", gotOpts.SourceMetadata)
	}
}

func TestTopLevelUpdateHermesAliasUsesSameInstallerPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	source := writeCmdUpdateSkill(t, filepath.Join(home, ".ask", "skills"), "alias-skill", "old")
	writeCmdUpdateSkill(t, filepath.Join(hermesHome, "skills"), "alias-skill", "old")
	checksum, err := hermesDirectoryChecksumForUpdateTest(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "alias-skill",
		Source:           config.RepoTypeHermes,
		URL:              "https://github.com/NousResearch/hermes-agent/optional-skills/alias-skill",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Ownership:        string(hermes.HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/alias-skill",
		SourcePath:       source,
		TargetPath:       filepath.Join(hermesHome, "skills", "alias-skill"),
		Checksum:         checksum,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".ask", "config.yaml"), []byte("version: \"1.2\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var gotInput string
	oldInstall := installHermesUpdate
	installHermesUpdate = func(input string, opts installer.InstallOptions) error {
		gotInput = input
		return nil
	}
	t.Cleanup(func() { installHermesUpdate = oldInstall })

	buf, err := executeUpdateCommandForTest("update", "alias-skill", "--agent", "hermes", "--global", "--force")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if gotInput != "official/alias-skill" {
		t.Fatalf("installer input = %q", gotInput)
	}
}

func TestSkillUpdateHermesImportedNamedReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("HERMES_HOME", filepath.Join(home, "hermes-home"))
	target := writeCmdUpdateSkill(t, filepath.Join(home, "hermes-home", "skills"), "local-only", "old")
	checksum, err := hermesDirectoryChecksumForUpdateTest(target)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:           "local-only",
		InstalledAt:    time.Now().UTC(),
		Agent:          "hermes",
		Ownership:      string(hermes.HermesSkillOwnershipImported),
		InstallMode:    "in-place",
		UpdateStrategy: "none",
		TargetPath:     target,
		Checksum:       checksum,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	buf, err := executeUpdateCommandForTest("skill", "update", "local-only", "--agent", "hermes", "--global")
	if err == nil {
		t.Fatalf("expected error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "update unavailable") {
		t.Fatalf("error %q missing update unavailable", err.Error())
	}
}

func executeUpdateCommandForTest(args ...string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	_ = updateCmd.Flags().Set("agent", "")
	_ = updateCmd.Flags().Set("force", "false")
	_ = rootCmd.PersistentFlags().Set("global", "false")
	return &buf, rootCmd.Execute()
}

func writeCmdUpdateSkill(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n"+body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func hermesDirectoryChecksumForUpdateTest(root string) (string, error) {
	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, path := range files {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, filepath.ToSlash(rel))
		_, _ = io.WriteString(h, "\x00")
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return "", err
		}
		if err := f.Close(); err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, "\x00")
	}
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil))), nil
}

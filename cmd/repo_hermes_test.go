package cmd

import (
	"strings"
	"testing"
)

func TestRejectBundledHermesRepoSourceRejectsBundledSkills(t *testing.T) {
	err := rejectBundledHermesRepoSource("NousResearch", "hermes-agent", "skills")
	if err == nil {
		t.Fatal("expected bundled Hermes skills error")
	}
	if got := err.Error(); !strings.Contains(got, "bundled Hermes skills") || !strings.Contains(got, "hermes-index") {
		t.Fatalf("expected bundled Hermes guidance, got %q", got)
	}
}

func TestRejectBundledHermesRepoSourceRejectsBundledSkillChildren(t *testing.T) {
	err := rejectBundledHermesRepoSource("NousResearch", "hermes-agent", "skills/core-skill")
	if err == nil {
		t.Fatal("expected bundled Hermes skills error")
	}
	if got := err.Error(); !strings.Contains(got, "bundled Hermes skills") {
		t.Fatalf("expected bundled Hermes guidance, got %q", got)
	}
}

func TestRejectBundledHermesRepoSourceAllowsOptionalSkills(t *testing.T) {
	if err := rejectBundledHermesRepoSource("NousResearch", "hermes-agent", "optional-skills/research/gitnexus-explorer"); err != nil {
		t.Fatalf("expected optional Hermes skills to be allowed, got %v", err)
	}
}

package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSkillMD_MetadataAllowsNestedJSONLikeValues(t *testing.T) {
	skillDir := t.TempDir()
	skillMD := `---
name: nested-metadata
description: Skill with Hermes-style nested metadata
metadata:
  category: research
  enabled: true
  priority: 7
  confidence: 0.75
  nothing: null
  labels:
    - git
    - 42
    - false
  owner:
    name: Hermes
    level: 2
---
# nested-metadata
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	meta, err := ParseSkillMD(skillDir)
	if err != nil {
		t.Fatalf("ParseSkillMD returned error: %v", err)
	}

	if got, ok := meta.Metadata.String("category"); !ok || got != "research" {
		t.Fatalf("metadata.category = %q/%v, want research/true", got, ok)
	}
	if got, ok := meta.Metadata["enabled"].Bool(); !ok || !got {
		t.Fatalf("metadata.enabled = %v/%v, want true/true", got, ok)
	}
	if got, ok := meta.Metadata["priority"].Number(); !ok || got != "7" {
		t.Fatalf("metadata.priority = %q/%v, want 7/true", got, ok)
	}
	if got, ok := meta.Metadata["confidence"].Number(); !ok || got != "0.75" {
		t.Fatalf("metadata.confidence = %q/%v, want 0.75/true", got, ok)
	}
	if !meta.Metadata["nothing"].IsNull() {
		t.Fatalf("metadata.nothing kind = %q, want null", meta.Metadata["nothing"].Kind())
	}
	labels, ok := meta.Metadata["labels"].Array()
	if !ok || len(labels) != 3 {
		t.Fatalf("metadata.labels = %#v/%v, want 3 item array", labels, ok)
	}
	owner, ok := meta.Metadata["owner"].Object()
	if !ok {
		t.Fatalf("metadata.owner is not object: %#v", meta.Metadata["owner"])
	}
	if got, ok := owner.String("name"); !ok || got != "Hermes" {
		t.Fatalf("metadata.owner.name = %q/%v, want Hermes/true", got, ok)
	}
}

func TestParseSkillMD_MetadataRejectsUnsupportedYAMLConstructs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "non-string map key",
			content: `---
name: bad-metadata
description: bad metadata
metadata:
  1: one
---
`,
			wantErr: "keys must be strings",
		},
		{
			name: "timestamp",
			content: `---
name: bad-metadata
description: bad metadata
metadata:
  released: 2026-04-29
---
`,
			wantErr: "unsupported metadata scalar type !!timestamp",
		},
		{
			name: "alias",
			content: `---
name: bad-metadata
description: bad metadata
metadata:
  base: &base
    label: shared
  duplicate: *base
---
`,
			wantErr: "aliases are not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skillDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tt.content), 0o644); err != nil {
				t.Fatalf("write SKILL.md: %v", err)
			}
			_, err := ParseSkillMD(skillDir)
			if err == nil {
				t.Fatal("ParseSkillMD returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

package hermes

import "testing"

func TestClassifyHermesSourceKnownSources(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantKind   HermesSourceKind
		manageable bool
	}{
		{
			name:       "bundled skills root bare path",
			input:      "NousResearch/hermes-agent/skills",
			wantKind:   HermesSourceBundled,
			manageable: false,
		},
		{
			name:       "bundled skill bare path",
			input:      "NousResearch/hermes-agent/skills/foo",
			wantKind:   HermesSourceBundled,
			manageable: false,
		},
		{
			name:       "bundled skill github tree url",
			input:      "https://github.com/NousResearch/hermes-agent/tree/main/skills/foo",
			wantKind:   HermesSourceBundled,
			manageable: false,
		},
		{
			name:       "official optional skill github tree url",
			input:      "https://github.com/NousResearch/hermes-agent/tree/main/optional-skills/foo",
			wantKind:   HermesSourceOfficialOptional,
			manageable: true,
		},
		{
			name:       "official optional root bare path",
			input:      "NousResearch/hermes-agent/optional-skills",
			wantKind:   HermesSourceOfficialOptional,
			manageable: true,
		},
		{
			name:       "official optional nested bare path",
			input:      "NousResearch/hermes-agent/optional-skills/research/gitnexus-explorer",
			wantKind:   HermesSourceOfficialOptional,
			manageable: true,
		},
		{
			name:       "canonical skills index",
			input:      "https://hermes-agent.nousresearch.com/docs/api/skills-index.json",
			wantKind:   HermesSourceIndex,
			manageable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyHermesSource(tt.input)
			if got.Kind != tt.wantKind {
				t.Fatalf("Kind = %q, want %q", got.Kind, tt.wantKind)
			}
			if got.Manageable != tt.manageable {
				t.Fatalf("Manageable = %v, want %v", got.Manageable, tt.manageable)
			}
			if got.Reason == "" {
				t.Fatal("Reason should be helpful and non-empty")
			}
		})
	}
}

func TestClassifyHermesSourceNormalizesInputs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantKind   HermesSourceKind
		manageable bool
	}{
		{
			name:       "trims whitespace slash and git suffix",
			input:      "  NousResearch/hermes-agent.git/skills/foo/  ",
			wantKind:   HermesSourceBundled,
			manageable: false,
		},
		{
			name:       "case-insensitive owner and repo",
			input:      "nousresearch/HERMES-agent/optional-skills/foo",
			wantKind:   HermesSourceOfficialOptional,
			manageable: true,
		},
		{
			name:       "github blob url arbitrary branch",
			input:      "https://github.com/NousResearch/hermes-agent/blob/feature-branch/skills/foo/README.md",
			wantKind:   HermesSourceBundled,
			manageable: false,
		},
		{
			name:       "github URL without scheme",
			input:      "github.com/NousResearch/hermes-agent/optional-skills/foo",
			wantKind:   HermesSourceOfficialOptional,
			manageable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyHermesSource(tt.input)
			if got.Kind != tt.wantKind {
				t.Fatalf("Kind = %q, want %q", got.Kind, tt.wantKind)
			}
			if got.Manageable != tt.manageable {
				t.Fatalf("Manageable = %v, want %v", got.Manageable, tt.manageable)
			}
		})
	}
}

func TestClassifyHermesSourceDoesNotClassifyNonGitHubURLsByPath(t *testing.T) {
	got := ClassifyHermesSource("https://example.com/NousResearch/hermes-agent/skills/foo")
	if got.Kind != HermesSourceUnknown {
		t.Fatalf("Kind = %q, want %q", got.Kind, HermesSourceUnknown)
	}
	if !got.Manageable {
		t.Fatal("arbitrary non-GitHub URLs should not be treated as bundled Hermes sources")
	}
}

func TestClassifyHermesSourceUnknownIsConservativelyManageable(t *testing.T) {
	got := ClassifyHermesSource("someone/else/skills/foo")
	if got.Kind != HermesSourceUnknown {
		t.Fatalf("Kind = %q, want %q", got.Kind, HermesSourceUnknown)
	}
	if !got.Manageable {
		t.Fatal("unknown non-bundled sources should be manageable")
	}
	if got.Reason == "" {
		t.Fatal("Reason should be helpful and non-empty")
	}
}

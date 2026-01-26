package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name  string
		input string
		min   float64
		max   float64
	}{
		{"Empty", "", 0, 0},
		{"Low Entropy (Repeated)", "aaaaaaaa", 0, 0.1},
		{"Low Entropy (Sequence)", "12345678", 2.9, 3.1}, // Log2(8) = 3
		{"High Entropy (Random)", "zbK-d8.3_9sj29s", 3.5, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEntropy(tt.input)
			if got < tt.min || got > tt.max {
				t.Errorf("CalculateEntropy(%q) = %v; want between %v and %v", tt.input, got, tt.min, tt.max)
			}
		})
	}
}

func TestCheckSafety(t *testing.T) {
	// Create a temporary directory for test skills
	tmpDir, err := os.MkdirTemp("", "skill-check-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a dummy SKILL.md
	skillMD := filepath.Join(tmpDir, "SKILL.md")
	if err := os.WriteFile(skillMD, []byte("---\nname: Test Skill\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file with a safe string
	safeFile := filepath.Join(tmpDir, "safe.py")
	if err := os.WriteFile(safeFile, []byte("print('Hello World')"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file with an AWS key
	unsafeFile := filepath.Join(tmpDir, "unsafe.py")
	awsKey := "AKIAIOSFODNN7EXAMPLE" // Example key
	if err := os.WriteFile(unsafeFile, []byte("key = '"+awsKey+"'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file with a generic low entropy secret (should be ignored)
	lowEntropyFile := filepath.Join(tmpDir, "low_entropy.py")
	if err := os.WriteFile(lowEntropyFile, []byte("password = '12345678'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run the check
	result, err := CheckSafety(tmpDir)
	if err != nil {
		t.Fatalf("CheckSafety failed: %v", err)
	}

	// Verify findings
	foundAWS := false
	foundLowEntropy := false

	for _, finding := range result.Findings {
		if finding.RuleID == "SECRET-AWS-KEY" {
			foundAWS = true
		}
		if finding.RuleID == "SECRET-GENERIC-TOKEN" && finding.File == "low_entropy.py" {
			foundLowEntropy = true
		}
	}

	if !foundAWS {
		t.Error("Expected to find AWS key, but didn't")
	}

	if foundLowEntropy {
		t.Error("Did not expect to find low entropy password, but did")
	}
}

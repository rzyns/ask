package skill

import (
	"strings"
	"testing"
)

func TestGenerateReport_Markdown(t *testing.T) {
	result := &CheckResult{
		SkillName: "Test Skill",
		Findings: []Finding{
			{
				RuleID:      "TEST-RULE",
				Severity:    SeverityCritical,
				Description: "Critical Issue",
				File:        "test.py",
				Line:        10,
				Match:       "bad_code",
			},
		},
	}

	report, err := GenerateReport(result, "md")
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if !strings.Contains(report, "# Security Report: Test Skill") {
		t.Error("Report missing title")
	}
	if !strings.Contains(report, "| 🔴 Critical | 1 |") {
		t.Error("Report missing summary count")
	}
	if !strings.Contains(report, "### 🔴 Critical Issue") {
		t.Error("Report missing finding description")
	}
	if !strings.Contains(report, "`test.py:10`") {
		t.Error("Report missing location")
	}
}

func TestGenerateReport_HTML(t *testing.T) {
	result := &CheckResult{
		SkillName: "Test Skill",
		Findings: []Finding{
			{
				RuleID:      "TEST-RULE",
				Severity:    SeverityWarning,
				Description: "Warning Issue",
				File:        "test.py",
				Line:        10,
				Match:       "warn_code",
			},
		},
	}

	report, err := GenerateReport(result, "html")
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if !strings.Contains(report, "<title>Security Report: Test Skill</title>") {
		t.Error("HTML missing title")
	}
	if !strings.Contains(report, "<span class=\"stat-value text-critical\">0</span>") {
		t.Error("HTML missing critical count (should be 0)")
	}
	if !strings.Contains(report, "<span class=\"stat-value text-warning\">1</span>") {
		t.Error("HTML missing warning count (should be 1)")
	}
	if !strings.Contains(report, "Warning Issue") {
		t.Error("HTML missing finding description")
	}
	if !strings.Contains(report, "bg-warning") {
		t.Error("HTML missing severity class")
	}
}

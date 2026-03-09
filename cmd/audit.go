package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// AuditReport represents the full audit report
type AuditReport struct {
	GeneratedAt   time.Time         `json:"generated_at"`
	Version       string            `json:"version"`
	TotalSkills   int               `json:"total_skills"`
	TotalFindings int               `json:"total_findings"`
	CriticalCount int               `json:"critical_count"`
	WarningCount  int               `json:"warning_count"`
	InfoCount     int               `json:"info_count"`
	Skills        []AuditSkillEntry `json:"skills"`
}

// AuditSkillEntry represents one skill in the audit
type AuditSkillEntry struct {
	Name        string          `json:"name"`
	URL         string          `json:"url,omitempty"`
	Version     string          `json:"version,omitempty"`
	Commit      string          `json:"commit,omitempty"`
	InstalledAt string          `json:"installed_at,omitempty"`
	Source      string          `json:"source,omitempty"`
	Findings    []skill.Finding `json:"findings,omitempty"`
	Status      string          `json:"status"` // "clean", "warning", "critical"
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Generate a security audit report of installed skills",
	Long: `Scan all installed skills and generate a comprehensive security audit report.

The report includes:
  - Complete inventory of installed skills with versions
  - Security scan results for each skill
  - Summary statistics of findings by severity
  - Source and provenance information

Output formats: console (default), json, html, markdown`,
	Example: `  # Console audit
  ask audit

  # Generate JSON report
  ask audit --format json --output audit-report.json

  # Generate HTML report
  ask audit --format html --output audit-report.html

  # Audit global skills
  ask audit --global`,
	Run: runAudit,
}

func runAudit(cmd *cobra.Command, _ []string) {
	global, _ := cmd.Flags().GetBool("global")
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")

	// Load lock file for version info
	lockFile, _ := config.LoadLockFileByScope(global)

	// Find skills directory
	skillsDir := config.GetSkillsDirByScope(global)

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No skills directory found at %s\n", skillsDir)
			return
		}
		fmt.Printf("Error reading skills directory: %v\n", err)
		os.Exit(1)
	}

	report := AuditReport{
		GeneratedAt: time.Now(),
		Version:     Version,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		auditEntry := AuditSkillEntry{
			Name:   entry.Name(),
			Status: "clean",
		}

		// Get lock info if available
		if lockFile != nil {
			lockEntry := lockFile.GetEntry(entry.Name())
			if lockEntry != nil {
				auditEntry.URL = lockEntry.URL
				auditEntry.Version = lockEntry.Version
				auditEntry.Commit = lockEntry.Commit
				auditEntry.Source = lockEntry.Source
				if !lockEntry.InstalledAt.IsZero() {
					auditEntry.InstalledAt = lockEntry.InstalledAt.Format(time.RFC3339)
				}
			}
		}

		// Run security check
		if skill.FindSkillMD(skillPath) {
			result, err := skill.CheckSafety(skillPath)
			if err == nil {
				auditEntry.Findings = result.Findings
				for _, f := range result.Findings {
					switch f.Severity {
					case skill.SeverityCritical:
						report.CriticalCount++
						auditEntry.Status = "critical"
					case skill.SeverityWarning:
						report.WarningCount++
						if auditEntry.Status != "critical" {
							auditEntry.Status = "warning"
						}
					case skill.SeverityInfo:
						report.InfoCount++
					}
				}
				report.TotalFindings += len(result.Findings)
			}
		}

		report.Skills = append(report.Skills, auditEntry)
		report.TotalSkills++
	}

	// Output
	switch format {
	case "json":
		data, _ := json.MarshalIndent(report, "", "  ")
		content := string(data)
		if output != "" {
			if err := os.WriteFile(output, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Audit report saved to %s\n", output)
		} else {
			fmt.Println(content)
		}
	case "html", "markdown", "md":
		// Build markdown content
		var sb strings.Builder
		sb.WriteString("# Security Audit Report\n\n")
		sb.WriteString(fmt.Sprintf("Generated: %s | ASK v%s\n\n", report.GeneratedAt.Format(time.RFC3339), report.Version))
		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total skills**: %d\n", report.TotalSkills))
		sb.WriteString(fmt.Sprintf("- **Critical**: %d\n", report.CriticalCount))
		sb.WriteString(fmt.Sprintf("- **Warnings**: %d\n", report.WarningCount))
		sb.WriteString(fmt.Sprintf("- **Info**: %d\n\n", report.InfoCount))
		sb.WriteString("## Skills\n\n")
		for _, s := range report.Skills {
			status := "✓"
			switch s.Status {
			case "critical":
				status = "✗"
			case "warning":
				status = "!"
			}
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", status, s.Name))
			if s.URL != "" {
				sb.WriteString(fmt.Sprintf("- URL: %s\n", s.URL))
			}
			if s.Version != "" {
				sb.WriteString(fmt.Sprintf("- Version: %s\n", s.Version))
			}
			if len(s.Findings) > 0 {
				sb.WriteString(fmt.Sprintf("- Findings: %d\n\n", len(s.Findings)))
				for _, f := range s.Findings {
					sb.WriteString(fmt.Sprintf("  - [%s] %s (%s:%d)\n", f.Severity, f.Description, f.File, f.Line))
				}
			} else {
				sb.WriteString("- No issues found\n")
			}
			sb.WriteString("\n")
		}

		content := sb.String()
		if output != "" {
			if err := os.WriteFile(output, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Audit report saved to %s\n", output)
		} else {
			fmt.Println(content)
		}
	default:
		// Console output
		fmt.Println("Security Audit Report")
		fmt.Println("=====================")
		fmt.Printf("Skills scanned: %d\n", report.TotalSkills)
		fmt.Printf("Findings: %d Critical, %d Warning, %d Info\n\n",
			report.CriticalCount, report.WarningCount, report.InfoCount)

		for _, s := range report.Skills {
			switch s.Status {
			case "critical":
				fmt.Printf("  %s %s", color.RedString("✗"), s.Name)
			case "warning":
				fmt.Printf("  %s %s", color.YellowString("!"), s.Name)
			default:
				fmt.Printf("  %s %s", color.GreenString("✓"), s.Name)
			}
			if s.Version != "" {
				fmt.Printf(" (%s)", s.Version)
			}
			fmt.Println()

			for _, f := range s.Findings {
				switch f.Severity {
				case skill.SeverityCritical:
					fmt.Printf("    %s %s\n", color.RedString("[CRITICAL]"), f.Description)
				case skill.SeverityWarning:
					fmt.Printf("    %s %s\n", color.YellowString("[WARNING]"), f.Description)
				}
			}
		}

		if report.CriticalCount > 0 {
			fmt.Printf("\n%s Critical issues found. Review and fix before deployment.\n", color.RedString("!"))
		} else {
			fmt.Printf("\n%s All skills passed security audit.\n", color.GreenString("✓"))
		}
	}

	if report.CriticalCount > 0 {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().String("format", "console", "Output format: console, json, html, markdown")
	auditCmd.Flags().StringP("output", "o", "", "Write report to file")
}

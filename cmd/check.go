package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/skill"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [skill-path]",
	Short: "Check a skill for security issues",
	Long: `Analyze a skill directory for potential security risks, 
including hardcoded secrets, dangerous commands, and network activity.

If skill-path is not provided, the current directory is checked.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runCheck,
}

var reportFile string

func init() {
	checkCmd.Flags().StringVar(&reportFile, "report", "", "Save report to file (supports .md, .html)")
}

func runCheck(cmd *cobra.Command, args []string) {
	skillPath := "."
	if len(args) > 0 {
		skillPath = args[0]
	}

	absPath, err := filepath.Abs(skillPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	if !skill.FindSkillMD(absPath) {
		fmt.Printf("Error: No SKILL.md found in %s. Is this a valid skill directory?\n", absPath)
		os.Exit(1)
	}

	fmt.Printf("Checking skill at %s...\n", absPath)
	result, err := skill.CheckSafety(absPath)
	if err != nil {
		fmt.Printf("Error checking skill: %v\n", err)
		os.Exit(1)
	}

	if reportFile != "" {
		handleReport(result, reportFile)
	} else {
		printReport(result)
	}
}

func handleReport(result *skill.CheckResult, filename string) {
	ext := filepath.Ext(filename)
	format := "md"
	if ext == ".html" || ext == ".htm" {
		format = "html"
	}

	content, err := skill.GenerateReport(result, format)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing report to %s: %v\n", filename, err)
		os.Exit(1)
	}

	fmt.Printf("\n%s Security report saved to %s\n", color.GreenString("✓"), filename)

	// Print a brief summary even when saving to file
	criticals := 0
	for _, f := range result.Findings {
		if f.Severity == skill.SeverityCritical {
			criticals++
		}
	}
	if criticals > 0 {
		fmt.Printf("%s Found %d critical issues. Please review the report.\n", color.RedString("!"), criticals)
		os.Exit(1)
	}
}

func printReport(result *skill.CheckResult) {
	fmt.Printf("\nSecurity Report for: %s\n", color.CyanString(result.SkillName))
	fmt.Println("----------------------------------------")

	if len(result.Findings) == 0 {
		color.Green("✓ No issues found. Skill appears safe.\n")
		return
	}

	criticals := 0
	warnings := 0
	infos := 0

	for _, finding := range result.Findings {
		switch finding.Severity {
		case skill.SeverityCritical:
			criticals++
			printFinding(color.RedString("[CRITICAL]"), finding)
		case skill.SeverityWarning:
			warnings++
			printFinding(color.YellowString("[WARNING] "), finding)
		case skill.SeverityInfo:
			infos++
			printFinding(color.BlueString("[INFO]    "), finding)
		}
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("Summary: %d Critical, %d Warning, %d Info\n", criticals, warnings, infos)

	if criticals > 0 {
		os.Exit(1) // Exit with error if critical issues found
	}
}

func printFinding(prefix string, finding skill.Finding) {
	fmt.Printf("%s %s\n", prefix, finding.Description)
	fmt.Printf("  File: %s:%d\n", finding.File, finding.Line)
	fmt.Printf("  Match: %s\n\n", finding.Match)
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [skill-path]",
	Short: "Check a skill for security issues",
	Long: `Analyze a skill directory for potential security risks, 
including hardcoded secrets, dangerous commands, and network activity.

If skill-path is not provided, the current directory is checked.
If the current directory is not a skill, all installed skills across all agents are checked.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runCheck,
}

var outputFile string

func init() {
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write report to file (supports .md, .html/.htm, .json)")
}

func runCheck(_ *cobra.Command, args []string) {
	// Case 1: Specific path provided
	if len(args) > 0 {
		skillPath := args[0]
		absPath, err := filepath.Abs(skillPath)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}

		if !skill.FindSkillMD(absPath) {
			fmt.Printf("Error: No SKILL.md found in %s. Is this a valid skill directory?\n", absPath)
			os.Exit(1)
		}

		result, err := checkSingleSkill(absPath)
		if err != nil {
			fmt.Printf("Error checking skill: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			handleReport(result, outputFile)
		} else {
			printReport(result)
		}
		if hasCriticalIssues(result) {
			os.Exit(1)
		}
		return
	}

	// Case 2: No path provided. Check current directory first.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	if skill.FindSkillMD(cwd) {
		fmt.Printf("Checking skill in current directory...\n")
		result, err := checkSingleSkill(cwd)
		if err != nil {
			fmt.Printf("Error checking skill: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			handleReport(result, outputFile)
		} else {
			printReport(result)
		}
		if hasCriticalIssues(result) {
			os.Exit(1)
		}
		return
	}

	// Case 3: Current directory is not a skill. Scan all *PROJECT* skills.
	fmt.Println("Current directory is not a skill. Scanning project skills...")
	fmt.Println()

	// Only scan project-level directories, not global ones
	dirs := []string{config.DefaultSkillsDir}
	for _, agentConfig := range config.SupportedAgents {
		dirs = append(dirs, agentConfig.ProjectDir)
	}

	foundSkills := 0
	issuesFound := false

	for _, relDir := range dirs {
		// Join with CWD to get absolute path
		dir := filepath.Join(cwd, relDir)

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				skillPath := filepath.Join(dir, entry.Name())
				if skill.FindSkillMD(skillPath) {
					foundSkills++
					displayPath := formatPathForDisplay(skillPath)
					fmt.Printf("Checking %s...\n", displayPath)

					result, err := checkSingleSkill(skillPath)
					if err != nil {
						fmt.Printf("  Error checking skill: %v\n", err)
						continue
					}

					// For bulk check, we primarily only show issues or summary
					printCompactReport(result)
					if hasCriticalIssues(result) {
						issuesFound = true
					}
				}
			}
		}
	}

	if foundSkills == 0 {
		fmt.Println("No skills found to check.")
	} else {
		fmt.Printf("\nScanned %d skills.\n", foundSkills)
	}

	if issuesFound {
		fmt.Println(color.RedString("\nCritical issues found in one or more skills."))
		os.Exit(1)
	}
}

func formatPathForDisplay(path string) string {
	cwd, err := os.Getwd()
	if err == nil {
		rel, err := filepath.Rel(cwd, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}

	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	return path
}

func checkSingleSkill(absPath string) (*skill.CheckResult, error) {
	return skill.CheckSafety(absPath)
}

func hasCriticalIssues(result *skill.CheckResult) bool {
	for _, f := range result.Findings {
		if f.Severity == skill.SeverityCritical {
			return true
		}
	}
	return false
}

func handleReport(result *skill.CheckResult, filename string) {
	ext := filepath.Ext(filename)
	format := "md"
	switch ext {
	case ".html", ".htm":
		format = "html"
	case ".json":
		format = "json"
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

	if hasCriticalIssues(result) {
		fmt.Printf("%s Found critical issues. Please review the report.\n", color.RedString("!"))
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
}

func printCompactReport(result *skill.CheckResult) {
	if len(result.Findings) == 0 {
		fmt.Printf("  %s %s: No issues\n", color.GreenString("✓"), result.SkillName)
		return
	}

	criticals := 0
	warnings := 0
	for _, f := range result.Findings {
		switch f.Severity {
		case skill.SeverityCritical:
			criticals++
		case skill.SeverityWarning:
			warnings++
		}
	}

	if criticals > 0 {
		fmt.Printf("  %s %s: %d Critical, %d Warnings\n", color.RedString("!"), result.SkillName, criticals, warnings)
		for _, f := range result.Findings {
			if f.Severity == skill.SeverityCritical {
				fmt.Printf("    - %s: %s\n", f.RuleID, f.Description)
			}
		}
	} else if warnings > 0 {
		fmt.Printf("  %s %s: %d Warnings\n", color.YellowString("!"), result.SkillName, warnings)
	} else {
		// Only info
		fmt.Printf("  %s %s: OK (Info only)\n", color.GreenString("✓"), result.SkillName)
	}
}

func printFinding(prefix string, finding skill.Finding) {
	fmt.Printf("%s %s\n", prefix, finding.Description)
	fmt.Printf("  File: %s:%d\n", finding.File, finding.Line)
	fmt.Printf("  Match: %s\n\n", finding.Match)
}

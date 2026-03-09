package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/skill"
)

var testCmd = &cobra.Command{
	Use:   "test [skill-path]",
	Short: "Run validation checks on a skill",
	Long: `Run a comprehensive suite of validation checks on a skill.

Checks include:
  - SKILL.md exists and has valid format
  - Required metadata fields are present
  - Version follows semantic versioning
  - Security scan passes
  - README.md exists
  - At least one prompt or content file exists`,
	Example: `  # Test skill in current directory
  ask skill test

  # Test a specific skill
  ask skill test ./my-skill`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTest,
}

var semverRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)

func runTest(_ *cobra.Command, args []string) {
	var targetPath string
	var err error

	if len(args) > 0 {
		targetPath, err = filepath.Abs(args[0])
	} else {
		targetPath, err = os.Getwd()
	}
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Testing skill at %s...\n\n", targetPath)

	passed := 0
	warned := 0
	failed := 0

	// Test 1: SKILL.md exists
	fmt.Print("  SKILL.md exists............. ")
	if !skill.FindSkillMD(targetPath) {
		color.Red("FAIL")
		failed++
		fmt.Println("\n  Cannot continue without SKILL.md")
		os.Exit(1)
	}
	color.Green("PASS")
	passed++

	// Test 2: SKILL.md format valid
	fmt.Print("  SKILL.md format valid....... ")
	meta, err := skill.ParseSkillMD(targetPath)
	if err != nil || meta == nil {
		color.Red("FAIL")
		if err != nil {
			fmt.Printf("    %v\n", err)
		}
		failed++
	} else {
		color.Green("PASS")
		passed++
	}

	// Test 3: Name present
	fmt.Print("  Name present................ ")
	if meta != nil && meta.Name != "" {
		color.Green("PASS")
		passed++
	} else {
		color.Red("FAIL")
		failed++
	}

	// Test 4: Description present
	fmt.Print("  Description present......... ")
	if meta != nil && meta.Description != "" {
		color.Green("PASS")
		passed++
	} else {
		color.Red("FAIL")
		failed++
	}

	// Test 5: Version follows semver
	fmt.Print("  Version follows semver...... ")
	if meta != nil && meta.Version != "" {
		if semverRegex.MatchString(meta.Version) {
			color.Green("PASS")
			passed++
		} else {
			color.Yellow("WARN (invalid format: %s)", meta.Version)
			warned++
		}
	} else {
		color.Yellow("WARN (no version)")
		warned++
	}

	// Test 6: Security scan
	fmt.Print("  Security scan passed........ ")
	result, err := skill.CheckSafety(targetPath)
	if err != nil {
		color.Red("FAIL")
		fmt.Printf("    %v\n", err)
		failed++
	} else {
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
			color.Red("FAIL (%d critical, %d warning)", criticals, warnings)
			failed++
		} else if warnings > 0 {
			color.Yellow("WARN (%d warnings)", warnings)
			warned++
		} else {
			color.Green("PASS")
			passed++
		}
	}

	// Test 7: README exists
	fmt.Print("  README.md exists............ ")
	if _, err := os.Stat(filepath.Join(targetPath, "README.md")); err == nil {
		color.Green("PASS")
		passed++
	} else {
		color.Yellow("WARN (optional)")
		warned++
	}

	// Test 8: Content files exist
	fmt.Print("  Content files exist......... ")
	hasContent := false
	for _, pattern := range []string{"*.md", "prompts/*.md", "prompts/*.txt", "*.py", "*.js", "*.sh"} {
		matches, _ := filepath.Glob(filepath.Join(targetPath, pattern))
		for _, m := range matches {
			base := filepath.Base(m)
			if base != "SKILL.md" && base != "README.md" {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if hasContent {
		color.Green("PASS")
		passed++
	} else {
		color.Yellow("WARN (no prompt/content files found)")
		warned++
	}

	// Summary
	fmt.Println()
	fmt.Printf("Results: %d passed, %d warnings, %d failed\n", passed, warned, failed)

	if failed == 0 && warned == 0 {
		color.Green("\n✓ All checks passed! Ready to publish.\n")
	} else if failed == 0 {
		color.Yellow("\n! Passed with warnings. Consider fixing before publishing.\n")
	} else {
		color.Red("\n✗ Some checks failed. Fix issues before publishing.\n")
		os.Exit(1)
	}
}

func init() {
	skillCmd.AddCommand(testCmd)
}

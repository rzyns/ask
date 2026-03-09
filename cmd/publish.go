package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/skill"
)

var publishCmd = &cobra.Command{
	Use:   "publish [skill-path]",
	Short: "Validate and prepare a skill for publishing",
	Long: `Validate a skill and prepare it for publishing to the ASK registry.

This command runs comprehensive checks including:
  - SKILL.md format validation
  - Security scanning
  - Required files verification
  - Git repository status check

After validation passes, it provides instructions for submitting
the skill to the registry.`,
	Example: `  # Publish skill in current directory
  ask skill publish

  # Publish skill at a specific path
  ask skill publish ./my-skill`,
	Args: cobra.MaximumNArgs(1),
	Run:  runPublish,
}

func runPublish(_ *cobra.Command, args []string) {
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

	fmt.Printf("Preparing to publish skill at %s...\n\n", targetPath)

	// Step 1: Check SKILL.md exists
	fmt.Print("  Checking SKILL.md... ")
	if !skill.FindSkillMD(targetPath) {
		color.Red("FAIL")
		fmt.Println("    No SKILL.md found. Create one with 'ask skill create'.")
		os.Exit(1)
	}
	color.Green("OK")

	// Step 2: Validate SKILL.md format
	fmt.Print("  Validating SKILL.md format... ")
	meta, err := skill.ParseSkillMD(targetPath)
	if err != nil {
		color.Red("FAIL")
		fmt.Printf("    %v\n", err)
		os.Exit(1)
	}

	validationErrors := validateSkillMeta(meta)
	if len(validationErrors) > 0 {
		color.Red("FAIL")
		for _, e := range validationErrors {
			fmt.Printf("    - %s\n", e)
		}
		os.Exit(1)
	}
	color.Green("OK")

	// Step 3: Security scan
	fmt.Print("  Running security scan... ")
	result, err := skill.CheckSafety(targetPath)
	if err != nil {
		color.Red("FAIL")
		fmt.Printf("    %v\n", err)
		os.Exit(1)
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
		color.Red("FAIL")
		fmt.Printf("    %d critical issues found. Fix them before publishing.\n", criticals)
		for _, f := range result.Findings {
			if f.Severity == skill.SeverityCritical {
				fmt.Printf("    - %s: %s (%s:%d)\n", f.RuleID, f.Description, f.File, f.Line)
			}
		}
		os.Exit(1)
	}

	if warnings > 0 {
		color.Yellow("WARN (%d warnings)", warnings)
	} else {
		color.Green("OK")
	}

	// Step 4: Check for README
	fmt.Print("  Checking README.md... ")
	readmePath := filepath.Join(targetPath, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		color.Yellow("MISSING (recommended)")
	} else {
		color.Green("OK")
	}

	// Step 5: Check git status
	fmt.Print("  Checking git status... ")
	gitCmd := exec.Command("git", "-C", targetPath, "status", "--porcelain")
	output, err := gitCmd.Output()
	if err != nil {
		color.Yellow("SKIP (not a git repo)")
	} else if len(output) > 0 {
		color.Yellow("WARN (uncommitted changes)")
	} else {
		color.Green("OK")
	}

	// Summary
	fmt.Println()
	if criticals == 0 {
		color.Green("✓ Skill '%s' is ready for publishing!\n", meta.Name)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Push your skill to a public GitHub repository")
		fmt.Println("  2. Submit it to the ASK registry:")
		fmt.Println("     https://github.com/yeasy/awesome-agent-skills/issues/new")
		fmt.Println()
		fmt.Printf("  Install command: ask install <your-github-username>/<repo-name>\n")
	}
}

func validateSkillMeta(meta *skill.Meta) []string {
	var errors []string
	if meta == nil {
		return []string{"failed to parse SKILL.md metadata"}
	}
	if meta.Name == "" {
		errors = append(errors, "name is required in SKILL.md frontmatter")
	}
	if meta.Description == "" {
		errors = append(errors, "description is required in SKILL.md frontmatter")
	}
	return errors
}

func init() {
	skillCmd.AddCommand(publishCmd)
}

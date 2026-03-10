package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/git"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/skill"
)

var scoreCmd = &cobra.Command{
	Use:   "score <path-or-url>",
	Short: "Compute a trust score for a skill",
	Long: `Compute a comprehensive trust score for a skill, evaluating:
  - Security: static analysis for secrets, malware, and dangerous commands
  - Quality: SKILL.md metadata, README, prompts structure
  - Publisher: GitHub account reputation, stars, organization status
  - Transparency: data exfiltration patterns and obfuscated code

The score ranges from 0-100 with grades A/B/C/D/F.
Similar to Snyk or Socket.dev for the agent skill ecosystem.`,
	Example: `  # Score a local skill directory
  ask score ./my-skill

  # Score with JSON output
  ask score ./my-skill --json

  # Score a remote skill (cloned to temp dir)
  ask score anthropics/browser-use`,
	Args: cobra.ExactArgs(1),
	Run:  runScore,
}

func runScore(cmd *cobra.Command, args []string) {
	target := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Resolve target path
	skillPath := target
	var publisher *skill.PublisherInfo

	// Check if target is a local directory
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		// Try as a GitHub reference — clone to temp
		fmt.Printf("Resolving %s...\n", target)
		tmpDir, cloneErr := cloneForScore(target)
		if cloneErr != nil {
			fmt.Printf("Error: cannot resolve '%s' as local path or GitHub repo: %v\n", target, cloneErr)
			os.Exit(1)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()
		skillPath = tmpDir

		// Fetch publisher info from GitHub
		publisher = fetchPublisherInfo(target)
	}

	// Run scoring
	result, err := skill.ScoreSkill(skillPath, publisher)
	if err != nil {
		fmt.Printf("Error computing score: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if encErr := encoder.Encode(result); encErr != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", encErr)
		}
		return
	}

	// Pretty print
	printScoreResult(result)
}

func printScoreResult(result *skill.ScoreResult) {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	fmt.Println()
	_, _ = bold.Printf("  Trust Score: %s %s\n", result.SkillName, formatGrade(result))
	fmt.Println()

	// Score bar
	barLen := 40
	filled := int(result.TotalScore / 100 * float64(barLen))
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
	fmt.Printf("  [%s] %.1f/100\n\n", bar, result.TotalScore)

	// Category breakdown
	for _, cat := range result.Categories {
		icon := "✓"
		printer := green
		if cat.Score < 70 {
			icon = "⚠"
			printer = yellow
		}
		if cat.Score < 50 {
			icon = "✗"
			printer = red
		}

		_, _ = printer.Printf("  %s %-14s", icon, cat.Name)
		fmt.Printf("  %5.1f / 100  (weight: %.0f%%)\n", cat.Score, cat.Weight*100)

		if cat.Details != "" {
			fmt.Printf("    %s\n", cat.Details)
		}
		for _, d := range cat.Deducts {
			_, _ = red.Printf("    -%.*f  ", 0, d.Points)
			fmt.Printf("%s\n", d.Reason)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("  %s\n\n", result.Summary)
}

func formatGrade(result *skill.ScoreResult) string {
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	grade := string(result.Grade)
	switch result.Grade {
	case skill.GradeA, skill.GradeB:
		return green.Sprint(grade)
	case skill.GradeC:
		return yellow.Sprint(grade)
	default:
		return red.Sprint(grade)
	}
}

func cloneForScore(target string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ask-score-*")
	if err != nil {
		return "", err
	}

	url := target
	if !strings.HasPrefix(target, "http") {
		url = "https://github.com/" + target
	}

	cloneErr := git.Clone(url, tmpDir)
	if cloneErr != nil {
		_ = os.RemoveAll(tmpDir)
		return "", cloneErr
	}

	return tmpDir, nil
}

func fetchPublisherInfo(target string) *skill.PublisherInfo {
	// Extract owner from target like "owner/repo" or "owner/repo/path"
	parts := strings.Split(strings.TrimPrefix(target, "https://github.com/"), "/")
	if len(parts) < 2 {
		return nil
	}
	owner := parts[0]
	repo := parts[1]

	info := &skill.PublisherInfo{
		Owner: owner,
	}

	// Fetch repo metadata via GitHub API
	client := github.NewClient()
	repoInfo, err := client.GetRepoInfo(owner, repo)
	if err == nil {
		info.RepoStars = repoInfo.Stars
		info.IsOrg = repoInfo.IsOrg
		info.HasLicense = repoInfo.HasLicense
		info.AccountAge = repoInfo.OwnerAge
		info.RepoForks = repoInfo.Forks
	}

	return info
}

func init() {
	skillCmd.AddCommand(scoreCmd)
	scoreCmd.Flags().Bool("json", false, "Output score as JSON")
}

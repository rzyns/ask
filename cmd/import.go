package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/hermes"
)

var importCmd = &cobra.Command{
	Use:   "import [skill-name...]",
	Short: "Import installed native skills into ASK tracking",
	Long:  "Import existing agent-native skills into ASK tracking without moving files.",
	Run:   runImport,
}

func runImport(cmd *cobra.Command, args []string) {
	global, _ := cmd.Flags().GetBool("global")
	agents, _ := cmd.Flags().GetStringSlice("agent")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	importAll, _ := cmd.Flags().GetBool("all")

	if !hasHermesAgent(agents) {
		fmt.Fprintln(os.Stderr, "Error: import currently supports only --agent hermes")
		os.Exit(1)
	}
	if !importAll && len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: specify one or more skills or use --all")
		os.Exit(1)
	}

	dir, err := config.GetAgentSkillsDir(config.AgentHermes, global)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving Hermes skills directory: %v\n", err)
		os.Exit(1)
	}
	lockFile, lockErr := config.LoadLockFileByScope(global)
	if lockErr != nil || lockFile == nil {
		lockFile = &config.LockFile{Version: 1, Skills: []config.LockEntry{}}
	}
	installed, err := hermes.ScanInstalledSkills(dir, hermes.InstalledScanOptions{LockFile: lockFile})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning Hermes skills: %v\n", err)
		os.Exit(1)
	}
	plan := hermes.PlanImport(installed, lockFile, args, importAll)

	fmt.Printf("Hermes Skills Import (%s):\n\n", dir)
	if len(plan) == 0 {
		fmt.Println("  (none)")
		return
	}
	nameW, classW, actionW := 4, 14, 6
	for _, candidate := range plan {
		if len(candidate.Skill.Name) > nameW {
			nameW = len(candidate.Skill.Name)
		}
		if len(candidate.Classification) > classW {
			classW = len(candidate.Classification)
		}
		if len(candidate.Action) > actionW {
			actionW = len(candidate.Action)
		}
	}
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  PATH", nameW, "NAME", classW, "CLASSIFICATION", actionW, "ACTION")
	fmt.Println(header)
	fmt.Println("  " + strings.Repeat("-", len(header)-2))
	for _, candidate := range plan {
		fmt.Printf("  %-*s  %-*s  %-*s  %s\n", nameW, candidate.Skill.Name, classW, candidate.Classification, actionW, candidate.Action, candidate.Skill.Path)
	}
	if dryRun {
		return
	}

	changed := false
	for _, candidate := range plan {
		if candidate.Action != "import as local" {
			continue
		}
		entry, err := hermes.LockEntryForImportedSkill(candidate.Skill)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing %s: %v\n", candidate.Skill.Name, err)
			os.Exit(1)
		}
		lockFile.AddEntry(entry)
		changed = true
	}
	if changed {
		updatedCfg, err := config.LoadConfigByScope(global)
		if err == nil && updatedCfg != nil {
			for _, candidate := range plan {
				if candidate.Action == "import as local" {
					updatedCfg.AddSkillInfo(config.SkillInfo{Name: candidate.Skill.Name, Description: candidate.Skill.Description})
				}
			}
			if err := updatedCfg.SaveByScope(global); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}
		}
		if err := lockFile.SaveByScope(global); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving lock file: %v\n", err)
			os.Exit(1)
		}
	}
}

func hasHermesAgent(agents []string) bool {
	for _, agentName := range agents {
		agentType, ok := config.ResolveAgentType(agentName)
		if ok && agentType == config.AgentHermes {
			return true
		}
	}
	return false
}

func init() {
	skillCmd.AddCommand(importCmd)
	importCmd.Flags().StringSliceP("agent", "a", []string{}, "target agent(s) for import")
	importCmd.Flags().Bool("dry-run", false, "show what would be imported without writing")
	importCmd.Flags().Bool("all", false, "import all eligible skills")
	_ = importCmd.RegisterFlagCompletionFunc("agent", completeAgentNames)
}

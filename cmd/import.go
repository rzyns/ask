package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/hermes"
)

var importCmd = &cobra.Command{
	Use:   "import [skill...]",
	Short: "Import existing agent skills into ask.lock",
	RunE:  runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	agent, _ := cmd.Flags().GetString("agent")
	if strings.TrimSpace(agent) != "hermes" {
		return fmt.Errorf("skill import MVP requires exactly --agent hermes; mixed or omitted agents are not supported")
	}
	all, _ := cmd.Flags().GetBool("all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	global, _ := cmd.Flags().GetBool("global")
	if dryRun && !all && len(args) == 0 {
		all = true
	}
	if all && len(args) > 0 {
		return fmt.Errorf("use either --all or specific skill name(s), not both")
	}
	if !all && len(args) == 0 {
		return fmt.Errorf("specify skill name(s) or --all")
	}

	dir, err := config.GetAgentSkillsDir(config.AgentHermes, global)
	if err != nil {
		return err
	}
	lockFile, err := config.LoadLockFileByScope(global)
	if err != nil {
		return err
	}
	result, err := hermes.PlanImport(hermes.ImportOptions{SkillsDir: dir, LockFile: lockFile, All: all, Names: args})
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if dryRun {
		fmt.Fprintln(out, "Hermes skill import dry-run:")
	} else {
		fmt.Fprintln(out, "Hermes skill import:")
	}
	for _, candidate := range result.Importable {
		fmt.Fprintf(out, "  import %s -> %s\n", candidate.Entry.Name, candidate.Entry.TargetPath)
	}
	for _, skipped := range result.SkippedManaged {
		fmt.Fprintf(out, "  skip %s (already in ask.lock)\n", skipped.Name)
	}
	for _, skipped := range result.SkippedBundled {
		fmt.Fprintf(out, "  skip %s (bundled Hermes skill)\n", skipped.Name)
	}
	if len(result.UnmatchedNames) > 0 {
		return fmt.Errorf("installed Hermes skill(s) not found: %s", strings.Join(result.UnmatchedNames, ", "))
	}
	if len(result.Importable) == 0 {
		fmt.Fprintln(out, "  no importable Hermes skills found")
	}
	if dryRun {
		return nil
	}
	hermes.ApplyImport(lockFile, result)
	if err := lockFile.SaveByScope(global); err != nil {
		return err
	}
	return nil
}

func init() {
	skillCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("agent", "a", "", "agent to import from (currently only hermes)")
	importCmd.Flags().Bool("all", false, "import all eligible skills")
	importCmd.Flags().Bool("dry-run", false, "show what would be imported without writing ask.lock")
	importCmd.Flags().Bool("global", false, "import from global Hermes skills directory")
	_ = importCmd.RegisterFlagCompletionFunc("agent", completeAgentNames)
}

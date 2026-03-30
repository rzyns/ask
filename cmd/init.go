package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/installer"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ASK project",
	Long: `Initialize a new Agent Skills Kit project with interactive setup.

This walks you through selecting your AI agents and optionally
installing a starter skill pack. Use --yes to skip prompts.`,
	Example: `  # Interactive setup
  ask init

  # Non-interactive with defaults
  ask init --yes`,
	Run: runInteractiveInit,
}

func runInteractiveInit(cmd *cobra.Command, _ []string) {
	if _, err := os.Stat("ask.yaml"); err == nil {
		fmt.Println("ask.yaml already exists in this directory.")
		return
	}

	yes, _ := cmd.Flags().GetBool("yes")

	if yes {
		runNonInteractiveInit()
		return
	}

	// Welcome banner
	fmt.Println()
	fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("  Welcome to ASK — the package manager for agent skills!"))
	fmt.Println()

	// --- Step 1: Detect and select agents ---
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	detected := config.DetectExistingToolDirs(cwd)
	detectedSet := make(map[string]bool)
	for _, d := range detected {
		detectedSet[d.Name] = true
	}

	// Build agent options sorted by: detected first, then alphabetical
	type agentOption struct {
		key      string
		display  string
		detected bool
	}
	var agentOpts []agentOption
	for _, name := range config.GetSupportedAgentNames() {
		ac := config.SupportedAgents[config.AgentType(name)]
		label := ac.Name
		if detectedSet[name] {
			label += color.New(color.FgGreen).Sprint(" (detected)")
		}
		agentOpts = append(agentOpts, agentOption{
			key:      name,
			display:  label,
			detected: detectedSet[name],
		})
	}
	sort.Slice(agentOpts, func(i, j int) bool {
		if agentOpts[i].detected != agentOpts[j].detected {
			return agentOpts[i].detected
		}
		return agentOpts[i].key < agentOpts[j].key
	})

	// Pre-select detected agents
	var selectedAgents []string
	for _, opt := range agentOpts {
		if opt.detected {
			selectedAgents = append(selectedAgents, opt.key)
		}
	}

	huhOpts := make([]huh.Option[string], 0, len(agentOpts))
	for _, opt := range agentOpts {
		huhOpts = append(huhOpts, huh.NewOption(opt.display, opt.key))
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which agents do you use?").
				Description("Space to toggle, Enter to confirm").
				Options(huhOpts...).
				Value(&selectedAgents),
		),
	).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// --- Step 2: Create config and directories ---
	skillsDir := config.DefaultSkillsDir
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating skills directory: %v\n", err)
		os.Exit(1)
	}

	if err := config.CreateConfigWithAgents(selectedAgents); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating ask.yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	color.Green("✓ Initialized ASK project")
	fmt.Println("  Created: ask.yaml")
	fmt.Printf("  Created: %s/\n", skillsDir)

	if len(selectedAgents) > 0 {
		names := make([]string, 0, len(selectedAgents))
		for _, a := range selectedAgents {
			if ac, ok := config.SupportedAgents[config.AgentType(a)]; ok {
				names = append(names, ac.Name)
			}
		}
		fmt.Printf("  Agents:  %s\n", strings.Join(names, ", "))
	}

	// --- Step 3: Offer starter pack ---
	packChoices := make([]huh.Option[string], 0, len(skillPacks)+1)
	for _, pack := range skillPacks {
		label := fmt.Sprintf("%s — %s", pack.Name, pack.Description)
		packChoices = append(packChoices, huh.NewOption(label, pack.Name))
	}
	packChoices = append(packChoices, huh.NewOption("Skip for now", "skip"))

	var selectedPack string
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("\nInstall a starter pack?").
				Description("Get productive quickly with curated skills").
				Options(packChoices...).
				Value(&selectedPack),
		),
	).Run()
	if err != nil {
		// User canceled, that's fine
		selectedPack = "skip"
	}

	if selectedPack != "skip" {
		installStarterPack(selectedPack, selectedAgents)
	}

	// --- Step 4: Next steps ---
	fmt.Println()
	color.Cyan("Next steps:")
	fmt.Println("  ask search          Search for skills")
	fmt.Println("  ask install <name>  Install a skill")
	fmt.Println("  ask list            View installed skills")
	fmt.Println("  ask doctor          Check your setup")
	fmt.Println()
}

func runNonInteractiveInit() {
	skillsDir := config.DefaultSkillsDir
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating skills directory: %v\n", err)
		os.Exit(1)
	}

	if err := config.CreateDefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating ask.yaml: %v\n", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	detected := config.DetectExistingToolDirs(cwd)

	color.Green("✓ Initialized ASK project")
	fmt.Println("  Created: ask.yaml")
	fmt.Printf("  Created: %s/\n", skillsDir)

	if len(detected) > 0 {
		names := make([]string, 0, len(detected))
		for _, d := range detected {
			names = append(names, d.Name)
		}
		fmt.Printf("  Detected agents: %s\n", strings.Join(names, ", "))
		fmt.Println("  Skills will be synced to all detected agents automatically.")
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  ask search          Browse available skills")
	fmt.Println("  ask install <name>  Install a skill")
	fmt.Println("  ask doctor          Check your setup")
}

func installStarterPack(packName string, agents []string) {
	var pack *struct {
		Name        string
		Description string
		Skills      []string
	}
	for i := range skillPacks {
		if skillPacks[i].Name == packName {
			pack = &skillPacks[i]
			break
		}
	}
	if pack == nil {
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		def := config.DefaultConfig()
		cfg = &def
	}

	opts := installer.InstallOptions{
		Agents: agents,
		Config: cfg,
	}

	fmt.Printf("\nInstalling %s pack (%d skills)...\n\n", pack.Name, len(pack.Skills))

	var succeeded, failed int
	for _, skillInput := range pack.Skills {
		err := installer.Install(skillInput, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", skillInput, err)
			failed++
		} else {
			color.Green("  ✓ %s", skillInput)
			succeeded++
		}
	}

	fmt.Printf("\nDone! %d installed", succeeded)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println(".")
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("yes", "y", false, "Non-interactive mode with defaults")
}

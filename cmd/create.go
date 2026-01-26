package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/skill"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new skill from a template",
	Long: `Create a new skill directory with a standardized structure.
This will generate a SKILL.md and necessary subdirectories (scripts, references, assets).`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]

		// Validate name (alphanumeric and dashes only)
		match, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", name)
		if !match {
			fmt.Println("Error: Skill name must contain only alphanumeric characters and dashes.")
			os.Exit(1)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Creating skill '%s'...\n", name)

		if err := skill.CreateSkillTemplate(name, cwd); err != nil {
			fmt.Printf("Error creating skill: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSuccessfully created skill '%s'!\n", name)
		fmt.Println("\nDirectory structure:")
		fmt.Printf("  %s/\n", name)
		fmt.Println("  ├── SKILL.md")
		fmt.Println("  ├── scripts/")
		fmt.Println("  ├── references/")
		fmt.Println("  └── assets/")
		fmt.Println("\nNext steps:")
		fmt.Printf("1. cd %s\n", name)
		fmt.Println("2. Edit SKILL.md to describe your skill")
		fmt.Println("3. Add scripts to the 'scripts' directory")
	},
}

func init() {
	skillCmd.AddCommand(createCmd)
}

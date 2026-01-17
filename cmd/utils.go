package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yeasy/ask/internal/config"
)

// ensureInitialized checks if ask.yaml exists. If not, prompts user to init.
// Returns true if initialized (or user chose to init), false if user declined.
func ensureInitialized() bool {
	if _, err := os.Stat("ask.yaml"); err == nil {
		return true // Already initialized
	}

	fmt.Print("Project not initialized. Run 'ask init' now? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Default to yes
	if input == "" || input == "y" || input == "yes" {
		runInit()
		return true
	}

	fmt.Println("Aborted. Run 'ask init' to initialize the project.")
	return false
}

// runInit executes the initialization logic
func runInit() {
	skillsDir := config.DefaultSkillsDir
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		fmt.Printf("Error creating skills directory: %v\n", err)
		return
	}

	if err := config.CreateDefaultConfig(); err != nil {
		fmt.Printf("Error creating ask.yaml: %v\n", err)
		return
	}

	fmt.Println("✓ Initialized ASK project")
	fmt.Printf("  Created: ask.yaml, %s/\n", skillsDir)
}

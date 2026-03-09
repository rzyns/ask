package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/installer"
)

var lockInstallCmd = &cobra.Command{
	Use:   "lock-install",
	Short: "Install exact versions from ask.lock",
	Long: `Install skills at the exact versions specified in ask.lock.
Similar to 'npm ci', this ensures reproducible installations.

This command reads the lock file and installs each skill using the
recorded URL and version, ensuring consistent environments across
team members and CI/CD pipelines.`,
	Example: `  # Install from project lock file
  ask lock-install

  # Install from global lock file
  ask lock-install --global`,
	Run: func(cmd *cobra.Command, _ []string) {
		global, _ := cmd.Flags().GetBool("global")
		agents, _ := cmd.Flags().GetStringSlice("agent")

		lockFile, err := config.LoadLockFileByScope(global)
		if err != nil {
			fmt.Printf("Error loading lock file: %v\n", err)
			os.Exit(1)
		}

		if len(lockFile.Skills) == 0 {
			fmt.Println("No skills found in lock file.")
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			def := config.DefaultConfig()
			cfg = &def
		}

		opts := installer.InstallOptions{
			Global: global,
			Agents: agents,
			Config: cfg,
		}

		fmt.Printf("Installing %d skills from ask.lock...\n\n", len(lockFile.Skills))

		var succeeded, failed int
		for _, entry := range lockFile.Skills {
			input := entry.URL
			if input == "" {
				input = entry.Name
			}
			// Append version if available
			if entry.Version != "" && input != "" {
				input = input + "@" + entry.Version
			}

			err := installer.Install(input, opts)
			if err != nil {
				fmt.Printf("  ✗ %s: %v\n", entry.Name, err)
				failed++
			} else {
				fmt.Printf("  ✓ %s\n", entry.Name)
				succeeded++
			}
		}

		fmt.Printf("\nDone: %d installed, %d failed.\n", succeeded, failed)

		if failed > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(lockInstallCmd)
	lockInstallCmd.Flags().StringSliceP("agent", "a", []string{}, "Target agent(s)")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/yeasy/ask/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ASK project",
	Long: `Initialize a new Agent Skills Kit project. 
This will create a default ask.yaml file in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("ask.yaml"); err == nil {
			fmt.Println("ask.yaml already exists in this directory.")
			return
		}

		err := config.CreateDefaultConfig()
		if err != nil {
			fmt.Printf("Error creating ask.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialized empty ASK project in " + "ask.yaml")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yeasy/ask/internal/github"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ask",
	Short: "Agent Skills Kit - The Package Manager for Agent Skills",
	Long: `ASK (Agent Skills Kit) is a CLI tool designed to help you discover, 
install, and manage skills for your AI Agents. 

It works similarly to package managers like Homebrew or npm, but specified for 
the Agent ecosystem.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Version: "0.7.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ask.yaml)")
	rootCmd.PersistentFlags().Bool("offline", false, "run in offline mode (no network requests)")
	rootCmd.PersistentFlags().BoolP("global", "g", false, "use global installation (~/.ask/skills)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if vid := rootCmd.PersistentFlags().Lookup("offline"); vid != nil && vid.Changed {
		if val, _ := rootCmd.PersistentFlags().GetBool("offline"); val {
			github.SetOffline(true)
		}
	}
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ask" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ask")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

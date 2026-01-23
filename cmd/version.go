package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ask",
	Long:  `All software has versions. This is ask's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ask version %s\n", rootCmd.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

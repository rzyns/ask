// Package main is the entry point for the ask CLI.
package main

import (
	"os"

	"github.com/yeasy/ask/cmd"
)

func main() {
	// If no arguments are provided, default to the GUI command
	if len(os.Args) == 1 {
		cmd.ExecuteGUI()
		return
	}
	cmd.Execute()
}

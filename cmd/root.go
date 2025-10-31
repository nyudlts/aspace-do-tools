package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "do-tools",
	Version: "v0.2.1",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("** ASPACE DO TOOLS **")
	},
}

// print help
func printNeedSubcommandHelp(cmd *cobra.Command) {
	str := getUseSequence(cmd)
	fmt.Printf("Error: missing subcommand.\nFor options, please type: %s -h\n", str)
}

func getUseSequence(cmd *cobra.Command) string {
	var seq string

	if cmd.HasParent() {
		seq = getUseSequence(cmd.Parent()) + " "
	}
	return seq + cmd.Use
}

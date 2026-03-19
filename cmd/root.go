package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represent the base command that is called without any subcomand in these case cloak is the base comand

var rootCmd = &cobra.Command{
	Use:   "cloak",
	Short: "A backup tool for untracked and modified git files",
	Long:  "Cloak safely backs up untracked or modified files in your git respository.",
	//Si no se pasa un subcomando muestra el menu de ayuda
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all the subcommands or child commands and adds the flags appropriately
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//Global flags will be here when they are needed
}

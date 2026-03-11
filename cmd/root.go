package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd representa el comando base cuando es llamado sin ningun subcomando este caso cloak
var rootCmd = &cobra.Command{
	Use:   "cloak",
	Short: "A backup tool for untracked and modified git files",
	Long:  "Cloak safely backs up untracked or modified files in your git respository.",
	//Si no se pasa un subcomando muestra el menu de ayuda
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute añade todos los subcomandos o comandos hijos y añade las flags de forma aporpiada
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//Aqui se pondrian banderas globales
}

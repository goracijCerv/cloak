package cmd

import (
	"fmt"
	"os"

	"github.com/goracijCerv/cloak/internal/config"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	outputJSON bool
	appConfig  *config.AppConfig
)

// rootCmd represent the base command that is called without any subcomand in these case cloak is the base comand

var rootCmd = &cobra.Command{
	Use:   "cloak",
	Short: "A backup tool for untracked and modified git files",
	Long:  "Cloak safely backs up untracked or modified files in your git repository.",
	//Si no se pasa un subcomando muestra el menu de ayuda
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := logger.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not initialize log file: %v\n", err)
		}

		cfg, err := config.Load()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to load config.json: %v", err))
			appConfig = &config.AppConfig{} // Fallback
		} else {
			appConfig = cfg
		}

		if appConfig.DefaultJSONOutput && !cmd.Flags().Changed("json") {
			outputJSON = true
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
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
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output results in JSON format instead of plain text.")
}

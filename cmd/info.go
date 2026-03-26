package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [backup path]",
	Short: "Show the info of the given back up folder",
	Long:  "Show the info of the given back up folder. This command takes one argument, the path of the back up folder that you want to check.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: info")
		backUpPath := args[0]
		if backUpPath == "" {
			fmt.Println("The argument is empty.")
			return
		}

		if !filepath.IsAbs(backUpPath) {
			fmt.Println("The given path is not absolute path.")
			return
		}

		_, err := os.Stat(backUpPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("The given directory doesnt exist:", backUpPath)
				return
			}
			logger.Error(fmt.Sprintf("failed to check the backup directory %s: %v", backUpPath, err))
			fmt.Println("Something went wrong for more info check the log file.")
			return
		}

		infoText, err := fileops.GetBackupInfo(backUpPath)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get the backup info: %v", err))
			fmt.Println("Failed to get the info of the back up for more info check the logs file.")
			return
		}

		fmt.Println(infoText)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

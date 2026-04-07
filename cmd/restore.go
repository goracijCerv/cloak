package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goracijCerv/cloak/internal/display"
	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	backupDirectory  string
	restoreTargetDir string
	skipConfirm      bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup of a git repository into the original folder",
	Run: func(cmd *cobra.Command, args []string) {
		shouldSkip := skipConfirm || appConfig.AlwaysSkipConfirm
		if outputJSON && !shouldSkip {
			display.PrintJSON("error", "The --yes flag is required when using --json to prevent the terminal from hanging", nil, fmt.Errorf("missing --yes flag"))
			return
		}

		if !shouldSkip {
			fmt.Println(" WARNING: You are going to overwrite files in the specific directory. Are you sure? [y/n]:")
			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				fmt.Println("Restoration canceled by the user.")
				return
			}
		}
		executeRestore()
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	actualDirectory, _ := os.Getwd()

	restoreCmd.Flags().StringVarP(&restoreTargetDir, "dir", "d", actualDirectory, "Target git repository to restore the backup into.")
	restoreCmd.Flags().StringVarP(&backupDirectory, "back", "b", "", "Backup directory to get the files that will be restore.")
	restoreCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Avoid asking the user for confirmation if they are sure they want to perform the restore.")
	err := restoreCmd.MarkFlagRequired("back")
	if err != nil {
		if outputJSON {
			display.PrintJSON("error", err.Error(), nil, err)
			return
		}
		fmt.Println(err)
		return
	}
}

func executeRestore() {
	logger.Info("COMMAND: restore\nPROCESS:restoring files")
	if restoreTargetDir == "" || backupDirectory == "" {
		if outputJSON {
			display.PrintJSON("error", "No path provided for --dir or --back", nil, fmt.Errorf("missing paths"))
			return
		}
		fmt.Println("No path provided for --dir or --back.")
		return
	}

	if !filepath.IsAbs(backupDirectory) {
		if outputJSON {
			display.PrintJSON("error", "The path to the backup directory is not absolute", nil, fmt.Errorf("not absolute path"))
			return
		}
		fmt.Println("The path to the backup directory is not absolute.")
		return
	}

	_, err := os.Stat(backupDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			if outputJSON {
				display.PrintJSON("error", "The given directory does not exist", nil, err)
				return
			}
			fmt.Println("The given directory does not exist:", backupDirectory)
			return
		}
		logger.Error(fmt.Sprintf("failed to check the backup directory %s: %v", backupDirectory, err))
		if outputJSON {
			display.PrintJSON("error", "Something went wrong checking the directory", nil, err)
			return
		}
		fmt.Println("Something went wrong. For more info check the log file.")
		return
	}

	if err := fileops.RestoreBackUp(backupDirectory, restoreTargetDir); err != nil {
		logger.Error(fmt.Sprintf("failded to make the restore: %v", err))

		if outputJSON {
			display.PrintJSON("error", "Failed to restore backup", nil, err)
			return
		}

		switch {
		case errors.Is(err, fileops.ErrEmptyManifest):
			fmt.Println("The backup folder does not have any files to restore.")
		case errors.Is(err, fileops.ErrGetManifest):
			fmt.Println("The manifest file does not exist.")
		case errors.Is(err, fileops.ErrRestoreWithErrors):
			fmt.Println("The restore completed with errors. You can find the list in the cloak log file.")
		default:
			fmt.Println("Something went wrong. Please check the cloak log file.")
		}
		return
	}

	if outputJSON {
		data := map[string]interface{}{
			"RestoredFrom": backupDirectory,
			"RestoredTo":   restoreTargetDir,
		}
		display.PrintJSON("success", "Restoration completed successfully", data, nil)
		return
	}

	fmt.Println("Restoration completed successfully.")
}

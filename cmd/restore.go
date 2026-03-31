package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		if !skipConfirm {
			fmt.Println("⚠️ WARNING: You are going to overwrite files in the specific directory. Are you sure? [y/n]:")
			var response string
			fmt.Scanln(&response)
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
	restoreCmd.MarkFlagRequired("back")
}

func executeRestore() {
	logger.Info("COMMAND: restore\nPROCESS:restoring files")
	if restoreTargetDir == "" || backupDirectory == "" {
		fmt.Println("No paths for the flags dir or back.")
		return
	}

	if !filepath.IsAbs(backupDirectory) {
		fmt.Println("The path to the backup directory is not absolute.")
		return
	}

	_, err := os.Stat(backupDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("The given directory doesnt exist:", backupDirectory)
			return
		}
		logger.Error(fmt.Sprintf("failed to check the backup directory %s: %v", backupDirectory, err))
		fmt.Println("Something went wrong for more info check the log file.")
		return
	}

	if err := fileops.RestoreBackUp(backupDirectory, restoreTargetDir); err != nil {
		logger.Error(fmt.Sprintf("failded to make the restore: %v", err))
		switch {
		case errors.Is(err, fileops.ErrEmptyManifest):
			fmt.Println("The backup folder doesnt have any files to restore.")
		case errors.Is(err, fileops.ErrGetManifest):
			fmt.Println("The manifest file doesnt exit.")
		case errors.Is(err, fileops.ErrRestoreWithErrors):
			fmt.Println("The restore completed with errors. You can find the list in the cloak logs files.")
		default:
			fmt.Println("Someting went wrong please check the cloak logs file.")
		}
		return
	}

	fmt.Println("restoration completed successfully.")
}

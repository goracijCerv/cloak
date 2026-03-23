package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	backupDirectory string
	gitDirectory2   string
	skipConfirm     bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup of a git repository into the origina folder",
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

	restoreCmd.Flags().StringVarP(&gitDirectory2, "dir", "d", actualDirectory, "Git repository directory to back up untracked changes from.")
	restoreCmd.Flags().StringVarP(&backupDirectory, "back", "b", "", "Backup directory to get the files that will be restore.")
	restoreCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Avoid asking the user for confirmation if they are sure they want to perform the restore.")
	restoreCmd.MarkFlagRequired("back")
}

func executeRestore() {
	logger.Info("COMMAND: restore\nPROCESS:restoring files")
	if err := fileops.RestoreBackUp(backupDirectory, gitDirectory2); err != nil {
		logger.Error(fmt.Sprintf("failded to make the restore: %v", err))
		switch {
		case errors.Is(err, fileops.ErrNoPaths):
			fmt.Println("The directorys paths are empty.")
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

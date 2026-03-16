package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	backupDirectory string
	gitDirectory2   string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup of a git repository into the origina folder",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	actualDirectory, _ := os.Getwd()

	restoreCmd.Flags().StringVarP(&gitDirectory2, "dir", "d", actualDirectory, "Git repository directory to back up untracked changes from.")
	restoreCmd.Flags().StringVarP(&backupDirectory, "back", "b", "", "Backup directory to get the files that will be restore")
}

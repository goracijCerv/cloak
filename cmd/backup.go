package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/git"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	gitDirectory    string
	outPutDirectory string
	messageComment  string
	dryRun          bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Back up untracked and modified files",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: backup")

		if !filepath.IsAbs(gitDirectory) {
			fmt.Println("The path of the directory to back up is not absolute.")
			return
		}

		if dryRun {
			executeDryRun()
			return
		}

		executeBackup()

	},
}

func init() {
	//We attach the command backup to the root command ('the command cloak')
	rootCmd.AddCommand(backupCmd)

	actualDirectory, _ := os.Getwd()
	//We defiend the flags for the backup subcommand

	backupCmd.Flags().StringVarP(&gitDirectory, "dir", "d", actualDirectory, "Git repository directory to back up untracked changes from.")
	backupCmd.Flags().StringVarP(&outPutDirectory, "output", "o", "", "Directory where the backup will be created (defaults to parent of -d backup folder).")
	backupCmd.Flags().StringVarP(&messageComment, "message", "m", "", "Optional label to include in the backup folder name.")
	backupCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the files to copy and the backup directory without making any changes.")
}

func executeDryRun() {
	logger.Info("PROCESS: dry-run")
	fmt.Printf("Searching in: %s\n", gitDirectory)
	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get files from git: %v", err))
		fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
		return
	}

	finalOutPutDir, err := fileops.BuildOutPutDir(outPutDirectory, &gitDirectory, messageComment)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to solve output directory: %v", err))
		fmt.Println("Unable to solve output directory. Check the cloak logs file.")
		return
	}
	fmt.Printf("The output directory will be %s \n", filepath.Clean(finalOutPutDir))
	if len(filesToCopy) == 0 {
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	fmt.Println("A list of the files that will be copied:")
	for i := range filesToCopy {
		fmt.Printf("--- %s\n", filepath.Clean(filesToCopy[i]))
	}
}

func executeBackup() {
	logger.Info("PROCESS: backing up files")
	if err := isWriteble(gitDirectory); err != nil {
		logger.Error(fmt.Sprintf("failed to check if the direcotry have writing permission: %v", err))
		fmt.Println("Is posible that a  writing permission error happened for more info check the log file.")
		return
	}

	fmt.Printf("Searching in: %s\n", gitDirectory)
	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get files from git: %v", err))
		fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
		return
	}

	if len(filesToCopy) == 0 {
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	if err := fileops.CreateNewBackUp(filesToCopy, outPutDirectory, messageComment, &gitDirectory); err != nil {
		logger.Error(fmt.Sprintf("failed to make the backup: %v", err))
		switch {
		case errors.Is(err, fileops.ErrNoFiles):
			fmt.Println("No files to back up")

		case errors.Is(err, fileops.ErrResolveOutputDir):
			fmt.Println("Unable to solve output directory")
			fmt.Println("Please check the cloak logs file to see more details of the error")

		case errors.Is(err, fileops.ErrFaildedBackDir):
			fmt.Println("Failed to create backup directory")

		case errors.Is(err, fileops.ErrBackupWithErrors):
			fmt.Println("The backup finished with errors. You can find the list of failed files in the cloak logs")

		case errors.Is(err, fileops.ErrFailedManFile):
			fmt.Println("Error in the process of generating the manifest file")
		case errors.Is(err, fileops.ErrFailedManData):
			fmt.Println("Error in the process of generating the manifest file")
		default:
			fmt.Println("Someting went wrong please check the cloak logs file.")
		}
		return
	}

	fmt.Println("Backup completed successfully.")

}

func isWriteble(path string) error {
	temFile := filepath.Join(path, ".cloak_write_test")

	file, err := os.Create(temFile)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("there are not writing permission in %s.", path)
		}
		return err
	}

	file.Close()
	os.Remove(temFile)

	return nil
}

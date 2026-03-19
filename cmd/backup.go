package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/git"
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
		// //set log file
		// logFile,err := os.OpenFile("cloak_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		// if err != nil {
		// 	fmt.Println("ERROR: Failed to open log file",err)
		// }
		// defer logFile.Close()
		// log.SetOutput(logFile)
		// log.Println("--- Application started (Backup Command) ---")

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
	fmt.Printf("Searching in: %s\n", gitDirectory)
	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		//log.Println("ERROR: Failed to get files from git:", err)
		fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
		return
	}

	finalOutPutDir, err := fileops.BuildOutPutDir(outPutDirectory, &gitDirectory, messageComment)
	if err != nil {
		//log.Println("ERROR: Unable to solve output directory:", err)
		fmt.Println("Unable to solve output directory. Check the cloak logs file.")
		return
	}
	fmt.Printf("The output directory will be %s \n", filepath.Clean(finalOutPutDir))
	if len(filesToCopy) == 0 {
		//log.Println("No files to copy")
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	fmt.Println("A list of the files that will be copied:")
	for i := range filesToCopy {
		fmt.Printf("--- %s\n", filepath.Clean(filesToCopy[i]))
	}
}

func executeBackup() {
	fmt.Printf("Searching in: %s\n", gitDirectory)
	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		//log.Println("ERROR: Failed to get files from git:", err)
		fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
		return
	}

	if len(filesToCopy) == 0 {
		//log.Println("No files to copy")
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	if err := fileops.CreateNewBackUp(filesToCopy, outPutDirectory, messageComment, &gitDirectory); err != nil {
		//log.Println("ERROR: Backup failed:", err)
		errorString := err.Error()
		switch {
		case strings.Contains(errorString, "no files provided to back up"):
			fmt.Println("No files to back up")

		case strings.Contains(errorString, "failed to resolve output directory"):
			fmt.Println("Unable to solve output directory")
			fmt.Println("Please check the cloak logs file to see more details of the error")

		case strings.Contains(errorString, "failed to create backup directory"):
			fmt.Println("Failed to create backup directory")

		case strings.Contains(errorString, "backup completed with"):
			fmt.Println("The backup finished with errors. You can find the list of failed files in the cloak logs")

		case strings.Contains(errorString, "error generating the manifest"):
			fmt.Println("Error in the process of generating the manifest file")

		}
		return
	}

	fmt.Println("Backup completed successfully.")

}

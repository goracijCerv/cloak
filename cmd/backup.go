package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goracijCerv/cloak/internal/display"
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

		_, err := os.Stat(gitDirectory)
		if err != nil {
			if os.IsNotExist(err) {
				if outputJSON {
					display.PrintJSON("error", "The given directory does not exist", nil, err)
					return
				}
				fmt.Println("The given directory does not exist:", gitDirectory)
				return
			}
			logger.Error(fmt.Sprintf("failed to check the git directory %s: %v", gitDirectory, err))
			if outputJSON {
				display.PrintJSON("error", "Something went wrong checking the directory", nil, err)
				return
			}
			fmt.Println("Something went wrong. For more info check the log file.")
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
	if outPutDirectory == "" && appConfig.DefaultOutputDir != "" {
		outPutDirectory = appConfig.DefaultOutputDir
	}

	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get files from git: %v", err))
		if outputJSON {
			display.PrintJSON("error", "Failed to get files from git", nil, err)
			return
		}
		fmt.Println("Something went wrong getting the files. Please check the cloak log file.")
		return
	}

	finalOutPutDir, _, err := fileops.BuildOutPutDir(outPutDirectory, &gitDirectory, messageComment) //the time.Time is not required in the dry run
	if err != nil {
		logger.Error(fmt.Sprintf("failed to solve output directory: %v", err))
		if outputJSON {
			display.PrintJSON("error", "Unable to resolve output directory", nil, err)
			return
		}
		fmt.Println("Unable to resolve output directory. Check the cloak log file.")
		return
	}

	if outputJSON {
		data := map[string]interface{}{
			"OutputDirectory": filepath.Clean(finalOutPutDir),
			"TotalFiles":      len(filesToCopy),
			"FilesToCopy":     filesToCopy,
		}
		display.PrintJSON("success", "Dry-run executed successfully", data, nil)
		return
	}

	fmt.Printf("Searching in: %s\n", gitDirectory)
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
	if outPutDirectory == "" && appConfig.DefaultOutputDir != "" {
		outPutDirectory = appConfig.DefaultOutputDir
	}

	finalOutPutDir, timeCreated, err := fileops.BuildOutPutDir(outPutDirectory, &gitDirectory, messageComment)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to solve output directory: %v", err))
		if outputJSON {
			display.PrintJSON("error", "Unable to resolve output directory", nil, err)
			return
		}
		fmt.Println("Unable to resolve output directory. Check the cloak log file.")
		return
	}

	if outPutDirectory != "" {
		if !filepath.IsAbs(finalOutPutDir) {
			if outputJSON {
				display.PrintJSON("error", "Output directory path is not absolute", nil, fmt.Errorf("not absolute path"))
				return
			}
			fmt.Println("Output directory path is not absolute.")
			return
		}

		if _, err := os.Stat(finalOutPutDir); errors.Is(err, os.ErrNotExist) {
			if err := isWritable(filepath.Dir(finalOutPutDir)); err != nil {
				logger.Error(fmt.Sprintf("parent output directory is not writable: %v", err))
				if outputJSON {
					display.PrintJSON("error", "No write permission to create the output directory", nil, err)
					return
				}
				fmt.Println("No write permission to create the output directory. Check the log file.")
				return
			}
		} else {

			if err := isWritable(finalOutPutDir); err != nil {
				logger.Error(fmt.Sprintf("output directory is not writable: %v", err))
				if outputJSON {
					display.PrintJSON("error", "No write permission for the output directory", nil, err)
					return
				}
				fmt.Println("No write permission for the output directory. Check the log file.")
				return
			}
		}

	}

	filesToCopy, err := git.GetFiles(&gitDirectory)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get files from git: %v", err))
		if outputJSON {
			display.PrintJSON("error", "Failed to get files from git", nil, err)
			return
		}
		fmt.Println("Something went wrong getting the files. Please check the cloak log file.")
		return
	}

	if len(filesToCopy) == 0 {
		if outputJSON {
			// noting to backup.
			display.PrintJSON("success", "Nothing to back up", map[string]interface{}{"TotalFiles": 0}, nil)
			return
		}
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	if !outputJSON {
		fmt.Printf("Searching in: %s\n", gitDirectory)
		fmt.Println("Backup destination:", finalOutPutDir)
	}

	if err := fileops.CreateNewBackUp(filesToCopy, finalOutPutDir, &gitDirectory, timeCreated); err != nil {
		logger.Error(fmt.Sprintf("failed to make the backup: %v", err))

		if outputJSON {
			display.PrintJSON("error", "Failed to make the backup", nil, err)
			return
		}

		switch {
		case errors.Is(err, fileops.ErrNoFiles):
			fmt.Println("No files to back up.")

		case errors.Is(err, fileops.ErrFaildedBackDir):
			fmt.Println("Failed to create the backup directory.")

		case errors.Is(err, fileops.ErrBackupWithErrors):
			fmt.Println("The backup finished with errors. You can find the failed files in the cloak log.")

		case errors.Is(err, fileops.ErrFailedManFile):
			fmt.Println("Error generating the manifest file.")
		case errors.Is(err, fileops.ErrFailedManData):
			fmt.Println("Error generating the manifest data.")
		default:
			fmt.Println("Something went wrong. Please check the cloak log file.")
		}
		return
	}

	if outputJSON {
		data := map[string]interface{}{
			"OutputDirectory": finalOutPutDir,
			"TotalFiles":      len(filesToCopy),
			"FilesBackedUp":   filesToCopy,
		}
		display.PrintJSON("success", "Backup completed successfully", data, nil)
		return
	}

	fmt.Println("Backup completed successfully.")

}

func isWritable(path string) error {
	tempFile := filepath.Join(path, ".cloak_write_test")
	// #nosec G304 -- This tool needs to read arbitrary files by design
	file, err := os.Create(tempFile)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("there are not writing permission in %s.", path)
		}
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}
	err = os.Remove(tempFile)
	if err != nil {
		return err
	}
	return nil
}

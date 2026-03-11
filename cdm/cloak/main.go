package main

import (
	"cloak/internal/fileops"
	"cloak/internal/git"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	logFile, err := os.OpenFile("cloak_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("ERROR: Failed to open log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Application started")

	actualDirectory, err := os.Getwd()
	if err != nil {
		log.Println("ERROR: Failed to get current working directory:", err)
	}

	gitDirectory := flag.String("d", actualDirectory, "Git repository directory to back up untracked changes from.")
	outPutDirectory := flag.String("o", "", "Directory where the backuo will be created (defaults to parent of -d backup folder).")
	messageComment := flag.String("m", "", "Optional label to include in the backup folder name.")
	dryRun := flag.Bool("dry-run", false, "Show the files to copy and the backup directory without making any changes.")
	flag.Parse()

	if *dryRun {
		filesToCopy, err := git.GetFiles(gitDirectory)
		if err != nil {
			log.Println("ERROR: Failed to get files from git:", err)
			switch err.Error() {
			case "directory path must not be empty":
				fmt.Println("The directory path is empty string")
				return
			case "directory path must be absolute":
				fmt.Println("The directory path is not an absolute path")
				return
			case "must be run inside a git repository":
				fmt.Println("Error: Cloak must be run inside a git repository")
				return
			default:
				fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
				return
			}

		}
		finalOutPutDir, err := fileops.BuildOutPutDir(*outPutDirectory, gitDirectory, *messageComment)
		if err != nil {
			log.Println("ERROR: Failed to get output directory: ", err)
			fmt.Println("Unable to solve output directory")
			fmt.Println("Please check the cloak logs file to see more details of the error")
			return
		}

		fmt.Printf("The output directory will be: %s \n", filepath.Clean(finalOutPutDir))
		fmt.Println("A list of the files that will be copy:")
		for i := range filesToCopy {
			fmt.Printf("--- %s\n", filepath.Clean(filesToCopy[i]))
		}

		return
	}

	filesToCopy, err := git.GetFiles(gitDirectory)
	if err != nil {
		log.Fatalln("ERROR: Failed to get files from git:", err)
		switch err.Error() {
		case "directory path must not be empty":
			fmt.Println("The directory path is empty string")
			return
		case "directory path must be absolute":
			fmt.Println("The directory path is not an absolute path")
			return

		case "must be run inside a git repository":
			fmt.Println("Error: Cloak must be run inside a git repository")
			return
		default:
			fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
			return
		}
	}

	if len(filesToCopy) == 0 {
		log.Println("ERROR: No files to copy")
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	//fmt.Println("Files to back up:", filesToCopy)

	if err := fileops.CreateNewBackUp(filesToCopy, *outPutDirectory, *messageComment, gitDirectory); err != nil {
		log.Println("ERROR: Backup failed:", err)
		errorString := err.Error()
		switch {
		case strings.Contains(errorString, "no files provided to back up"):
			fmt.Println("No files to back up")
			return
		case strings.Contains(errorString, "failed to resolve output directory"):
			fmt.Println("Unable to solve output directory")
			fmt.Println("Please check the cloak logs file to see more details of the error")
			return
		case strings.Contains(errorString, "failed to create backup directory"):
			fmt.Println("Failed to create backup directory")
			return
		case strings.Contains(errorString, "backup completed with"):
			fmt.Println("The backup finished with errors. You can find the list of failed files in the cloak logs")
			return
		}
	}

	fmt.Println("Backup completed successfully.")
}

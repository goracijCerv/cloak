package main

import (
	"cloak/internal/fileops"
	"cloak/internal/git"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	logFile, err := os.OpenFile("cloak_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Application started")

	actualDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get current working directory:", err)
	}

	gitDirectory := flag.String("d", actualDirectory, "Git repository directory to back up untracked changes from.")
	outPutDirectory := flag.String("o", "", "Directory where the backuo will be created (defaults to parent of -d backup folder).")
	messageComment := flag.String("m", "", "Optional label to include in the backup folder name.")
	flag.Parse()

	filesToCopy, err := git.GetFiles(gitDirectory)
	if err != nil {
		log.Fatalln("Failed to get files from git:", err)
		switch err.Error() {
		case "directory path must not be empty":
			fmt.Println("The directory path is empty string")
			return
		case "directory path must be absolute":
			fmt.Println("The directory path is not an absolute path")
			return
		default:
			fmt.Println("Something went wrong when getting the files, please check the cloak logs file")
			return
		}
	}

	if len(filesToCopy) == 0 {
		log.Println("No files to copy")
		fmt.Println("Nothing to back up: no untracked or modified files found.")
		return
	}

	//fmt.Println("Files to back up:", filesToCopy)

	if err := fileops.CreateNewBackUp(filesToCopy, *outPutDirectory, *messageComment, gitDirectory); err != nil {
		log.Fatalln("Backup failed:", err)
		errorString := err.Error()
		switch {
		case strings.Contains(errorString, "no files provided to back up"):
			fmt.Println("No files to back up")
			return
		case strings.Contains(errorString, "failed to resolve output directory"):
			fmt.Println("Unable to solve output directory")
			fmt.Println("Please check the cloak logs file to see more detail error")
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

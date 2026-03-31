package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	deleteBackups     []string
	deleteAll         bool
	deleteBefore      string
	deleteAfter       string
	deleteSkipConfirm bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Permently delete one or more backups.",
	Long: `Delete specific backups by path, or filter them by date. 
You can delete multiple backups at once.

Examples:
  cloak delete --back /ruta/al/backup1 --back /ruta/al/backup2
  cloak delete --all
  cloak delete --before 2026-03-01 --after 2026-01-01
  cloak delete --all -y`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: delete")

		if len(deleteBackups) == 0 && !deleteAll && deleteBefore == "" && deleteAfter == "" {
			fmt.Println("Error: You must specify what to delete (--back, --all, --before, or --after).")
			cmd.Help()
			return
		}

		executeDelete()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringArrayVarP(&deleteBackups, "back", "b", []string{}, "Path to the back up directory to delete (can be used multiple times).")
	deleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all the backups for the current repository.")
	deleteCmd.Flags().StringVar(&deleteBefore, "before", "", "Delete backups created before this date (Format: YYYY-MM-DD).")
	deleteCmd.Flags().StringVar(&deleteAfter, "after", "", "Delete backups created after this date (Format: YYYY-MM-DD).")

	deleteCmd.Flags().BoolVarP(&deleteSkipConfirm, "yes", "y", false, "Avoid asking for confirmation before deleting.")

	//exclusivity rules for the flags
	deleteCmd.MarkFlagsMutuallyExclusive("back", "all")
	deleteCmd.MarkFlagsMutuallyExclusive("back", "before")
	deleteCmd.MarkFlagsMutuallyExclusive("back", "after")

	deleteCmd.MarkFlagsMutuallyExclusive("all", "before")
	deleteCmd.MarkFlagsMutuallyExclusive("all", "after")
}

func executeDelete() {

	if deleteAll {
		logger.Info("PROCESS: delete all")
		backkUpsPaths, err := fileops.AllBackUpsPaths("", "")
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get all the back up of the working directory: %v", err))
			switch {
			case errors.Is(err, fileops.ErrNoBackUpDir):
				fmt.Println("There is't any  backups folder.")
			case errors.Is(err, fileops.ErrGetManifest):
				fmt.Println("An error happen trying to reed the manifest file of a backup. for more info checl the log file.")
			case errors.Is(err, fileops.ErrNoBackUps):
				fmt.Println("The working directory doesn't have any backup.")
			default:
				fmt.Println("Something went wrong for more info check the log file.")
			}
			return
		}

		if !deleteSkipConfirm {
			fmt.Printf("⚠️ WARNING: The following %d backup(s) will be permanently deleted:\n", len(backkUpsPaths))
			for _, v := range backkUpsPaths {
				fmt.Printf("  %s\n", filepath.Base(v))
			}
			fmt.Println("Are you sure? [y/n]:")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				fmt.Println("Delete canceled by the user.")
				return
			}
		}

		if err := fileops.DeleteByPaths(backkUpsPaths...); err != nil {
			logger.Error(fmt.Sprintf("failed to get delete all the paths: %v", err))
			switch {
			case errors.Is(err, fileops.ErrFaildedDelete):
				fmt.Println("An error happen trying to delete the backups for more info check the log file.")
			default:
				fmt.Println("Something went wrong for more info check the log file.")
			}
			return
		}

		fmt.Println("The backups were deleted correctly.")
		return
	}

	if deleteBefore != "" || deleteAfter != "" {
		logger.Info("PROCESS: delete by dates")
		backkUpsPaths, err := fileops.AllBackUpsPaths(deleteAfter, deleteBefore)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get all the back up of the working directory: %v", err))
			switch {
			case errors.Is(err, fileops.ErrNoBackUpDir):
				fmt.Println("There is't any  backups folder.")
			case errors.Is(err, fileops.ErrGetManifest):
				fmt.Println("An error happen trying to reed the manifest file of a backup. for more info checl the log file.")
			case errors.Is(err, fileops.ErrNoBackUps):
				fmt.Println("The working directory doesn't have any backup.")
			case errors.Is(err, fileops.ErrDateNotValidFormat):
				fmt.Println("The format of the dates are not valid, the valid format is YYYY-MM-DD.")
			default:
				fmt.Println("Something went wrong for more info check the log file.")
			}
			return
		}

		if !deleteSkipConfirm {
			fmt.Printf("⚠️ WARNING: The following %d backup(s) will be permanently deleted:\n", len(backkUpsPaths))
			for _, v := range backkUpsPaths {
				fmt.Printf("  %s\n", filepath.Base(v))
			}
			fmt.Println("Are you sure? [y/n]:")
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				fmt.Println("Delete canceled by the user.")
				return
			}
		}

		if err := fileops.DeleteByPaths(backkUpsPaths...); err != nil {
			logger.Error(fmt.Sprintf("failed to get delete all the paths: %v", err))
			switch {
			case errors.Is(err, fileops.ErrFaildedDelete):
				fmt.Println("An error happen trying to delete the backups for more info check the log file.")
			default:
				fmt.Println("Something went wrong for more info check the log file.")
			}
			return
		}

		fmt.Println("The backups were deleted correctly.")
		return
	}

	logger.Info("PROCESS: delete given backups paths")
	if !deleteSkipConfirm {
		fmt.Printf("⚠️ WARNING: The following %d backup(s) will be permanently deleted:\n", len(deleteBackups))
		for _, v := range deleteBackups {
			fmt.Printf("  %s\n", filepath.Base(v))
		}
		fmt.Println("Are you sure? [y/n]:")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("Delete canceled by the user.")
			return
		}
	}

	if err := fileops.DeleteByPaths(deleteBackups...); err != nil {
		logger.Error(fmt.Sprintf("failed to get delete all the paths: %v", err))
		switch {
		case errors.Is(err, fileops.ErrFaildedDelete):
			fmt.Println("An error happen trying to delete the backups for more info check the log file.")
		default:
			fmt.Println("Something went wrong for more info check the log file.")
		}
		return
	}

	fmt.Println("The backups were deleted correctly.")
}

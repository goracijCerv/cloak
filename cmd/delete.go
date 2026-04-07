package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goracijCerv/cloak/internal/display"
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
	Short: "Permanently delete one or more backups.",
	Long: `Permanently delete one or more backups by path, or filter them by date.
You can delete multiple backups at once.

Examples:
  cloak delete --back /home/user/backup/[myrepo]2026-03-01_10-00-00
  cloak delete --back /home/user/backup/[myrepo]2026-03-01_10-00-00 --back /home/user/backup/[myrepo]2026-04-01_10-00-00
  cloak delete --all
  cloak delete --before 2026-03-01
  cloak delete --after 2026-01-01
  cloak delete --before 2026-03-01 --after 2026-01-01
  cloak delete --all --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: delete")

		if len(deleteBackups) == 0 && !deleteAll && deleteBefore == "" && deleteAfter == "" {
			if outputJSON {
				display.PrintJSON("error", "You must specify what to delete (--back, --all, --before, or --after).", nil, fmt.Errorf("no target specified"))
				return
			}
			fmt.Println("You must specify what to delete (--back, --all, --before, or --after).")
			_ = cmd.Help()
			return
		}

		if outputJSON && !deleteSkipConfirm {
			display.PrintJSON("error", "The --yes flag is required when using --json to prevent the terminal from hanging", nil, fmt.Errorf("missing --yes flag"))
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
			if outputJSON {
				display.PrintJSON("error", "Failed to get backups for this directory", nil, err)
				return
			}
			switch {
			case errors.Is(err, fileops.ErrNoBackUpDir):
				fmt.Println("There is no backups folder for this directory.")
			case errors.Is(err, fileops.ErrGetManifest):
				fmt.Println("An error occurred reading the manifest file of a backup. For more info check the log file.")
			case errors.Is(err, fileops.ErrNoBackUps):
				fmt.Println("The current directory does not have any backups.")
			default:
				fmt.Println("Something went wrong. For more info check the log file.")
			}
			return
		}
		deletePaths(backkUpsPaths)
		return
	}

	if deleteBefore != "" || deleteAfter != "" {
		logger.Info("PROCESS: delete by dates")
		backkUpsPaths, err := fileops.AllBackUpsPaths(deleteAfter, deleteBefore)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get all the back up of the working directory: %v", err))
			if outputJSON {
				display.PrintJSON("error", "Failed to filter backups by date", nil, err)
				return
			}
			switch {
			case errors.Is(err, fileops.ErrNoBackUpDir):
				fmt.Println("There is no backups folder for this directory.")
			case errors.Is(err, fileops.ErrGetManifest):
				fmt.Println("An error occurred reading the manifest file of a backup. For more info check the log file.")
			case errors.Is(err, fileops.ErrNoBackUps):
				fmt.Println("The current directory does not have any backups.")
			case errors.Is(err, fileops.ErrDateNotValidFormat):
				fmt.Println("The date format is not valid. The expected format is YYYY-MM-DD.")
			default:
				fmt.Println("Something went wrong. For more info check the log file.")
			}
			return
		}
		deletePaths(backkUpsPaths)
		return
	}

	logger.Info("PROCESS: delete given backups paths")
	deletePaths(deleteBackups)
}

func deletePaths(paths []string) {
	shouldSkip := deleteSkipConfirm || appConfig.AlwaysSkipConfirm
	if !shouldSkip && !confirmDeletion(paths) {
		if outputJSON {
			display.PrintJSON("success", "Delete canceled by the user", nil, nil)
			return
		}
		fmt.Println("Delete canceled by the user.")
		return
	}

	if err := fileops.DeleteByPaths(paths...); err != nil {
		logger.Error(fmt.Sprintf("failed to get all the back up of the working directory: %v", err))
		if outputJSON {
			display.PrintJSON("error", "Failed to delete one or more backups", nil, err)
			return
		}
		switch {
		case errors.Is(err, fileops.ErrFaildedDelete):
			fmt.Println("An error occurred trying to delete the backups. For more info check the log file.")
		default:
			fmt.Println("Something went wrong. For more info check the log file.")
		}
		return
	}

	if outputJSON {
		data := map[string]interface{}{
			"DeletedCount": len(paths),
			"DeletedPaths": paths,
		}
		display.PrintJSON("success", "The backups were deleted successfully", data, nil)
		return
	}

	fmt.Println("The backups were deleted successfully.")
}

func confirmDeletion(paths []string) bool {
	fmt.Printf(" WARNING: The following %d backup(s) will be permanently deleted:\n", len(paths))
	for _, v := range paths {
		fmt.Printf("  %s\n", filepath.Base(v))
	}
	fmt.Println("Are you sure? [y/n]:")
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

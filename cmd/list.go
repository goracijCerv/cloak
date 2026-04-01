package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/goracijCerv/cloak/internal/display"
	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all backups available for the current repository",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: list\nPROCESS: list")
		currentDir, err := os.Getwd()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get the working directory: %v", err))
			if outputJSON {
				display.PrintJSON("error", "Failed to get the working directory", nil, err)
				return
			}
			fmt.Println("Failed to get the working directory. For more info check the log file.")
			return
		}

		repoDir := filepath.Clean(currentDir)
		backups, err := fileops.ListBackups(repoDir)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get the list: %v", err))
			if outputJSON {
				display.PrintJSON("error", "Failed to retrieve backups", nil, err)
				return
			}
			switch {
			case errors.Is(err, fileops.ErrNoBackUps):
				fmt.Println("There are no backups for this directory.")

			case errors.Is(err, fileops.ErrNoBackUpDir):
				fmt.Println("The backup directory does not exist.")

			case errors.Is(err, fileops.ErrFailedBackDir):
				fmt.Println("Failed to read the backup directory. For more info check the log file.")

			default:
				fmt.Println("Something went wrong. For more info check the log file.")
			}
			return
		}

		if outputJSON {
			display.PrintJSON("success", "Backups retrieved successfully", backups, nil)
			return
		}

		fmt.Printf("%d backup(s) found for '%s';\n\n", len(backups), filepath.Base(repoDir))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "BACKUP NAME\tFILES\tPATH")
		fmt.Fprintln(w, "-------------------\t--------\t----")

		for _, b := range backups {
			fmt.Fprintf(w, "%s\t%d\t%s\n", b.Name, b.FileCount, b.Path)
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

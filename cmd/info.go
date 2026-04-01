package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goracijCerv/cloak/internal/display"
	"github.com/goracijCerv/cloak/internal/fileops"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [backup path]",
	Short: "Show the info of the given back up folder",
	Long:  "Show the info of the given back up folder. This command takes one argument, the path of the back up folder that you want to check.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("COMMAND: info")
		backUpPath := args[0]
		if backUpPath == "" {

			if outputJSON {
				display.PrintJSON("error", "The argument is empty", nil, fmt.Errorf("empty path argument"))
				return
			}

			fmt.Println("The argument is empty.")
			return
		}

		if !filepath.IsAbs(backUpPath) {

			if outputJSON {
				display.PrintJSON("error", "The given path is not an absolute path", nil, fmt.Errorf("non-absolute path: %s", backUpPath))
				return
			}

			fmt.Println("The given path is not an absolute path.")
			return
		}

		_, err := os.Stat(backUpPath)
		if err != nil {
			if os.IsNotExist(err) {

				if outputJSON {
					display.PrintJSON("error", "The given directory does not exist", nil, err)
					return
				}

				fmt.Println("The given directory does not exist:", backUpPath)
				return
			}
			logger.Error(fmt.Sprintf("failed to check the backup directory %s: %v", backUpPath, err))

			if outputJSON {
				display.PrintJSON("error", "Something went wrong checking the directory", nil, err)
				return
			}

			fmt.Println("Something went wrong. For more info check the log file.")
			return
		}

		infoData, err := fileops.GetBackupInfo(backUpPath)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get the backup info: %v", err))

			if outputJSON {
				display.PrintJSON("error", "Failed to get the backup info", nil, err)
				return
			}

			fmt.Println("Failed to get the backup info. For more info check the log file.")
			return
		}

		if outputJSON {
			display.PrintJSON("success", "Backup info retrieved successfully", infoData, nil)
			return
		}

		text := writeText(isColorSupported(), *infoData)
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func isColorSupported() bool {
	return os.Getenv("NO_COLOR") == ""
}

func writeText(color bool, infoData fileops.BackupDetails) string {
	var sb strings.Builder
	if color {
		formattedTime := infoData.CreatedAt.Format("02 Jan 2006 at 3:04 pm")
		divider := display.ColorGray + "=================================================" + display.ColorReset + "\n"
		sb.WriteString("\n" + display.ColorCyan + "📦 BACKUP DETAILS" + display.ColorReset + "\n")
		sb.WriteString(divider)
		fmt.Fprintf(&sb, "%s📍 Path:    %s %s%s%s\n", display.ColorYellow, display.ColorReset, display.ColorGreen, infoData.Path, display.ColorReset)
		fmt.Fprintf(&sb, "%s📅 Created:  %s %s%s%s\n", display.ColorYellow, display.ColorReset, display.ColorGreen, formattedTime, display.ColorReset)
		fmt.Fprintf(&sb, "%s📄 Files:%s %s%d total%s\n", display.ColorYellow, display.ColorReset, display.ColorGreen, len(infoData.Entries), display.ColorReset)
		sb.WriteString(divider)

		if len(infoData.Entries) > 0 {
			sb.WriteString("List of original files saved:\n")
			for _, entry := range infoData.Entries {
				fmt.Fprintf(&sb, "  %s-%s %s\n", display.ColorGray, display.ColorReset, entry.OriginalPath)
			}
		} else {
			sb.WriteString(display.ColorYellow + "The backup doesn't have any files.\n" + display.ColorReset)
		}
		sb.WriteString("\n")
	} else {
		formattedTime := infoData.CreatedAt.Format("02 Jan 2006 at 3:04 pm")
		divider := "=================================================\n"
		sb.WriteString("\n📦 BACKUP DETAILS\n")
		sb.WriteString(divider)
		fmt.Fprintf(&sb, "📍 Path: %s\n", infoData.Path)
		fmt.Fprintf(&sb, "📅 Created: %s\n", formattedTime)
		fmt.Fprintf(&sb, "📄 Files: %d total\n", len(infoData.Entries))
		sb.WriteString(divider)

		if len(infoData.Entries) > 0 {
			sb.WriteString("List of original files saved:\n")
			for _, entry := range infoData.Entries {
				fmt.Fprintf(&sb, "  - %s\n", entry.OriginalPath)
			}
		} else {
			sb.WriteString("The backup doesn't have any files.\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

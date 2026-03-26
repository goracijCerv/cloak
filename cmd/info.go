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
			fmt.Println("The argument is empty.")
			return
		}

		if !filepath.IsAbs(backUpPath) {
			fmt.Println("The given path is not absolute path.")
			return
		}

		_, err := os.Stat(backUpPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("The given directory doesnt exist:", backUpPath)
				return
			}
			logger.Error(fmt.Sprintf("failed to check the backup directory %s: %v", backUpPath, err))
			fmt.Println("Something went wrong for more info check the log file.")
			return
		}

		infoData, err := fileops.GetBackupInfo(backUpPath)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get the backup info: %v", err))
			fmt.Println("Failed to get the info of the back up for more info check the logs file.")
			return
		}

		var sb strings.Builder
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

		fmt.Println(sb.String())
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

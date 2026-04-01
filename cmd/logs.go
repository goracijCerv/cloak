package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/goracijCerv/cloak/internal/display"
	"github.com/goracijCerv/cloak/internal/logger"
	"github.com/spf13/cobra"
)

var (
	showPath  bool
	clearLogs bool
	tailLines int
	showAll   bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View or manage the cloak log file",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//skip logger.init for this subcommand
	},
	Run: func(cmd *cobra.Command, args []string) {
		logPath, err := logger.LogPath()
		if err != nil {
			if outputJSON {
				display.PrintJSON("error", "Could not resolve log file path", nil, err)
				return
			}
			fmt.Println("Could not resolve log file path:", err)
			return
		}

		if showPath {
			if outputJSON {
				display.PrintJSON("success", "Log path retrieved successfully", map[string]string{"path": logPath}, nil)
				return
			}
			fmt.Println(logPath)
			return
		}

		if clearLogs {
			if err := os.WriteFile(logPath, []byte{}, 0644); err != nil {
				if outputJSON {
					display.PrintJSON("error", "Failed to clear log file", nil, err)
					return
				}
				fmt.Println("Failed to clear log file:", err)
				return
			}
			if outputJSON {
				display.PrintJSON("success", "Log file cleared", nil, nil)
				return
			}
			fmt.Println("Log file cleared.")
			return
		}

		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			if outputJSON {
				display.PrintJSON("error", "No log file found. Run a backup or restore first.", nil, fmt.Errorf("file not found"))
				return
			}
			fmt.Println("No log file found. Run a backup or restore first.")
			return
		}

		var lines []string
		var readErr error

		if showAll {
			lines, readErr = getAllLines(logPath)
		} else {
			lines, readErr = getTailLines(logPath, tailLines)
		}

		if readErr != nil {
			if outputJSON {
				display.PrintJSON("error", "Failed to read log file", nil, readErr)
				return
			}
			fmt.Println("Failed to read log file.", readErr)
			return
		}

		if outputJSON {
			display.PrintJSON("success", "Logs retrieved successfully", map[string]interface{}{"lines": lines}, nil)
			return
		}

		if len(lines) == 0 {
			fmt.Println("Log file is empty.")
			return
		}

		for _, line := range lines {
			fmt.Println(line)
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVar(&showPath, "path", false, "Print the path of the log file.")
	logsCmd.Flags().BoolVar(&clearLogs, "clear", false, "Clear the log file.")
	logsCmd.Flags().BoolVar(&showAll, "all", false, "Print the entire log file.")
	logsCmd.Flags().IntVar(&tailLines, "tail", 20, "Number of recent lines to show.")
}

func getAllLines(logpath string) ([]string, error) {
	content, err := os.ReadFile(logpath)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return []string{}, nil
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

func getTailLines(logPath string, n int) ([]string, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

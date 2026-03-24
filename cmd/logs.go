package cmd

import (
	"bufio"
	"fmt"
	"os"

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
			fmt.Println("Could not resolve log file path:", err)
			return
		}

		if showPath {
			fmt.Println(logPath)
			return
		}

		if clearLogs {
			if err := os.WriteFile(logPath, []byte{}, 0644); err != nil {
				fmt.Println("Failed to clear log file:", err)
				return
			}
			fmt.Println("Log file cleared.")
			return
		}

		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Println("No log file found. Run a backup or restore first.")
			return
		}

		if showAll {
			printAll(logPath)
			return
		}

		printTail(logPath, tailLines)
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVar(&showPath, "path", false, "Print the path of the log file.")
	logsCmd.Flags().BoolVar(&clearLogs, "clear", false, "Clear the log file.")
	logsCmd.Flags().BoolVar(&showAll, "all", false, "Print the entire log file.")
	logsCmd.Flags().IntVar(&tailLines, "tail", 20, "Number of recent lines to show.")
}

func printAll(logpath string) {
	content, err := os.ReadFile(logpath)
	if err != nil {
		fmt.Println("Failed to read log file.", err)
		return
	}

	if len(content) == 0 {
		fmt.Println("Log file is empty.")
		return
	}

	fmt.Print(string(content))
}

func printTail(logPath string, n int) {
	f, err := os.Open(logPath)
	if err != nil {
		fmt.Println("Failed to open log file.", err)
		return
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

	if len(lines) == 0 {
		fmt.Println("Log file is empty.")
		return
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

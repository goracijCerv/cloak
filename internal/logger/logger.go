package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const maxLogSize = 5 * 1024 * 1024 //5MB
var fileLogger *log.Logger

func Init() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not find config directory: %w", err)
	}

	cloakDir := filepath.Join(configDir, "cloak")
	if err := os.MkdirAll(cloakDir, 0750); err != nil {
		return fmt.Errorf("could not create cloak config directory: %w", err)
	}

	logPath := filepath.Join(cloakDir, "cloak.log")

	if err := rotateFileIfNeeded(logPath); err != nil {
		// rotation failing shouldn't block the tool
		fmt.Fprintf(os.Stderr, "warning: could not rotate log file: %v\n", err)
	}

	// #nosec G304 -- This tool needs to read arbitrary files by design
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}

	fileLogger = log.New(f, "", log.LstdFlags)
	return nil

}

func rotateFileIfNeeded(logPath string) error {
	info, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if info.Size() < maxLogSize {
		return nil
	}

	return os.Rename(logPath, logPath+".1")
}

func Error(msg string) {
	if fileLogger != nil {
		fileLogger.Println("[ERROR]", msg)
	}
}

func Info(msg string) {
	if fileLogger != nil {
		fileLogger.Println("[INFO]", msg)
	}
}

func LogPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cloak", "cloak.log"), nil
}

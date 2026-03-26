package fileops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Al the info of an backup
type BackupInfo struct {
	Name      string
	FileCount int
	Path      string
}

var (
	ErrFailedBackDir = errors.New("failed to read the backup directory")
	ErrNoBackUps     = errors.New("there are no backups")
	ErrNoBackUpDir   = errors.New("there are no backup folder")
)

const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[1;36m" // Cian en negrita
	ColorYellow = "\033[33m"
	ColorGreen  = "\033[32m"
	ColorGray   = "\033[90m"
)

// ListBackups search all the backups releted to an origiin directory
func ListBackups(repoDir string) ([]BackupInfo, error) {
	parentDir := filepath.Dir(repoDir)
	repoName := filepath.Base(repoDir)
	backupBaseDir := filepath.Join(parentDir, "backup")

	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoBackUpDir //there are no backups
		}
		return nil, fmt.Errorf("%w: %w", ErrFailedBackDir, err)
	}

	prefix := fmt.Sprintf("[%s]", repoName)
	var backups []BackupInfo

	for _, entry := range entries {
		//we only search for the folders that start with the prefix
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backupPath := filepath.Join(backupBaseDir, entry.Name())
			fileCount := 0
			if manifest, err := readManifest(backupPath); err == nil {
				fileCount = len(manifest.Entries)
			}

			backups = append(backups, BackupInfo{
				Name:      entry.Name(),
				FileCount: fileCount,
				Path:      backupPath,
			})
		}
	}

	if len(backups) == 0 {
		return nil, ErrNoBackUps
	}

	return backups, nil
}

func GetBackupInfo(backUpPath string) (string, error) {
	manifest, err := readManifest(backUpPath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGetManifest, err)
	}

	var sb strings.Builder
	formattedTime := manifest.CreatedAt.Format("02 Jan 2006 at 3:04 pm")
	divider := ColorGray + "=================================================" + ColorReset + "\n"
	sb.WriteString("\n" + ColorCyan + "📦 BACKUP DETAILS" + ColorReset + "\n")
	sb.WriteString(divider)
	fmt.Fprintf(&sb, "%s📍 Path:    %s %s%s%s\n", ColorYellow, ColorReset, ColorGreen, backUpPath, ColorReset)
	fmt.Fprintf(&sb, "%s📅 Created:  %s %s%s%s\n", ColorYellow, ColorReset, ColorGreen, formattedTime, ColorReset)
	fmt.Fprintf(&sb, "%s📄 Files:%s %s%d total%s\n", ColorYellow, ColorReset, ColorGreen, len(manifest.Entries), ColorReset)
	sb.WriteString(divider)

	// 3. Iterar sobre la lista de archivos
	if len(manifest.Entries) > 0 {
		sb.WriteString("List of original files saved:\n")
		for _, entry := range manifest.Entries {
			fmt.Fprintf(&sb, "  %s-%s %s\n", ColorGray, ColorReset, entry.OriginalPath)
		}
	} else {
		sb.WriteString(ColorYellow + "The back uo doesnt have any files.\n" + ColorReset)
	}

	sb.WriteString("\n")

	return sb.String(), nil

}

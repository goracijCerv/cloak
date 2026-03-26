package fileops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Al the info of an backup
type BackupInfo struct {
	Name      string
	FileCount int
	Path      string
}

type BackupDetails struct {
	Path      string
	CreatedAt time.Time
	Entries   []ManifestEntry
}

var (
	ErrFailedBackDir = errors.New("failed to read the backup directory")
	ErrNoBackUps     = errors.New("there are no backups")
	ErrNoBackUpDir   = errors.New("there are no backup folder")
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

func GetBackupInfo(backUpPath string) (*BackupDetails, error) {
	manifest, err := readManifest(backUpPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetManifest, err)
	}

	backUpDetails := BackupDetails{
		Path:      backUpPath,
		CreatedAt: manifest.CreatedAt,
		Entries:   manifest.Entries,
	}

	return &backUpDetails, nil

}

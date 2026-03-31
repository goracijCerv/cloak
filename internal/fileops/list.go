package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupInfo holds the summary information of a backup
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

// ListBackups search all the backups releted to an origin directory
func ListBackups(repoDir string) ([]BackupInfo, error) {
	repoName := filepath.Base(repoDir)
	entries, backupBaseDir, err := readBackUpEntries(repoDir)
	if err != nil {
		return nil, err
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

// reads the manifest file and then returns and backup details
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

// returns a list of the PATHS of the backups that are related to the working directory
func AllBackUpsPaths(after string, before string) ([]string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repoDir := filepath.Clean(currentDir)
	repoName := filepath.Base(repoDir)

	entries, backupBaseDir, err := readBackUpEntries(repoDir)
	if err != nil {
		return nil, err
	}
	//geting the targets
	var targetDateAfter, targetDateBefore time.Time
	if after != "" {
		targetDateAfter, err = time.Parse("2006-01-02", after)
		if err != nil {
			return nil, ErrDateNotValidFormat
		}
	}

	if before != "" {
		targetDateBefore, err = time.Parse("2006-01-02", before)
		if err != nil {
			return nil, ErrDateNotValidFormat
		}
	}

	prefix := fmt.Sprintf("[%s]", repoName)
	var backUpsPaths []string

	for _, entry := range entries {
		//we only search for the folders that start with the prefix
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backupPath := filepath.Join(backupBaseDir, entry.Name())
			manifest, err := readManifest(backupPath)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrGetManifest, err)
			}

			switch {
			case after != "" && before != "":
				if manifest.CreatedAt.After(targetDateAfter) && manifest.CreatedAt.Before(targetDateBefore) {
					backUpsPaths = append(backUpsPaths, backupPath)
				}

			case after != "" && before == "":
				if manifest.CreatedAt.After(targetDateAfter) {
					backUpsPaths = append(backUpsPaths, backupPath)
				}

			case after == "" && before != "":
				if manifest.CreatedAt.Before(targetDateBefore) {
					backUpsPaths = append(backUpsPaths, backupPath)
				}

			default:

				backUpsPaths = append(backUpsPaths, backupPath)

			}

		}
	}

	if len(backUpsPaths) == 0 {
		return nil, ErrNoBackUps
	}

	return backUpsPaths, nil

}

func readBackUpEntries(repoDir string) ([]os.DirEntry, string, error) {

	parentDir := filepath.Dir(repoDir)
	backupBaseDir := filepath.Join(parentDir, "backup")

	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrNoBackUpDir //there are no backups
		}
		return nil, "", fmt.Errorf("%w: %w", ErrFailedBackDir, err)
	}

	return entries, backupBaseDir, nil
}

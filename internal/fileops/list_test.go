package fileops

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Creates an fake backup with an specific date in the manifest.json
func createBackupWithDate(t *testing.T, backupBaseDir, repoName, folderSuffix, dateStr string) string {
	folderName := "[" + repoName + "]" + folderSuffix
	bakupPath := filepath.Join(backupBaseDir, folderName)
	if err := os.MkdirAll(bakupPath, os.ModePerm); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	parseDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		t.Fatalf("failed to parse date for test: %v", err)
	}

	manifest := Manifest{
		CreatedAt: parseDate,
		Entries: []ManifestEntry{
			{BackupName: "file1.txt", OriginalPath: "file1.txt"},
		},
	}

	data, _ := json.Marshal(manifest)
	manifestPath := filepath.Join(bakupPath, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	return bakupPath
}

func TestListBackups(t *testing.T) {

	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "my_project")
	backupDir := filepath.Join(tempDir, "backup")
	os.MkdirAll(projectDir, os.ModePerm)
	os.MkdirAll(backupDir, os.ModePerm)

	t.Run("Without  backups", func(t *testing.T) {
		_, err := ListBackups(projectDir)
		if !errors.Is(err, ErrNoBackUps) {
			t.Errorf("wanted ErrNoBackUps, got %v", err)
		}
	})

	t.Run("Filter and Sorting", func(t *testing.T) {
		createBackupWithDate(t, backupDir, "my_project", "old", "2024-01-01T10:00:00Z")
		createBackupWithDate(t, backupDir, "my_project", "new", "2026-01-01T10:00:00Z")
		createBackupWithDate(t, backupDir, "my_project", "middle", "2025-01-01T10:00:00Z")

		createBackupWithDate(t, backupDir, "other_project", "x", "2026-01-01T10:00:00Z")

		backups, err := ListBackups(projectDir)
		if err != nil {
			t.Fatalf("unwanted error: %v", err)
		}

		if len(backups) != 3 {
			t.Fatalf("wanted 3 backups, got %d", len(backups))
		}

		if backups[0].CreatedAt.Year() != 2026 {
			t.Errorf("the sorting failed. the first element should the 2026 but it is %d", backups[0].CreatedAt.Year())
		}

		if backups[2].CreatedAt.Year() != 2024 {
			t.Errorf("the sorting failed. the last element should the 2024 but it is %d", backups[2].CreatedAt.Year())

		}
	})

}

func TestAllBackUpsPaths_DateFilters(t *testing.T) {
	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "my_project")
	backupDir := filepath.Join(tempDir, "backup")
	os.MkdirAll(projectDir, os.ModePerm)
	os.MkdirAll(backupDir, os.ModePerm)

	originalWd, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(originalWd)

	createBackupWithDate(t, backupDir, "my_project", "enero", "2026-01-15T10:00:00Z")
	createBackupWithDate(t, backupDir, "my_project", "febrero", "2026-02-15T10:00:00Z")
	createBackupWithDate(t, backupDir, "my_project", "marzo", "2026-03-15T10:00:00Z")

	t.Run("Filtering --after (after janaury)", func(t *testing.T) {
		paths, err := AllBackUpsPaths("2026-01-31", "")
		if err != nil {
			t.Fatalf("unwanted error: %v", err)
		}
		if len(paths) != 2 {
			t.Errorf("wanted 2 backups (february and march), got %d", len(paths))
		}
	})

	t.Run("Filtering --before (before march)", func(t *testing.T) {
		paths, err := AllBackUpsPaths("", "2026-03-01")
		if err != nil {
			t.Fatalf("unwanted error: %v", err)
		}
		if len(paths) != 2 {
			t.Errorf("wanted 2 backups (janaury and february), got %d", len(paths))
		}
	})

	t.Run("Filtering --after and --before (only febraury)", func(t *testing.T) {
		paths, err := AllBackUpsPaths("2026-02-01", "2026-02-28")
		if err != nil {
			t.Fatalf("unwanted error: %v", err)
		}
		if len(paths) != 1 {
			t.Errorf("wanted 1 backup (february), got %d", len(paths))
		}
	})
}

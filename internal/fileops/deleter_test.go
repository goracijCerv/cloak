package fileops

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func createFakeBackup(t *testing.T, baseDir, name string) string {
	backupPath := filepath.Join(baseDir, name)
	if err := os.MkdirAll(backupPath, os.ModePerm); err != nil {
		t.Fatalf("failed to create fake backup dir: %v", err)
	}

	manifestPath := filepath.Join(backupPath, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"entries":[]}`), 0644); err != nil {
		t.Fatalf("failed to create fake manifest: %v", err)
	}

	return backupPath
}

func TestValidatePath(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Empty path", func(t *testing.T) {
		if err := validatePath(""); !errors.Is(err, ErrNoPaths) {
			t.Errorf("wanted ErrNoPaths, got %v", err)
		}
	})

	t.Run("Relative path", func(t *testing.T) {
		if err := validatePath("a/relative/path"); !errors.Is(err, ErrNoAbsolute) {
			t.Errorf("wanted ErrNoAbsolute, got %v", err)
		}
	})

	t.Run("Path does not exist", func(t *testing.T) {
		fakeAbs := filepath.Join(tempDir, "i_dont_exist")
		if err := validatePath(fakeAbs); !errors.Is(err, ErrNoPathExist) {
			t.Errorf("wanted ErrNoPathExist, got %v", err)
		}
	})

	t.Run("It exist but is not a backup (without an manifest.json)", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "normal_folder")
		os.MkdirAll(emptyDir, os.ModePerm)

		if err := validatePath(emptyDir); !errors.Is(err, ErrGetManifest) {
			t.Errorf("wanted ErrGetManifest, got %v", err)
		}
	})

	t.Run("Valid path", func(t *testing.T) {
		validBackUp := createFakeBackup(t, tempDir, "valid_backup")
		if err := validatePath(validBackUp); err != nil {
			t.Errorf("no error wanted, got %v", err)
		}
	})
}

func TestDeleteByPaths(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Fail-Fast: if an invalid path is given it protect", func(t *testing.T) {
		valid1 := createFakeBackup(t, tempDir, "failfast_1")
		valid2 := createFakeBackup(t, tempDir, "failfast_2")
		invalidPath := filepath.Join(tempDir, "ghost")

		err := DeleteByPaths(valid1, invalidPath, valid2)

		if err == nil {
			t.Errorf("wanted an error because of the invalid path, got nil")
		}

		if _, err := os.Stat(valid1); os.IsNotExist(err) {
			t.Errorf("Fail-Fast falied: valid1 got deleted by accident")
		}
		if _, err := os.Stat(valid2); os.IsNotExist(err) {
			t.Errorf("Fail-Fast falied: valid2 got deleted by accident")
		}
	})

	t.Run("successful delete of valid paths", func(t *testing.T) {
		valid1 := createFakeBackup(t, tempDir, "success_1")
		valid2 := createFakeBackup(t, tempDir, "success_2")

		err := DeleteByPaths(valid1, valid2)
		if err != nil {
			t.Fatalf("unespected error: %v", err)
		}

		if _, err := os.Stat(valid1); !os.IsNotExist(err) {
			t.Errorf("valid1 didnt get delete")
		}
		if _, err := os.Stat(valid2); !os.IsNotExist(err) {
			t.Errorf("valid2 didnt get delete")
		}
	})
}

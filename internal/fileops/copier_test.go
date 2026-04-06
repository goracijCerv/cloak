package fileops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFolderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "Normal name",
			input:    "my_feature",
			maxLen:   50,
			expected: "my_feature",
		}, {
			name:     "Name with invalid characters",
			input:    "fix/bug:123?",
			maxLen:   50,
			expected: "fix_bug_123_",
		},
		{
			name:     "Windows reserved name",
			input:    "CON",
			maxLen:   50,
			expected: "_CON",
		},
		{
			name:     "Empty name fallback",
			input:    "   ",
			maxLen:   50,
			expected: "BACKUP",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeFolderName(tc.input, tc.maxLen)
			if result != tc.expected {

			}
		})
	}
}

func TestBuildOutPutDir(t *testing.T) {
	temDir := t.TempDir()
	originDir := filepath.Join(temDir, "my_project")

	t.Run("with explicit output dir", func(t *testing.T) {
		customOut := filepath.Join(temDir, "my_backups")
		finalDir, _, err := BuildOutPutDir(customOut, &originDir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if finalDir != customOut {
			t.Errorf("expected %q, got %q", customOut, finalDir)
		}
	})

	t.Run(" output dir (usinf parent folder)", func(t *testing.T) {
		finalDir, _, err := BuildOutPutDir("", &originDir, "test_msg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parentOfFinal := filepath.Dir(finalDir)
		if filepath.Base(parentOfFinal) != "backup" {
			t.Errorf("expected parent directory to be 'backup', got %q", filepath.Base(parentOfFinal))
		}
	})

}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "project")
	destDir := filepath.Join(tempDir, "dest")
	os.MkdirAll(projectDir, os.ModePerm)
	os.MkdirAll(destDir, os.ModePerm)

	originFile := filepath.Join(projectDir, "test.txt")
	content := []byte("Hola mundo Cloak")
	if err := os.WriteFile(originFile, content, 0644); err != nil {
		t.Fatalf("failed to create origin file: %v", err)
	}

	relPath, finalName, err := copyFile(originFile, destDir, projectDir)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	if relPath != "test.txt" {
		t.Errorf("expected relative path 'test.txt', got %q", relPath)
	}

	copiedFilePath := filepath.Join(destDir, finalName)
	copiedContent, err := os.ReadFile(copiedFilePath)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if string(copiedContent) != string(content) {
		t.Errorf("expected content %q, got %q", string(content), string(copiedContent))
	}
}

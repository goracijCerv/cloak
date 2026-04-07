package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Get the list of files that are untracked, modified, and staged
func GetFiles(dir *string) ([]string, error) {

	if dir == nil || *dir == "" {
		return nil, fmt.Errorf("directory path must not be empty")
	}

	if !filepath.IsAbs(*dir) {
		return nil, fmt.Errorf("directory path must be absolute")
	}

	// Clean removes redundancies like // or ./
	finalPath := filepath.Clean(*dir)

	//Verify that it is a git repository
	if err := checkIsGitRepo(finalPath); err != nil {
		return nil, fmt.Errorf("must be run inside a git repository: %w", err)
	}

	// We use a map to avoid duplicate files (e.g., a file that is staged and modified)
	uniqueFiles := make(map[string]struct{})

	// 1. Staged Files (Modified)
	modifiedStagedFiles, err := getStagedFiles(finalPath, "M")
	if err != nil {
		return nil, fmt.Errorf("failed to get modified staged files: %w", err)
	}
	for _, f := range modifiedStagedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 2. Staged Files (Added/New)
	untrackedStagedFiles, err := getStagedFiles(finalPath, "A")
	if err != nil {
		return nil, fmt.Errorf("failed to get added staged files: %w", err)
	}
	for _, f := range untrackedStagedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 3. Modified Files (Working directory)
	modifiedFiles, err := runGitLsFiles(finalPath, "--modified")
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}
	for _, f := range modifiedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 4. Untracked Files (Ignoring .gitignore)
	untrackedFiles, err := runGitLsFiles(finalPath, "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}
	for _, f := range untrackedFiles {
		uniqueFiles[f] = struct{}{}
	}

	/// Convert the map back to a slice with absolute paths
	var finalFiles []string
	for file := range uniqueFiles {
		filepathComplete := filepath.Join(finalPath, file)
		finalFiles = append(finalFiles, filepath.Clean(filepathComplete))
	}

	return finalFiles, nil
}

// Accepts multiple flags to make the function more flexible
func runGitLsFiles(dir string, flags ...string) ([]string, error) {
	args := append([]string{"ls-files"}, flags...)
	// #nosec G204 -- The git command is safe in this CLI context
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w\noutput: %s", err, string(output))
	}
	return strings.Fields(string(output)), nil
}

func checkIsGitRepo(dir string) error {
	// #nosec G204 -- The git command is safe in this CLI context
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("not a git repository: %s", string(output))
	}
	return nil
}

func getStagedFiles(dir string, filter string) ([]string, error) {
	flag := fmt.Sprintf("--diff-filter=%s", filter)
	// #nosec G204 -- The git command is safe in this CLI context
	cmd := exec.Command("git", "diff", "--name-only", "--cached", flag)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w\noutput: %s", err, string(output))
	}
	return strings.Fields(string(output)), nil
}

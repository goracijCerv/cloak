package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Obtener la lista de archivos que estan untracked, modificados y staged
func GetFiles(dir *string) ([]string, error) {

	if dir == nil || *dir == "" {
		return nil, fmt.Errorf("directory path must not be empty")
	}

	if !filepath.IsAbs(*dir) {
		return nil, fmt.Errorf("directory path must be absolute")
	}

	// Clean elimina redundancias como // o ./
	finalPath := filepath.Clean(*dir)

	// Verificar que sea un repositorio de git
	if err := checkIsGitRepo(finalPath); err != nil {
		return nil, fmt.Errorf("must be run inside a git repository: %w", err)
	}

	fmt.Printf("Searching in: %s\n", finalPath)

	// Usamos un mapa para evitar archivos duplicados (ej. un archivo que está staged Y modificado)
	uniqueFiles := make(map[string]struct{})

	// 1. Archivos Staged (Modificados)
	modifiedStagedFiles, err := getStagedFiles(finalPath, "M")
	if err != nil {
		return nil, fmt.Errorf("failed to get modified staged files: %w", err)
	}
	for _, f := range modifiedStagedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 2. Archivos Staged (Añadidos/Nuevos)
	untrackedStagedFiles, err := getStagedFiles(finalPath, "A")
	if err != nil {
		return nil, fmt.Errorf("failed to get added staged files: %w", err)
	}
	for _, f := range untrackedStagedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 3. Archivos Modificados (Working directory)
	modifiedFiles, err := runGitLsFiles(finalPath, "--modified")
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}
	for _, f := range modifiedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// 4. Archivos Untracked (Ignorando .gitignore)
	untrackedFiles, err := runGitLsFiles(finalPath, "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}
	for _, f := range untrackedFiles {
		uniqueFiles[f] = struct{}{}
	}

	// Convertir el mapa de vuelta a un slice con rutas absolutas
	var finalFiles []string
	for file := range uniqueFiles {
		filepathComplete := filepath.Join(finalPath, file)
		finalFiles = append(finalFiles, filepath.Clean(filepathComplete))
	}

	return finalFiles, nil
}

// Acepta múltiples flags para hacer la función más flexible
func runGitLsFiles(dir string, flags ...string) ([]string, error) {
	args := append([]string{"ls-files"}, flags...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w\noutput: %s", err, string(output))
	}
	return strings.Fields(string(output)), nil
}

func checkIsGitRepo(dir string) error {
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
	cmd := exec.Command("git", "diff", "--name-only", "--cached", flag)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w\noutput: %s", err, string(output))
	}
	return strings.Fields(string(output)), nil
}

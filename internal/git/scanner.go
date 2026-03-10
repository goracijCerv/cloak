package git

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Obtener lo lista de archivos que estan untracked y que han sido modificados
func GetFiles(dir *string) ([]string, error) {

	if dir == nil || *dir == "" {
		log.Fatalln("directory path is empty")
		return nil, fmt.Errorf("directory path must not be empty")
	}

	if !filepath.IsAbs(*dir) {
		log.Fatalln("directory path isnt absolute")
		return nil, fmt.Errorf("directory path must be absolute")
	}

	var finalFiles []string
	var finalPath string
	// Clean elimina redundancias como // o ./
	finalPath = filepath.Clean(*dir)

	fmt.Printf("Searching in: %s\n", finalPath)

	//obteniendo los archivos modificados
	modifiedFiles, err := runGitLsFiles(finalPath, "--modified")
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files")
	}

	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, modifiedFiles...)

	//obteniendo los archivos untracked
	untrackedFiles, err := runGitLsFiles(finalPath, "--others")
	if err != nil {
		return nil, fmt.Errorf("faild to get untrackeds fields")
	}

	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, untrackedFiles...)

	for i := range finalFiles {
		filepathComplete := filepath.Join(finalPath, finalFiles[i])
		finalFiles[i] = filepath.Clean(filepathComplete)
	}

	return finalFiles, nil

}

func runGitLsFiles(dir string, flag string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", flag)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("git ls-files %s failed: %w\noutput: %s", flag, err, string(output))
	}
	return strings.Fields(string(output)), nil
}

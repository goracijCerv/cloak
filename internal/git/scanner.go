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

	//Verificar que sea un repositorio de git
	if err := checkIsGitRepo(*dir); err != nil {
		log.Fatalln("it is not a git repository")
		return nil, fmt.Errorf("must be run inside a git repository")
	}

	var finalFiles []string
	var finalPath string
	// Clean elimina redundancias como // o ./
	finalPath = filepath.Clean(*dir)

	fmt.Printf("Searching in: %s\n", finalPath)

	//Obteniendo los archivos Staged modificados

	modifiedStagFiles, err := getsStaggeFiles(finalPath, "M")
	if err != nil {
		return nil, fmt.Errorf("faild to get modified staged files")
	}

	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, modifiedStagFiles...)

	untrackedStagFiles, err := getsStaggeFiles(finalPath, "A")
	if err != nil {
		return nil, fmt.Errorf("faild to get modified staged files")
	}

	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, untrackedStagFiles...)

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
	var cmd *exec.Cmd
	if flag == "--others" {
		cmd = exec.Command("git", "ls-files", "--exclude-standard", flag)
	} else {
		cmd = exec.Command("git", "ls-files", flag)
	}

	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("git ls-files %s failed: %w\noutput: %s", flag, err, string(output))
	}
	return strings.Fields(string(output)), nil
}

func checkIsGitRepo(dir string) error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
		return err
	}
	return nil
}

func getsStaggeFiles(dir string, flag string) ([]string, error) {
	var cmd *exec.Cmd
	if flag == "M" {
		cmd = exec.Command("git", "diff", "--name-only", "--cached", "--diff-filter=M")
	} else {
		cmd = exec.Command("git", "diff", "--name-only", "--cached", "--diff-filter=A")
	}
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("git diff --name-only %s failed: %w\noutput: %s", flag, err, string(output))
	}
	return strings.Fields(string(output)), nil
}

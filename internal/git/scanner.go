package git

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Obtener lo lista de archivos que estan untracked y que han sido modificados
func GetFiles(dir *string) []string {
	var finalFiles []string
	var finalPath string
	if filepath.IsAbs(*dir) {
		finalPath = *dir
	}

	// Clean elimina redundancias como // o ./
	finalPath = filepath.Clean(finalPath)

	fmt.Printf("Buscando en: %s\n", finalPath)

	//obteniendo los archivos untracked
	cmd := exec.Command("git", "ls-files", "--others")
	cmd.Dir = finalPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
	}
	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, strings.Fields(string(output))...)

	//obteniendo los archivos modificados
	cmd = exec.Command("git", "ls-files", "--modified")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
	}

	//los añadimos a la lista de los archivos pendientes
	finalFiles = append(finalFiles, strings.Fields(string(output))...)
	fmt.Println(finalFiles)

	for i := range finalFiles {
		filepathComplete := filepath.Join(finalPath, finalFiles[i])
		finalFiles[i] = filepath.Clean(filepathComplete)
	}

	return finalFiles

}

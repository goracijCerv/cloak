package main

import (
	"cloak/internal/fileops"
	"cloak/internal/git"
	"flag"
	"fmt"
	"os"
)

func main() {
	actualDirectory, err := os.Getwd()
	if err != nil {
		panic("Something went wrong getting the actual directory")
	}
	gitDirectory := flag.String("d", actualDirectory, "Directorio al que se quiere hacer el backup de los cambios untracked")
	outPutDirectory := flag.String("o", "", "Directorio donde se creara el backup de los cambios untracked")
	messageComentari := flag.String("m", "", "Mensaje adicional para el folder")

	flag.Parse()

	fmt.Println(*gitDirectory)
	fmt.Println(*outPutDirectory)

	// cmd := exec.Command("git", "ls-files", "--others")
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Fatalf("cmd.Run() failed with %s\nOutput: %s", err, string(output))
	// }

	// fmt.Printf("Output: %s\n", string(output))
	filesToCopy := git.GetFiles(gitDirectory)

	fmt.Println(filesToCopy)
	fileops.CreateNewBackUp(filesToCopy, *outPutDirectory, *messageComentari, gitDirectory)
}

package fileops

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const DefaultMaxFolderLength = 50

var (
	invalidChars    = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	multiUnderscore = regexp.MustCompile(`_+`)
)

var windowsReserverd = map[string]struct{}{
	"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {}, "COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
	"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {}, "LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
}

func sanitizeFolderName(name string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = DefaultMaxFolderLength
	}

	name = norm.NFC.String(name)
	clean := invalidChars.ReplaceAllString(name, "_")
	clean = strings.TrimSpace(clean)
	clean = strings.TrimRight(clean, ".")
	clean = multiUnderscore.ReplaceAllString(clean, "_")

	if clean == "" {
		clean = "BACKUP"
	}

	upper := strings.ToUpper(clean)
	if _, reserved := windowsReserverd[upper]; reserved {
		clean = "_" + clean
	}

	if utf8.RuneCountInString(clean) > maxLen {
		runes := []rune(clean)
		clean = string(runes[:maxLen])
	}

	return clean
}

func copiarArchivo(rutaOrigen string, dirDestiono string, raizProyecto string) error {
	//obtenemos el path relativo
	rutaRelativa, err := filepath.Rel(raizProyecto, rutaOrigen)
	if err != nil {
		return fmt.Errorf("failed to compute relative path for %q: %w", rutaOrigen, err)
	}
	//se convierte en _ los / del path por ejemplo cdm/cloack/main a cdm_cloack_main
	nuevoNombre := strings.ReplaceAll(rutaRelativa, string(filepath.Separator), "_")
	rutaDestino := filepath.Join(dirDestiono, nuevoNombre)
	//esto valida que no haya repetidos y si hay creaa otro nombre pero con un numero
	if _, err := os.Stat(rutaDestino); err == nil {
		ext := filepath.Ext(nuevoNombre)
		nombreSinExt := strings.TrimSuffix(nuevoNombre, ext)
		contador := 1

		for {
			rutaDestino = filepath.Join(dirDestiono, fmt.Sprintf("%s_%d%s", nombreSinExt, contador, ext))
			if _, err := os.Stat(rutaDestino); os.IsNotExist(err) {
				break
			}
			contador++
		}
	}

	//Empezamos a hacer la compia
	origen, err := os.Open(rutaOrigen)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", rutaOrigen, err)
	}

	defer origen.Close()

	destino, err := os.Create(rutaDestino)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", rutaDestino, err)
	}

	defer destino.Close()

	if _, err = io.Copy(destino, origen); err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", rutaOrigen, rutaDestino, err)
	}

	return nil

}

// Obtenemos el directorio final del backup
func BuildOutPutDir(outPutDir string, dirOrigen *string, message string) (string, error) {
	if outPutDir != "" {
		return filepath.Clean(outPutDir), nil
	}

	parentDir := filepath.Dir(*dirOrigen)
	if parentDir == "." {
		return "", fmt.Errorf("source directory has no parent directory")
	}

	folderName := filepath.Base(*dirOrigen)

	currentTime := time.Now()
	timestamp := fmt.Sprintf("%d-%02d-%02d_%02d-%02d-%02d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second())

	var backupFolderName string

	if message != "" {
		safeMessage := sanitizeFolderName(message, 0)
		backupFolderName = fmt.Sprintf("[%s][%s]-%s", folderName, safeMessage, timestamp)
	} else {
		backupFolderName = fmt.Sprintf("[%s]%s", folderName, timestamp)
	}
	return filepath.Clean(filepath.Join(parentDir, "backup", backupFolderName)), nil
}

func CreateNewBackUp(files []string, outPutDir string, messagge string, dirOrigen *string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided to back up")
	}

	finalOutPutDir, err := BuildOutPutDir(outPutDir, dirOrigen, messagge)
	if err != nil {
		return fmt.Errorf("failed to resolve output directory: %w", err)
	}

	fmt.Println("Backup destination:", finalOutPutDir)

	//Crear el direcotrio
	if _, err := os.Stat(finalOutPutDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(finalOutPutDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create backup directory %q: %w", finalOutPutDir, err)
		}
	}

	//copiar archivos
	var copyErrors []string
	for i := range files {
		if err := copiarArchivo(files[i], finalOutPutDir, *dirOrigen); err != nil {
			copyErrors = append(copyErrors, err.Error())

		}
	}
	if len(copyErrors) > 0 {
		return fmt.Errorf("backup completed with %d error(s): \n%s", len(copyErrors), strings.Join(copyErrors, "\n"))
	}

	return nil
}

// Obtiene las rutas destio al repositorio original, basado en los archivos del backupdir
func getDestinyRoutes(backupDir string, originalDir string) ([]string, error) {
	if backupDir == "" {
		return nil, fmt.Errorf("no backup folder")
	}

	info, err := os.Stat(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("route does not exist")
		}
		return nil, fmt.Errorf("something weird happen %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("route is not an folder")
	}
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong: %w", err)
	}
	var finalRoutes []string
	for in := range files {
		baseRute := strings.ReplaceAll(files[in].Name(), "_", string(filepath.Separator))
		rutaDestino := filepath.Clean(filepath.Join(originalDir, baseRute))
		finalRoutes = append(finalRoutes, rutaDestino)
	}

	// for i := range finalRoutes {
	// 	fmt.Println(finalRoutes[i])
	// }
	return finalRoutes, nil
}

func getFilesRoutesFromBackUp(backupDir string) ([]string, error) {
	if backupDir == "" {
		return nil, fmt.Errorf("empty path")
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong: %w", err)
	}

	var paths []string
	for _, i := range files {
		fullpath := filepath.Clean(filepath.Join(backupDir, i.Name()))
		paths = append(paths, fullpath)
	}

	return paths, nil
}

func restoreFile(fileToRestore string, destinyPath string) error {
	if fileToRestore == "" || destinyPath == "" {
		return fmt.Errorf("empty path")
	}

	//Verificamos que exista el directorio
	dir := filepath.Dir(destinyPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories %q: %w", dir, err)
	}

	//Empezamos a hacer la compia
	origen, err := os.Open(fileToRestore)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", fileToRestore, err)
	}

	defer origen.Close()

	destino, err := os.Create(destinyPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", destinyPath, err)
	}

	defer destino.Close()

	if _, err = io.Copy(destino, origen); err != nil {
		return fmt.Errorf("failed to restore %q to %q: %w", fileToRestore, destinyPath, err)
	}

	return nil
}

func RestorBackUp(backupDir string, originalDir string) error {
	if backupDir == "" || originalDir == "" {
		return fmt.Errorf("empty path")
	}

	//opteniendo las rutas de los archvios a copiar del backup
	backUpRoutes, err := getFilesRoutesFromBackUp(backupDir)
	if err != nil {
		return err
	}

	if len(backUpRoutes) == 0 {
		return fmt.Errorf("no backup files to restore")
	}

	//obteniendo las rutas destino de la carpeta orignal
	destinyRoutes, err := getDestinyRoutes(backupDir, originalDir)

	if err != nil {
		return err
	}

	if len(destinyRoutes) == 0 {
		return fmt.Errorf("something went weong getting the destiny routes")
	}

	//copiar archivos
	var copyErrors []string
	for i := range backUpRoutes {
		if err := restoreFile(backUpRoutes[i], destinyRoutes[i]); err != nil {
			copyErrors = append(copyErrors, err.Error())

		}
	}
	if len(copyErrors) > 0 {
		return fmt.Errorf("restore completed with %d error(s): \n%s", len(copyErrors), strings.Join(copyErrors, "\n"))
	}

	return nil
}

//Estructura de la direccion que debe hacer
/*
 lugarDondeEstaRepo/backup/MensajeOpcional*[NombreRepo]Fecha

 lugarDondeEstaRepo/backup/[NombreRepo]MensajeOpciona-Fecha



  func copiarInteligente(rutaOrigen string, dirDestino string, dirOrigen *string) error {
	// 1. Determine the path we will use strictly for naming purposes
	rutaParaNombres := rutaOrigen
	if *dirOrigen != "" {
		// Strip the C:\Users\IT\go\src\cloak part away
		if rel, err := filepath.Rel(*dirOrigen, rutaOrigen); err == nil {
			rutaParaNombres = rel
		}
	}

	nombreArchivo := filepath.Base(rutaParaNombres)
	rutaDestino := filepath.Join(dirDestino, nombreArchivo)

	// 2. Verify if the file exists and resolve collisions
	if _, err := os.Stat(rutaDestino); err == nil {
		ext := filepath.Ext(nombreArchivo)
		nombreSinExt := strings.TrimSuffix(nombreArchivo, ext)

		// Start looking at the parent folders of our relative path
		currentDir := filepath.Dir(rutaParaNombres)
		var prefix string
		nombreEncontrado := false

		// Walk up the directory tree
		for {
			// Stop if we reach the relative root (".") or filesystem root
			if currentDir == "." || currentDir == string(filepath.Separator) || currentDir == "" {
				break
			}

			baseDir := filepath.Base(currentDir)

			// Build the prefix dynamically (e.g., "utils", then "cmd_utils")
			if prefix == "" {
				prefix = baseDir
			} else {
				prefix = baseDir + "_" + prefix
			}

			nuevoNombre := fmt.Sprintf("%s_%s%s", prefix, nombreSinExt, ext)
			rutaDestino = filepath.Join(dirDestino, nuevoNombre)

			// If it doesn't exist, we found our unique name!
			if _, err := os.Stat(rutaDestino); os.IsNotExist(err) {
				nombreEncontrado = true
				break
			}

			// Move up one directory level for the next iteration
			nextDir := filepath.Dir(currentDir)
			if nextDir == currentDir {
				break // Failsafe
			}
			currentDir = nextDir
		}

		// FALLBACK: If we ran out of folders in our project root and STILL have a collision
		if !nombreEncontrado {
			contador := 1
			for {
				// If prefix is empty (it was at the root of the project), handle formatting safely
				if prefix == "" {
					rutaDestino = filepath.Join(dirDestino, fmt.Sprintf("%s_%d%s", nombreSinExt, contador, ext))
				} else {
					rutaDestino = filepath.Join(dirDestino, fmt.Sprintf("%s_%s_%d%s", prefix, nombreSinExt, contador, ext))
				}

				if _, err := os.Stat(rutaDestino); os.IsNotExist(err) {
					break
				}
				contador++
			}
		}
	}

	// 3. Perform the copy using the ORIGINAL absolute path
	origen, err := os.Open(rutaOrigen)
	if err != nil {
		return err
	}
	defer origen.Close()

	destino, err := os.Create(rutaDestino)
	if err != nil {
		return err
	}
	defer destino.Close()

	_, err = io.Copy(destino, origen)
	return err
}

*/

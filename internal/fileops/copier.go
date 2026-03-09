package fileops

import (
	"errors"
	"fmt"
	"io"
	"log"
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
		return err
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

func CreateNewBackUp(files []string, outPutDir string, messagge string, dirOrigen *string) {

	var finalOutPutDir string
	if outPutDir != "" {
		finalOutPutDir = filepath.Clean(outPutDir)
	} else {
		parentDirectory := filepath.Dir(*dirOrigen)
		folderName := filepath.Base(*dirOrigen)
		if parentDirectory == "." {
			log.Fatalln("No tiene Folder padre")
		}
		if err := os.Chdir(parentDirectory); err != nil {
			log.Fatal("Error changing directory")

		}

		dirPadre, err := os.Getwd()
		if err != nil {
			log.Fatal("Error al obtener el directorio")
		}

		fmt.Println(dirPadre)

		//obtener fecha
		currentTime := time.Now()
		fecha := fmt.Sprintf("%d-%d-%d_%d-%d-%d",
			currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second())

		if messagge != "" {
			messagge = sanitizeFolderName(messagge, 0)
			finalOutPutDir = fmt.Sprintf("/backup/[%s][%s]-%s", folderName, messagge, fecha)
		} else {
			finalOutPutDir = fmt.Sprintf("/backup/[%s]%s", folderName, fecha)
		}
		finalOutPutDir = filepath.Join(dirPadre, finalOutPutDir)
		finalOutPutDir = filepath.Clean(finalOutPutDir)
	}

	fmt.Println(finalOutPutDir)

	//Crear el direcotrio
	if _, err := os.Stat(finalOutPutDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(finalOutPutDir, os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}

	//copiar archivos
	for i := range files {
		err := copiarArchivo(files[i], finalOutPutDir, *dirOrigen)
		if err != nil {
			log.Fatalf("Error al tratar de copiar el archivo %s en el directorio %s\n", files[i], finalOutPutDir)
		}
	}

	log.Println("Archivos copiados")
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

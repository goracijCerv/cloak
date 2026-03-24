package fileops

import (
	"encoding/json"
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

type Manifest struct {
	Warning string          `json:"_WARNING_"`
	Entries []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	BackupName   string `json:"backup_name"`
	OriginalPath string `json:"original_path"`
}

var (
	invalidChars    = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	multiUnderscore = regexp.MustCompile(`_+`)
)

var (
	ErrNoFiles           = errors.New("no files provided to back up")
	ErrResolveOutputDir  = errors.New("failed to resolve output directory")
	ErrBackupWithErrors  = errors.New("backup completed with")
	ErrFaildedBackDir    = errors.New("failed to create backup directory")
	ErrFailedManData     = errors.New("failed to generated manifest data")
	ErrFailedManFile     = errors.New("failed to generated manifest file")
	ErrGetManifest       = errors.New("failed to get manifest file")
	ErrEmptyManifest     = errors.New("manifest file is empty")
	ErrRestoreWithErrors = errors.New("restore completed with")
)

var windowsReserved = map[string]struct{}{
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
	if _, reserved := windowsReserved[upper]; reserved {
		clean = "_" + clean
	}

	if utf8.RuneCountInString(clean) > maxLen {
		runes := []rune(clean)
		clean = string(runes[:maxLen])
	}

	return clean
}

func copyFile(originPath string, destDir string, rootProject string) (string, string, error) {
	//getting the relative path
	relativePath, err := filepath.Rel(rootProject, originPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to compute relative path for %q: %w", originPath, err)
	}
	//converting the / into _ of the path for example cmd/cloak/main to cmd_cloak_main

	newName := strings.ReplaceAll(relativePath, string(filepath.Separator), "_")
	destPath := filepath.Join(destDir, newName)
	//this validate that there arent repeated files and in case that there are we create another one but with a number
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(newName)
		nameNoExt := strings.TrimSuffix(newName, ext)
		counter := 1

		for {
			destPath = filepath.Join(destDir, fmt.Sprintf("%s_%d%s", nameNoExt, counter, ext))
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break
			}
			counter++
		}
	}

	//Starting to make the copy
	origin, err := os.Open(originPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open source file %q: %w", originPath, err)
	}

	defer origin.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file %q: %w", destPath, err)
	}

	defer dest.Close()

	if _, err = io.Copy(dest, origin); err != nil {
		return "", "", fmt.Errorf("failed to copy %q to %q: %w", originPath, destPath, err)
	}

	return filepath.Clean(relativePath), filepath.Base(destPath), nil

}

// we get the final directory of the backup
func BuildOutPutDir(outPutDir string, originDir *string, message string) (string, error) {
	if outPutDir != "" {
		return filepath.Clean(outPutDir), nil
	}

	parentDir := filepath.Dir(*originDir)
	if parentDir == "." {
		return "", fmt.Errorf("source directory has no parent directory")
	}

	folderName := filepath.Base(*originDir)

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

func CreateNewBackUp(files []string, outPutDir string, message string, originDir *string) error {
	if len(files) == 0 {
		return ErrNoFiles
	}

	finalOutPutDir, err := BuildOutPutDir(outPutDir, originDir, message)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResolveOutputDir, err)
	}

	fmt.Println("Backup destination:", finalOutPutDir)

	//Creating directory
	if _, err := os.Stat(finalOutPutDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(finalOutPutDir, os.ModePerm); err != nil {
			return fmt.Errorf("%w %q: %w", ErrFaildedBackDir, finalOutPutDir, err)
		}
	}

	//copy the files
	var copyErrors []string
	manifestFile := Manifest{
		Warning: "DO NOT DELETE OR MODIFY THIS FILE and ALSO DO NOT CHANGE ITS PATH. It is strictly necessary for 'cloak restore' to work.",
	}
	for i := range files {
		relativePathOri, backupFileName, err := copyFile(files[i], finalOutPutDir, *originDir)
		if err != nil {
			copyErrors = append(copyErrors, err.Error())

		} else {
			entry := ManifestEntry{
				BackupName:   backupFileName,
				OriginalPath: relativePathOri,
			}
			manifestFile.Entries = append(manifestFile.Entries, entry)
		}

	}
	//Creating the manifest file
	jsonData, err := json.MarshalIndent(manifestFile, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedManData, err)
	}

	manifestPath := filepath.Join(finalOutPutDir, "manifest.json")
	err = os.WriteFile(manifestPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("%w :%w", ErrFailedManFile, err)
	}

	if len(copyErrors) > 0 {
		return fmt.Errorf("%w %d error(s): \n%s", ErrBackupWithErrors, len(copyErrors), strings.Join(copyErrors, "\n"))
	}

	return nil
}

// read the manifest.json file
func readManifest(backupDir string) (*Manifest, error) {
	manifestPath := filepath.Join(backupDir, "manifest.json")
	file, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest.json was not found in the backup directory")
		}
		return nil, fmt.Errorf("error trying to open the manifest.json file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error trying to read the manifest.json file: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		return nil, fmt.Errorf("the manifest.json file is dameged or it was modified: %w", err)
	}

	return &manifest, nil
}

func restoreFile(fileToRestore string, destinyPath string) error {
	if fileToRestore == "" || destinyPath == "" {
		return fmt.Errorf("empty path")
	}

	//we check that the directory exist
	dir := filepath.Dir(destinyPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories %q: %w", dir, err)
	}

	//making the copy
	origin, err := os.Open(fileToRestore)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", fileToRestore, err)
	}

	defer origin.Close()

	dest, err := os.Create(destinyPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", destinyPath, err)
	}

	defer dest.Close()

	if _, err = io.Copy(dest, origin); err != nil {
		return fmt.Errorf("failed to restore %q to %q: %w", fileToRestore, destinyPath, err)
	}

	return nil
}

func RestoreBackUp(backupDir string, originalDir string) error {

	manifest, err := readManifest(backupDir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetManifest, err)
	}
	if len(manifest.Entries) == 0 {
		return ErrEmptyManifest
	}
	//copying the files
	var copyErrors []string
	for _, entry := range manifest.Entries {
		//makeing the absolute path for the file to restore
		fileToRestore := filepath.Join(backupDir, entry.BackupName)
		//construimos la ruta  absoluta de la ruta que debe regresar en el proyecto original
		destinyPath := filepath.Join(originalDir, entry.OriginalPath)

		if err := restoreFile(fileToRestore, destinyPath); err != nil {
			copyErrors = append(copyErrors, fmt.Sprintf("fail in %s: %s", entry.OriginalPath, err.Error()))
		}
	}

	if len(copyErrors) > 0 {
		return fmt.Errorf("%w %d error(s): \n%s", ErrRestoreWithErrors, len(copyErrors), strings.Join(copyErrors, "\n"))
	}

	return nil
}

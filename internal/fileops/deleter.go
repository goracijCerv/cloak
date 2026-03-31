package fileops

import (
	"fmt"
	"os"
	"path/filepath"
)

func removeDirectory(pathVal string) error {
	return os.RemoveAll(pathVal)
}

func validatePath(pathVal string) error {
	//check if path is not empty
	if pathVal == "" {
		return ErrNoPaths
	}

	//check if the given path is absolute
	if !filepath.IsAbs(pathVal) {
		return ErrNoAbsolute
	}

	//check if the given path exist
	_, err := os.Stat(pathVal)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNoPathExist
		}
		return fmt.Errorf("failed to check the backup directory %s: %w", pathVal, err)
	}

	//check if it is an backup cloak directory
	_, err = readManifest(pathVal)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetManifest, err)
	}

	return nil

}

func DeleteByPaths(pathVal ...string) error {
	for _, v := range pathVal {
		if err := validatePath(v); err != nil {
			return err
		}
	}

	for _, v := range pathVal {
		if err := removeDirectory(v); err != nil {
			return fmt.Errorf("%w: %w", ErrFaildedDelete, err)
		}
	}

	return nil
}

package fileops

import (
	"fmt"
	"os"
	"path/filepath"
)

func removeDirectory(pathVal string) error {
	err := os.RemoveAll(pathVal)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func validatePath(pathVal string) (bool, error) {
	//check if path is not empty
	if pathVal == "" {
		return false, ErrNoPaths
	}

	//check if the given path is absolute
	if !filepath.IsAbs(pathVal) {
		return false, ErrNoAbsolute
	}

	//check if the given path exist
	_, err := os.Stat(pathVal)
	if err != nil {
		if os.IsNotExist(err) {
			return false, ErrNoPathExist
		}
		return false, fmt.Errorf("failed to check the backup directory %s: %w", pathVal, err)
	}

	return true, nil

}

func DeleteByPaths(pathVal ...string) error {
	for _, v := range pathVal {
		if ok, err := validatePath(v); ok == false && err != nil {
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

// func DeleteAll() error {
// 	backUpPaths, err := allBackUpsPaths("", "")
// 	if err != nil {
// 		return err
// 	}

// 	if err := DeleteByPaths(backUpPaths...); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func DeleteFilterdDates(after string, before string) error {
// 	backUpPaths, err := allBackUpsPaths(after, before)
// 	if err != nil {
// 		return err
// 	}

// 	if err := DeleteByPaths(backUpPaths...); err != nil {
// 		return err
// 	}

// 	return nil
// }

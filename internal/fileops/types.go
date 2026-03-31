package fileops

import (
	"errors"
	"time"
)

type Manifest struct {
	Warning   string          `json:"_WARNING_"`
	CreatedAt time.Time       `json:"createdAt"`
	Entries   []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	BackupName   string `json:"backup_name"`
	OriginalPath string `json:"original_path"`
}

var (
	ErrNoFiles            = errors.New("no files provided to back up")
	ErrBackupWithErrors   = errors.New("backup completed with")
	ErrFaildedBackDir     = errors.New("failed to create backup directory")
	ErrFailedManData      = errors.New("failed to generated manifest data")
	ErrFailedManFile      = errors.New("failed to generated manifest file")
	ErrGetManifest        = errors.New("failed to get manifest file")
	ErrEmptyManifest      = errors.New("manifest file is empty")
	ErrRestoreWithErrors  = errors.New("restore completed with")
	ErrNoPaths            = errors.New("paths are empty")
	ErrFailedBackDir      = errors.New("failed to read the backup directory")
	ErrNoBackUps          = errors.New("there are no backups")
	ErrNoBackUpDir        = errors.New("there are no backup folder")
	ErrNoAbsolute         = errors.New("the path is not absolute")
	ErrNoPathExist        = errors.New("path doesn't exist")
	ErrFaildedDelete      = errors.New("failed to delete directory")
	ErrDateNotValidFormat = errors.New("date does not have a valid format")
)

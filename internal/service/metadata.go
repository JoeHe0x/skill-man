package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const metadataFileName = ".skill-man.json"

type installMetadata struct {
	Name        string    `json:"name"`
	SourceKind  string    `json:"source_kind"`
	SourcePath  string    `json:"source_path"`
	InstalledAt time.Time `json:"installed_at"`
}

func metadataPathForDir(dir string) string {
	return filepath.Join(dir, metadataFileName)
}

func readInstallMetadata(dir string) (installMetadata, bool, error) {
	path := metadataPathForDir(dir)
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return installMetadata{}, false, nil
		}
		return installMetadata{}, false, err
	}

	var meta installMetadata
	if err := json.Unmarshal(content, &meta); err != nil {
		return installMetadata{}, false, err
	}
	return meta, true, nil
}

func writeInstallMetadata(dir string, meta installMetadata) error {
	content, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(metadataPathForDir(dir), content, 0o644)
}

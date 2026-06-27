package sbufs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func NormalizePath(path string) string {
	volume := filepath.VolumeName(path)
	withoutVolume := strings.TrimPrefix(path, volume)
	cleanedPath := filepath.Clean(withoutVolume)
	sepSlashPath := filepath.ToSlash(cleanedPath)
	return strings.TrimPrefix(sepSlashPath, "/")
}

func GetListFiles(src string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error reading file %s: %w", path, err)
		}

		if !d.IsDir() {
			files = append(files, path)
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk dir: %w", err)
	}

	return files, nil
}

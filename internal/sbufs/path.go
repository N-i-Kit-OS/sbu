package sbufs

import (
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

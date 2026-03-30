package server

import "os"

// isSymlink checks if a path is a symlink using Lstat.
// Used on all platforms as the common fallback/pre-check.
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

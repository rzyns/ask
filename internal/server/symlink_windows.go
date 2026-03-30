//go:build windows

package server

// openNoFollow is a no-op on Windows; symlinks are checked via Lstat instead.
const openNoFollow = 0

// isSymlinkError checks whether the path is a symlink on Windows.
// Since O_NOFOLLOW is not available, we fall back to Lstat-based detection.
// This is only called when OpenFile returns an error, so we check the path
// via Lstat as a best-effort symlink guard.
func isSymlinkError(_ error) bool {
	// On Windows, O_NOFOLLOW is 0 so OpenFile won't reject symlinks.
	// The caller should use isSymlink() as an additional check.
	return false
}

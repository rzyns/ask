//go:build !windows

package server

import (
	"errors"
	"syscall"
)

// openNoFollow is the flag to prevent following symlinks on open.
const openNoFollow = syscall.O_NOFOLLOW

// isSymlinkError reports whether an error from OpenFile with O_NOFOLLOW
// indicates the target was a symlink.
func isSymlinkError(err error) bool {
	return errors.Is(err, syscall.ELOOP)
}

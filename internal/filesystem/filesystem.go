// Package filesystem provides utility functions for file system operations.
package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const maxCopyDepth = 20

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Uses Lstat on the source root to reject symlinks and prevent following links
// to attacker-controlled locations.
func CopyDir(source string, destination string) error {
	return copyDirRecursive(source, destination, 0)
}

func copyDirRecursive(source, destination string, depth int) error {
	if depth > maxCopyDepth {
		return fmt.Errorf("copy directory recursion limit reached (max depth %d)", maxCopyDepth)
	}

	srcInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source is a symlink: rejecting for safety")
	}

	if err := os.MkdirAll(destination, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip symlinks to prevent following links outside intended directory
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		srcPath := filepath.Join(source, entry.Name())
		dstPath := filepath.Join(destination, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath, depth+1); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies a single file from src to dst.
// Rejects symlinks as source to prevent symlink-based attacks.
func CopyFile(src, dst string) (retErr error) {
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source is a symlink: rejecting for safety")
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := destFile.Close(); retErr == nil {
			retErr = cerr
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Preserve permissions but strip setuid, setgid, sticky bits and
	// group/other write bits for security.
	// Execute bits are preserved so shell scripts remain runnable.
	info, err := os.Stat(src)
	if err == nil {
		mode := info.Mode() & 0755 // Keep owner rwx, group/other r-x only
		_ = os.Chmod(dst, mode)
	}

	return nil
}

// CreateSymlinkOrCopy creates a symlink from target to source, or falls back to copy on failure.
// Uses relative paths for portability. Works on Linux, macOS, and Windows (with fallback).
func CreateSymlinkOrCopy(source, target string) error {
	// Calculate relative path for portability
	// The link is at 'target' pointing to 'source'
	// We need 'source' relative to 'target's directory
	targetDir := filepath.Dir(target)
	relSource, err := filepath.Rel(targetDir, source)
	if err != nil {
		relSource = source // Fallback to absolute if rel fails
	}

	// Debug print not available here without dependency cycle or logger injection
	// Using generic symlink creation

	err = os.Symlink(relSource, target)
	if err == nil {
		return nil
	}

	// Determine if we should fallback to copy
	// On Windows, symlinks require special permissions.
	// We can try a junction or just copy.
	if runtime.GOOS == "windows" {
		// Try deep copy
		fi, err := os.Stat(source)
		if err == nil && fi.IsDir() {
			return CopyDir(source, target)
		} else if err == nil {
			return CopyFile(source, target)
		}
	}

	// Fallback for other OS or file types
	// If symlink failed (permission denied?), try copy
	fi, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", source, err)
	}
	if fi.IsDir() {
		return CopyDir(source, target)
	}
	return CopyFile(source, target)
}

// IsSymlink checks if the given path is a symbolic link
func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

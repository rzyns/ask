package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	symlinkFile := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(regularFile, symlinkFile); err != nil {
		t.Fatal(err)
	}

	// Create a directory
	dir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to a directory
	symlinkDir := filepath.Join(tmpDir, "linkdir")
	if err := os.Symlink(dir, symlinkDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"regular file is not symlink", regularFile, false},
		{"symlink to file is symlink", symlinkFile, true},
		{"directory is not symlink", dir, false},
		{"symlink to dir is symlink", symlinkDir, true},
		{"nonexistent path is not symlink", filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSymlink(tt.path)
			if got != tt.expected {
				t.Errorf("isSymlink(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestIsSymlink_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a symlink pointing to a non-existent target (broken symlink)
	brokenLink := filepath.Join(tmpDir, "broken")
	if err := os.Symlink(filepath.Join(tmpDir, "nonexistent-target"), brokenLink); err != nil {
		t.Fatal(err)
	}

	// isSymlink should detect broken symlinks too since it uses Lstat
	if !isSymlink(brokenLink) {
		t.Error("isSymlink should return true for broken symlinks")
	}
}

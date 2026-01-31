package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyFile(t *testing.T) {
	// Create source file
	srcFile, err := os.CreateTemp("", "test-src-*")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(srcFile.Name()) }()

	content := []byte("hello world")
	_, err = srcFile.Write(content)
	assert.NoError(t, err)
	_ = srcFile.Close()

	// content dest
	dstDir, err := os.MkdirTemp("", "test-dst-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(dstDir) }()

	dstFile := filepath.Join(dstDir, "dest.txt")

	// Test Copy
	err = CopyFile(srcFile.Name(), dstFile)
	assert.NoError(t, err)

	// Verify content
	readContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
}

func TestCopyDir(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "src-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(srcDir) }()

	// Create file in src
	_ = os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644)
	_ = os.Mkdir(filepath.Join(srcDir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "subdir", "subfile.txt"), []byte("subcontent"), 0644)

	dstDir, err := os.MkdirTemp("", "dst-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(dstDir) }()

	// CopyDir expects dest NOT to exist or be empty? Code says "destination directory must *not* exist" implied by MkdirAll but looping?
	// The implementation calls MkdirAll(destination). If it exists, it merges.
	// But let's test copying to a new subdir
	targetPath := filepath.Join(dstDir, "target")

	err = CopyDir(srcDir, targetPath)
	assert.NoError(t, err)

	// Verify
	assert.FileExists(t, filepath.Join(targetPath, "file.txt"))
	assert.FileExists(t, filepath.Join(targetPath, "subdir", "subfile.txt"))
}

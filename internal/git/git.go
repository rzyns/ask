package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clone clones a git repository to the specified destination
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InstallSubdir clones a repository to a temp directory and copies a specific subdirectory to valid destination
func InstallSubdir(repoURL, subDir, dest string) error {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "ask-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	fmt.Printf("Cloning %s to temp storage...\n", repoURL)
	// Clone to temp dir
	if err := Clone(repoURL, tempDir); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// Copy subdirectory
	srcPath := filepath.Join(tempDir, subDir)

	// Check if srcPath exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("subdirectory %s not found in repo", subDir)
	}

	fmt.Printf("Copying skill from %s...\n", subDir)
	return copyDir(srcPath, dest)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// GetLatestTag returns the latest tag for a repository in the given path
func GetLatestTag(repoPath string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Checkout checks out a specific tag or branch
func Checkout(repoPath, ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

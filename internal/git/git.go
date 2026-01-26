// Package git provides git operations helpers.
package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yeasy/ask/internal/ui"
)

// Clone clones a git repository to the specified destination
func Clone(url, dest string) error {
	bar := ui.NewSpinner(fmt.Sprintf("Cloning %s...", filepath.Base(url)))
	cmd := exec.Command("git", "clone", "--depth", "1", "--progress", url, dest)
	cmd.Stdout = bar
	cmd.Stderr = bar
	err := cmd.Run()
	_ = bar.Finish()
	return err
}

// SparseClone clones only a specific subdirectory using sparse checkout
// This is much faster than full clone for large repos when only a subdir is needed
func SparseClone(repoURL, branch, subDir, dest string) error {
	bar := ui.NewSpinner(fmt.Sprintf("Sparse cloning %s from %s...", subDir, filepath.Base(repoURL)))
	defer func() { _ = bar.Finish() }()

	// Step 1: Clone with filter and no checkout
	ui.UpdateDescription(bar, "Initializing sparse clone...")
	args := []string{"clone", "--filter=blob:none", "--no-checkout", "--depth", "1", "--progress"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, dest)

	cmd := exec.Command("git", args...)
	cmd.Stdout = bar
	cmd.Stderr = bar
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sparse clone init failed: %w", err)
	}

	// Step 2: Initialize sparse-checkout in cone mode
	ui.UpdateDescription(bar, "Configuring sparse checkout...")
	cmd = exec.Command("git", "sparse-checkout", "init", "--cone")
	cmd.Dir = dest
	cmd.Stdout = bar
	cmd.Stderr = bar
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sparse-checkout init failed: %w", err)
	}

	// Step 3: Set the subdirectory to checkout
	ui.UpdateDescription(bar, fmt.Sprintf("Setting checkout path to %s...", subDir))
	cmd = exec.Command("git", "sparse-checkout", "set", subDir)
	cmd.Dir = dest
	cmd.Stdout = bar
	cmd.Stderr = bar
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sparse-checkout set failed: %w", err)
	}

	// Step 4: Checkout
	ui.UpdateDescription(bar, "Checking out files...")
	cmd = exec.Command("git", "checkout")
	cmd.Dir = dest
	cmd.Stdout = bar
	cmd.Stderr = bar
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("checkout failed: %w", err)
	}

	return nil
}

// InstallSubdir installs a subdirectory from a repository
// Uses sparse checkout for efficiency, falls back to full clone if sparse fails
func InstallSubdir(repoURL, branch, subDir, dest string) error {
	// Create temp dir for sparse clone
	tempDir, err := os.MkdirTemp("", "ask-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }() // Clean up

	// Try sparse checkout first
	if err := SparseClone(repoURL, branch, subDir, tempDir); err != nil {
		// Fallback to full clone
		fmt.Printf("Sparse checkout failed, falling back to full clone...\n")
		_ = os.RemoveAll(tempDir) // Clean up failed sparse clone
		tempDir, err = os.MkdirTemp("", "ask-install-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}

		if err := Clone(repoURL, tempDir); err != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}

		// If branch is specified, checkout that branch in full clone fallback
		if branch != "" {
			if err := Checkout(tempDir, branch); err != nil {
				return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
			}
		}
	}

	// Copy subdirectory to destination
	srcPath := filepath.Join(tempDir, subDir)

	// Check if srcPath exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("subdirectory %s not found in repo", subDir)
	}

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
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

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

// GetCurrentCommit returns the current commit hash of the repository
func GetCurrentCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

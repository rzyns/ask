package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGetLatestTag tests the GetLatestTag function
// Note: This test requires an actual git repository with tags
func TestGetLatestTag(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	// Configure git user for commits
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	// Create a test file and commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create a tag
	tagCmd := exec.Command("git", "tag", "v1.0.0")
	tagCmd.Dir = tmpDir
	if err := tagCmd.Run(); err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Test GetLatestTag
	tag, err := GetLatestTag(tmpDir)
	if err != nil {
		t.Fatalf("GetLatestTag failed: %v", err)
	}

	if tag != "v1.0.0" {
		t.Errorf("Expected tag 'v1.0.0', got '%s'", tag)
	}
}

// TestGetCurrentCommit tests the GetCurrentCommit function
func TestGetCurrentCommit(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	// Configure git user
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Test commit")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test GetCurrentCommit
	commit, err := GetCurrentCommit(tmpDir)
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}

	// Verify it's a valid SHA (40 characters)
	if len(commit) != 40 {
		t.Errorf("Expected 40 character SHA, got %d characters: %s", len(commit), commit)
	}
}

// TestCheckout tests the Checkout function
func TestCheckout(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	// Configure git user
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	// Create initial commit on main
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = tmpDir
	_ = addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "Main commit")
	commitCmd.Dir = tmpDir
	_ = commitCmd.Run()

	// Create a branch
	branchCmd := exec.Command("git", "branch", "test-branch")
	branchCmd.Dir = tmpDir
	if err := branchCmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Test checkout
	err := Checkout(tmpDir, "test-branch")
	if err != nil {
		t.Errorf("Checkout failed: %v", err)
	}

	// Verify we're on the branch
	verifyCmd := exec.Command("git", "branch", "--show-current")
	verifyCmd.Dir = tmpDir
	output, err := verifyCmd.Output()
	if err != nil {
		t.Fatalf("Failed to verify branch: %v", err)
	}

	currentBranch := string(output)
	if currentBranch != "test-branch\n" {
		t.Errorf("Expected to be on 'test-branch', got '%s'", currentBranch)
	}
}

// TestCopyDir tests the copyDir function
func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory with files
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	// Create test files
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subDir := filepath.Join(srcDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Copy directory
	dstDir := filepath.Join(tmpDir, "dst")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify files were copied
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt was not copied")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "file2.txt")); os.IsNotExist(err) {
		t.Error("subdir/file2.txt was not copied")
	}

	// Verify content
	content, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(content) != "content1" {
		t.Errorf("Expected content 'content1', got '%s'", string(content))
	}
}

// TestCopyFile tests the copyFile function
func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	// Create source file
	testContent := "test file content"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify content
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}
}

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGit executes a git command in the specified directory with a clean environment
func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Filter out GIT_ vars from environment to prevent contamination from parent repo
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_") {
			env = append(env, e)
		}
	}
	cmd.Env = env

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
}

// setupTestRepo creates a temp dir, initializes git, and configures a user
func setupTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "commit.gpgsign", "false")
	return tmpDir
}

// TestGetLatestTag tests the GetLatestTag function
func TestGetLatestTag(t *testing.T) {
	tmpDir := setupTestRepo(t)

	// Create a test file and commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	// Create a tag
	runGit(t, tmpDir, "tag", "v1.0.0")

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
	tmpDir := setupTestRepo(t)

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "Test commit")

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
	tmpDir := setupTestRepo(t)

	// Create initial commit on main
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create initial commit
	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "Main commit")

	// Create a branch
	runGit(t, tmpDir, "branch", "test-branch")

	// Test checkout
	err := Checkout(tmpDir, "test-branch")
	if err != nil {
		t.Errorf("Checkout failed: %v", err)
	}

	// Verify we're on the branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = tmpDir
	// Sanitize env here too for verification
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_") {
			env = append(env, e)
		}
	}
	cmd.Env = env

	output, err := cmd.Output()
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

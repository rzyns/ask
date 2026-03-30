package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Error("NewApp() returned nil")
	}
}

func TestNewApp_ReturnsDistinctInstances(t *testing.T) {
	a1 := NewApp()
	a2 := NewApp()
	if a1 == a2 {
		t.Error("NewApp() should return distinct instances")
	}
}

func TestApp_Startup(t *testing.T) {
	app := NewApp()
	ctx := context.TODO()
	app.Startup(ctx)
	if app.ctx != ctx {
		t.Error("Startup did not save the context")
	}
}

func TestApp_Startup_WithLastProjectRoot(t *testing.T) {
	// Save original working directory to restore later
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Redirect HOME to a temp directory so we never touch the real ~/.ask/config.yaml
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Create a temporary directory to act as the last project root
	projectDir := t.TempDir()

	// Ensure the global config dir exists under the fake HOME
	_ = config.EnsureGlobalDirExists()

	// Write a config with LastProjectRoot set to projectDir
	cfg := config.DefaultConfig()
	cfg.LastProjectRoot = projectDir
	if err := cfg.SaveGlobal(); err != nil {
		t.Fatalf("failed to save global config: %v", err)
	}

	app := NewApp()
	app.Startup(context.Background())

	// Verify we changed to the last project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Resolve symlinks for comparison (e.g., /tmp vs /private/tmp on macOS)
	cwdReal, _ := filepath.EvalSymlinks(cwd)
	projectDirReal, _ := filepath.EvalSymlinks(projectDir)

	if cwdReal != projectDirReal {
		t.Errorf("expected cwd %s, got %s", projectDirReal, cwdReal)
	}
}

func TestApp_Startup_WithInvalidLastProjectRoot(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Redirect HOME to a temp directory so we never touch the real ~/.ask/config.yaml
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	_ = config.EnsureGlobalDirExists()

	// Set LastProjectRoot to a non-existent directory
	cfg := config.DefaultConfig()
	cfg.LastProjectRoot = "/nonexistent/path/that/does/not/exist"
	if err := cfg.SaveGlobal(); err != nil {
		t.Fatalf("failed to save global config: %v", err)
	}

	app := NewApp()
	// Should not panic; falls back to home
	app.Startup(context.Background())

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cwdReal, _ := filepath.EvalSymlinks(cwd)
	fakeHomeReal, _ := filepath.EvalSymlinks(fakeHome)

	if cwdReal != fakeHomeReal {
		t.Errorf("expected fallback to home %s, got %s", fakeHomeReal, cwdReal)
	}
}

func TestFallbackToHome(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Change to a temp dir first so we can verify fallbackToHome changes it
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	fallbackToHome()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cwdReal, _ := filepath.EvalSymlinks(cwd)
	homeReal, _ := filepath.EvalSymlinks(home)

	if cwdReal != homeReal {
		t.Errorf("fallbackToHome should chdir to %s, got %s", homeReal, cwdReal)
	}
}

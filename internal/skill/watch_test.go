package skill

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatchAndCheck_DetectsChanges(t *testing.T) {
	// Create a temporary skill directory
	dir := t.TempDir()

	// Create a minimal SKILL.md
	skillMD := `---
name: watch-test
version: 0.1.0
description: Test skill for watch mode
---
# watch-test
A test skill.
`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a clean prompt file
	promptDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptDir, "main.md"), []byte("Hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var events []string

	// Start watching in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchAndCheck(dir, func(event string, _ *CheckResult, _ error) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, event)
		})
	}()

	// Give the watcher time to initialize
	time.Sleep(500 * time.Millisecond)

	// Modify a file to trigger the watch
	if err := os.WriteFile(filepath.Join(promptDir, "main.md"), []byte("Modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for debounce + processing
	time.Sleep(1 * time.Second)

	mu.Lock()
	eventCount := len(events)
	mu.Unlock()

	if eventCount == 0 {
		t.Error("expected at least one callback event after file modification")
	}
}

func TestWatchAndCheck_IgnoresGitDir(t *testing.T) {
	dir := t.TempDir()

	skillMD := `---
name: watch-git-test
version: 0.1.0
description: Test skill
---
# watch-git-test
A test skill.
`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .git directory
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var events []string

	go func() {
		_ = WatchAndCheck(dir, func(event string, _ *CheckResult, _ error) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, event)
		})
	}()

	time.Sleep(500 * time.Millisecond)

	// Write to .git directory — should be ignored
	if err := os.WriteFile(filepath.Join(gitDir, "index"), []byte("git data"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(800 * time.Millisecond)

	mu.Lock()
	eventCount := len(events)
	mu.Unlock()

	if eventCount > 0 {
		t.Errorf("expected no events for .git changes, got %d", eventCount)
	}
}

func TestWatchAndCheck_DetectsNewFiles(t *testing.T) {
	dir := t.TempDir()

	skillMD := `---
name: watch-new-file-test
version: 0.1.0
description: Test skill
---
# watch-new-file-test
A test skill.
`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var events []string

	go func() {
		_ = WatchAndCheck(dir, func(event string, _ *CheckResult, _ error) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, event)
		})
	}()

	time.Sleep(500 * time.Millisecond)

	// Create a new file
	if err := os.WriteFile(filepath.Join(dir, "new-prompt.md"), []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	mu.Lock()
	eventCount := len(events)
	mu.Unlock()

	if eventCount == 0 {
		t.Error("expected at least one callback event for new file creation")
	}
}

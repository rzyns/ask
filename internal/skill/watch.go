package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchAndCheck watches a skill directory for file changes and re-runs security checks.
// It blocks until the context is canceled or an unrecoverable error occurs.
// The callback is invoked after each check with the result (nil result on error).
func WatchAndCheck(skillPath string, callback func(event string, result *CheckResult, err error)) error {
	absPath, err := filepath.Abs(skillPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Add the skill directory and all subdirectories
	if err := addDirRecursive(watcher, absPath); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Debounce: collect events within a short window before re-checking
	var debounceTimer *time.Timer
	debounceDelay := 300 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil // Watcher closed
			}

			// Ignore non-write events
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Remove) {
				continue
			}

			// Ignore .git directory changes
			rel, _ := filepath.Rel(absPath, event.Name)
			if strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) || rel == ".git" {
				continue
			}

			// Skip binary files
			ext := strings.ToLower(filepath.Ext(event.Name))
			if isBinaryExt(ext) {
				continue
			}

			// If a new directory was created, watch it too
			if event.Has(fsnotify.Create) {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addDirRecursive(watcher, event.Name)
				}
			}

			displayRel := rel
			if displayRel == "" {
				displayRel = filepath.Base(event.Name)
			}

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				result, checkErr := CheckSafety(absPath)
				callback(displayRel, result, checkErr)
			})

		case watchErr, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			callback("", nil, watchErr)
		}
	}
}

// addDirRecursive adds a directory and all its subdirectories to the watcher.
func addDirRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		// Skip symlinks to prevent following links outside intended directory
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
}

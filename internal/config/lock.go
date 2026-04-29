package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yeasy/ask/internal/filesystem"
	"gopkg.in/yaml.v3"
)

// LockFileName is the name of the lock file
const LockFileName = "ask.lock"

// LockEntry represents a locked skill version
type LockEntry struct {
	Name        string    `yaml:"name"`
	Source      string    `yaml:"source,omitempty"`
	URL         string    `yaml:"url,omitempty"`
	Commit      string    `yaml:"commit,omitempty"`
	Version     string    `yaml:"version,omitempty"`
	InstalledAt time.Time `yaml:"installed_at"`

	// Optional provenance metadata for adopted in-place agent skills.
	Agent          string `yaml:"agent,omitempty"`
	Ownership      string `yaml:"ownership,omitempty"`
	InstallMode    string `yaml:"install_mode,omitempty"`
	UpdateStrategy string `yaml:"update_strategy,omitempty"`
	TargetPath     string `yaml:"target_path,omitempty"`
	Checksum       string `yaml:"checksum,omitempty"`
}

// LockFile represents the ask.lock file structure
type LockFile struct {
	Version int         `yaml:"version"`
	Skills  []LockEntry `yaml:"skills"`
}

// maxLockFileSize limits the lock file size to prevent OOM on malformed files
const maxLockFileSize = 1024 * 1024 // 1MB

// loadLockFromPath loads a lock file from the given path with size validation.
// Uses Lstat pre-check for symlinks, then open-then-fstat for size validation.
func loadLockFromPath(path string) (*LockFile, error) {
	// Pre-check for symlinks (Lstat does not follow symlinks)
	linfo, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{Version: 1, Skills: []LockEntry{}}, nil
		}
		return nil, fmt.Errorf("read lock file: %w", err)
	}
	if linfo.Mode()&os.ModeSymlink != 0 || !linfo.Mode().IsRegular() {
		return nil, fmt.Errorf("refusing to read non-regular file: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read lock file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat lock file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("refusing to read non-regular file: %s", path)
	}
	if info.Size() > maxLockFileSize {
		return nil, fmt.Errorf("lock file too large: %d bytes (max %d)", info.Size(), maxLockFileSize)
	}
	data, err := io.ReadAll(io.LimitReader(f, maxLockFileSize))
	if err != nil {
		return nil, fmt.Errorf("read lock file: %w", err)
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse lock file: %w", err)
	}

	return &lock, nil
}

// LoadLockFile loads the ask.lock file
func LoadLockFile() (*LockFile, error) {
	return loadLockFromPath(LockFileName)
}

// Save saves the lock file to disk atomically
func (l *LockFile) Save() error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshal lock file: %w", err)
	}
	if err := filesystem.AtomicWriteFile(LockFileName, data, 0600); err != nil {
		return fmt.Errorf("write lock file: %w", err)
	}
	return nil
}

// AddEntry adds or updates a lock entry.
//
// Entries with agent metadata are keyed by (name, agent) so different agents can
// safely lock skills with the same name. Legacy entries without agent metadata
// retain the original name-only identity.
func (l *LockFile) AddEntry(entry LockEntry) {
	for i, e := range l.Skills {
		if lockEntrySameIdentity(e, entry) {
			l.Skills[i] = entry
			return
		}
	}
	l.Skills = append(l.Skills, entry)
}

func lockEntrySameIdentity(existing, incoming LockEntry) bool {
	if existing.Name != incoming.Name {
		return false
	}
	existingAgent := strings.TrimSpace(existing.Agent)
	incomingAgent := strings.TrimSpace(incoming.Agent)
	if existingAgent != "" || incomingAgent != "" {
		if !strings.EqualFold(existingAgent, incomingAgent) {
			return false
		}
		existingPath := normalizeLockTargetPath(existing.TargetPath)
		incomingPath := normalizeLockTargetPath(incoming.TargetPath)
		if existingPath != "" || incomingPath != "" {
			return existingPath != "" && existingPath == incomingPath
		}
		return true
	}
	return true
}

// RemoveEntry removes a lock entry by name
func (l *LockFile) RemoveEntry(name string) {
	for i, e := range l.Skills {
		if e.Name == name {
			l.Skills = append(l.Skills[:i], l.Skills[i+1:]...)
			return
		}
	}
}

// GetEntry gets a lock entry by name.
func (l *LockFile) GetEntry(name string) *LockEntry {
	for i := range l.Skills {
		if l.Skills[i].Name == name {
			return &l.Skills[i]
		}
	}
	return nil
}

// GetEntryForAgent gets a lock entry by name and agent. Legacy lock entries
// without agent metadata are treated as matching a requested agent for backward
// compatibility with lockfiles created before agent-scoped provenance existed.
func (l *LockFile) GetEntryForAgent(name, agent string) *LockEntry {
	requestedAgent := strings.TrimSpace(agent)
	for i := range l.Skills {
		if l.Skills[i].Name != name {
			continue
		}
		entryAgent := strings.TrimSpace(l.Skills[i].Agent)
		if strings.EqualFold(entryAgent, requestedAgent) || (requestedAgent != "" && entryAgent == "") {
			return &l.Skills[i]
		}
	}
	return nil
}

// GetEntryForAgentTargetPath gets a lock entry by name, agent, and target path.
// If a matching legacy entry has no target path, it is treated as a match for
// compatibility with pre-provenance lockfiles.
func (l *LockFile) GetEntryForAgentTargetPath(name, agent, targetPath string) *LockEntry {
	requestedAgent := strings.TrimSpace(agent)
	requestedPath := normalizeLockTargetPath(targetPath)
	for i := range l.Skills {
		if l.Skills[i].Name != name {
			continue
		}
		entryAgent := strings.TrimSpace(l.Skills[i].Agent)
		if !(strings.EqualFold(entryAgent, requestedAgent) || (requestedAgent != "" && entryAgent == "")) {
			continue
		}
		entryPath := normalizeLockTargetPath(l.Skills[i].TargetPath)
		if entryPath == "" || entryPath == requestedPath {
			return &l.Skills[i]
		}
	}
	return nil
}

func normalizeLockTargetPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

// LoadGlobalLockFile loads the global lock file (~/.ask/ask.lock)
func LoadGlobalLockFile() (*LockFile, error) {
	path, err := GetGlobalLockPath()
	if err != nil {
		return nil, err
	}
	return loadLockFromPath(path)
}

// SaveGlobal saves the lock file to the global location (~/.ask/ask.lock) atomically
func (l *LockFile) SaveGlobal() error {
	if err := EnsureGlobalDirExists(); err != nil {
		return fmt.Errorf("ensure global dir: %w", err)
	}

	path, err := GetGlobalLockPath()
	if err != nil {
		return fmt.Errorf("resolve global lock path: %w", err)
	}

	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshal global lock file: %w", err)
	}
	if err := filesystem.AtomicWriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write global lock file: %w", err)
	}
	return nil
}

// LoadLockFileByScope loads lock file based on global flag
func LoadLockFileByScope(global bool) (*LockFile, error) {
	if global {
		return LoadGlobalLockFile()
	}
	return LoadLockFile()
}

// SaveByScope saves lock file based on global flag
func (l *LockFile) SaveByScope(global bool) error {
	if global {
		return l.SaveGlobal()
	}
	return l.Save()
}

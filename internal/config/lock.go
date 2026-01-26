package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LockFileName is the name of the lock file
const LockFileName = "ask.lock"

// LockEntry represents a locked skill version
type LockEntry struct {
	Name        string    `yaml:"name"`
	Source      string    `yaml:"source,omitempty"`
	URL         string    `yaml:"url"`
	Commit      string    `yaml:"commit,omitempty"`
	Version     string    `yaml:"version,omitempty"`
	InstalledAt time.Time `yaml:"installed_at"`
}

// LockFile represents the ask.lock file structure
type LockFile struct {
	Version int         `yaml:"version"`
	Skills  []LockEntry `yaml:"skills"`
}

// LoadLockFile loads the ask.lock file
func LoadLockFile() (*LockFile, error) {
	data, err := os.ReadFile(LockFileName)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty lock file if doesn't exist
			return &LockFile{Version: 1, Skills: []LockEntry{}}, nil
		}
		return nil, err
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// Save saves the lock file to disk
func (l *LockFile) Save() error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(LockFileName, data, 0644)
}

// AddEntry adds or updates a lock entry
func (l *LockFile) AddEntry(entry LockEntry) {
	// Update if exists
	for i, e := range l.Skills {
		if e.Name == entry.Name {
			l.Skills[i] = entry
			return
		}
	}
	// Add new
	l.Skills = append(l.Skills, entry)
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

// GetEntry gets a lock entry by name
func (l *LockFile) GetEntry(name string) *LockEntry {
	for _, e := range l.Skills {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

// LoadGlobalLockFile loads the global lock file (~/.ask/ask.lock)
func LoadGlobalLockFile() (*LockFile, error) {
	lockPath := GetGlobalLockPath()
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty lock file if doesn't exist
			return &LockFile{Version: 1, Skills: []LockEntry{}}, nil
		}
		return nil, err
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// SaveGlobal saves the lock file to the global location (~/.ask/ask.lock)
func (l *LockFile) SaveGlobal() error {
	if err := EnsureGlobalDirExists(); err != nil {
		return err
	}

	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(GetGlobalLockPath(), data, 0644)
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

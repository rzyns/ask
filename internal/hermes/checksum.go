package hermes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ChecksumSkillDir computes a deterministic checksum for regular files under a
// skill directory. It skips hidden VCS/cache directories and never follows
// symlinks.
func ChecksumSkillDir(dir string) (string, error) {
	var files []string
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == ".hg" || name == ".svn" || name == "node_modules" || name == "__pycache__" || name == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(files)

	h := sha256.New()
	for _, path := range files {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return "", err
		}
		rel = filepath.ToSlash(rel)
		fmt.Fprintf(h, "path:%s\n", rel)
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(h, file)
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		_, _ = h.Write([]byte("\n"))
	}
	return "sha256:" + strings.ToLower(hex.EncodeToString(h.Sum(nil))), nil
}

package skill

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckConfig represents a .askcheck.yaml configuration file
type CheckConfig struct {
	// Ignore lists rule IDs to suppress (e.g., ["SECRET-GENERIC-TOKEN", "CMD-SUDO"])
	Ignore []string `yaml:"ignore"`
	// IgnorePaths lists file/directory glob patterns to skip (e.g., ["vendor/**", "*.test.js"])
	IgnorePaths []string `yaml:"ignore_paths"`
	// Rules defines additional custom rules
	Rules []CustomRuleDef `yaml:"rules"`
}

// CustomRuleDef represents a user-defined rule in .askcheck.yaml
type CustomRuleDef struct {
	ID          string `yaml:"id"`
	Pattern     string `yaml:"pattern"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
}

// LoadCheckConfig loads .askcheck.yaml from the given directory.
// Returns nil (no error) if the file does not exist.
func LoadCheckConfig(dir string) (*CheckConfig, error) {
	path := filepath.Join(dir, ".askcheck.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Also try .askcheck.yml
			path = filepath.Join(dir, ".askcheck.yml")
			data, err = os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, nil
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var cfg CheckConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BuildRules returns the effective rule set: default rules (minus ignored) plus custom rules.
func (cc *CheckConfig) BuildRules() []Rule {
	ignoreSet := make(map[string]bool)
	if cc != nil {
		for _, id := range cc.Ignore {
			ignoreSet[strings.ToUpper(id)] = true
		}
	}

	// Start with default rules, filtering out ignored
	var rules []Rule
	for _, r := range defaultRules {
		if !ignoreSet[strings.ToUpper(r.ID)] {
			rules = append(rules, r)
		}
	}

	// Append custom rules
	if cc != nil {
		for _, cr := range cc.Rules {
			compiled, err := regexp.Compile(cr.Pattern)
			if err != nil {
				continue // Skip invalid patterns
			}
			sev := SeverityWarning
			switch strings.ToLower(cr.Severity) {
			case "critical":
				sev = SeverityCritical
			case "info":
				sev = SeverityInfo
			}
			rules = append(rules, Rule{
				ID:          cr.ID,
				Description: cr.Description,
				Severity:    sev,
				Regex:       compiled,
			})
		}
	}

	return rules
}

// IsPathIgnored returns true if the relative path matches any ignore_paths pattern.
func (cc *CheckConfig) IsPathIgnored(relPath string) bool {
	if cc == nil {
		return false
	}
	for _, pattern := range cc.IgnorePaths {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		// Also check against just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
		// Support ** prefix by checking suffix
		if strings.HasPrefix(pattern, "**") {
			suffix := strings.TrimPrefix(pattern, "**")
			suffix = strings.TrimPrefix(suffix, "/")
			suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
			if strings.HasSuffix(relPath, suffix) || strings.Contains(relPath, suffix) {
				return true
			}
		}
		// Support directory prefix patterns like "vendor/**"
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(relPath, prefix+"/") || strings.HasPrefix(relPath, prefix+string(filepath.Separator)) || relPath == prefix {
				return true
			}
		}
	}
	return false
}

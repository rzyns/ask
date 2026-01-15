package deps

import (
	"fmt"
	"path/filepath"

	"github.com/yeasy/ask/internal/skill"
)

// Resolver handles dependency resolution for skills
type Resolver struct {
	visited  map[string]bool
	resolved map[string]bool
	order    []string
}

// NewResolver creates a new dependency resolver
func NewResolver() *Resolver {
	return &Resolver{
		visited:  make(map[string]bool),
		resolved: make(map[string]bool),
		order:    []string{},
	}
}

// Resolve returns the ordered list of dependencies to install
// Returns dependencies in topological order (deps first, then skill)
func (r *Resolver) Resolve(skillPath string) ([]string, error) {
	skillName := filepath.Base(skillPath)
	return r.resolve(skillName, skillPath, []string{})
}

func (r *Resolver) resolve(name, path string, chain []string) ([]string, error) {
	// Check for circular dependency
	for _, c := range chain {
		if c == name {
			return nil, fmt.Errorf("circular dependency detected: %v -> %s", chain, name)
		}
	}

	// Already resolved
	if r.resolved[name] {
		return nil, nil
	}

	// Mark as visited
	r.visited[name] = true
	chain = append(chain, name)

	// Parse SKILL.md for dependencies
	if skill.FindSkillMD(path) {
		meta, err := skill.ParseSkillMD(path)
		if err == nil && meta != nil && len(meta.Dependencies) > 0 {
			for _, dep := range meta.Dependencies {
				depPath := filepath.Join(filepath.Dir(path), dep)
				_, err := r.resolve(dep, depPath, chain)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Mark as resolved and add to order
	r.resolved[name] = true
	r.order = append(r.order, name)

	return r.order, nil
}

// GetOrder returns the resolved installation order
func (r *Resolver) GetOrder() []string {
	return r.order
}

// GetDependencies extracts dependencies from a skill's SKILL.md
func GetDependencies(skillPath string) ([]string, error) {
	if !skill.FindSkillMD(skillPath) {
		return nil, nil
	}

	meta, err := skill.ParseSkillMD(skillPath)
	if err != nil {
		return nil, err
	}

	return meta.Dependencies, nil
}

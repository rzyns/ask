// Package server provides an embedded HTTP server for the ask web UI.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/yeasy/ask/internal/cache"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/skill"
	"github.com/yeasy/ask/internal/ui"
)

//go:embed web/*
var webFS embed.FS

// Server represents the HTTP server
// Server represents the HTTP server
type Server struct {
	port    int
	server  *http.Server
	mu      sync.Mutex
	version string
}

// New creates a new Server instance
func New(port int, version string) *Server {
	return &Server{
		port:    port,
		version: version,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := s.setupRoutes()

	// Static file serving
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler:           corsMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.mu.Lock()
	s.server = server
	s.mu.Unlock()

	ui.Info(fmt.Sprintf("Starting server on http://127.0.0.1:%d", s.port))
	return server.ListenAndServe()
}

// setupRoutes returns the API mux
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/skills", s.handleSkills)
	mux.HandleFunc("/api/skills/search", s.handleSkillSearch)
	mux.HandleFunc("/api/skills/install", s.handleSkillInstall)
	mux.HandleFunc("/api/skills/uninstall", s.handleSkillUninstall)
	mux.HandleFunc("/api/repos", s.handleRepos)
	mux.HandleFunc("/api/repos/add", s.handleRepoAdd)
	mux.HandleFunc("/api/repos/remove", s.handleRepoRemove)
	mux.HandleFunc("/api/repos/sync", s.handleRepoSync)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/config/update", s.handleConfigUpdate)
	mux.HandleFunc("/api/cache/clear", s.handleCacheClear)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/skills/readme", s.handleSkillReadme)

	return mux
}

// Handler returns the HTTP handler for the server (exported for Wails integration)
func (s *Server) Handler() http.Handler {
	return s.setupRoutes()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	server := s.server
	s.mu.Unlock()

	if server != nil {
		return server.Shutdown(ctx)
	}
	return nil
}

// OpenBrowser opens the default browser to the server URL
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JSON response helpers
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// API Handlers

// SkillInfo represents a skill for API responses
type SkillInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url,omitempty"`
	Version     string   `json:"version,omitempty"`
	Path        string   `json:"path,omitempty"`
	Agents      []string `json:"agents,omitempty"`
	Repo        string   `json:"repo,omitempty"`
	IconURL     string   `json:"icon_url,omitempty"`
}

func getRepoFromGitConfig(path string) string {
	// Only check the skill's own .git directory
	// Skills are typically cloned directly, so .git should be at the skill path
	gitConfigPath := filepath.Join(path, ".git", "config")
	return parseGitConfigForRepo(gitConfigPath)
}

func parseGitConfigForRepo(gitConfigPath string) string {
	data, err := os.ReadFile(gitConfigPath)
	if err != nil {
		return ""
	}
	content := string(data)
	// Simple parsing for remote "origin" url
	// [remote "origin"]
	// 	url = https://github.com/owner/repo.git or git@github.com:owner/repo.git
	if idx := strings.Index(content, "[remote \"origin\"]"); idx != -1 {
		rest := content[idx:]
		if urlIdx := strings.Index(rest, "url = "); urlIdx != -1 {
			start := urlIdx + 6
			end := strings.Index(rest[start:], "\n")
			if end != -1 {
				url := strings.TrimSpace(rest[start : start+end])
				// Clean up URL to get owner/repo
				url = strings.TrimSuffix(url, ".git")
				if strings.Contains(url, "github.com") {
					parts := strings.Split(url, "github.com")
					if len(parts) > 1 {
						return strings.Trim(parts[1], "/:")
					}
				}
			}
		}
	}
	return ""
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			def := config.DefaultConfig()
			cfg = &def
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Map to aggregate skill info by name
	skillMap := make(map[string]*SkillInfo)

	// Helper to get or create skill info
	getOrCreate := func(name string) *SkillInfo {
		if _, exists := skillMap[name]; !exists {
			skillMap[name] = &SkillInfo{Name: name, Agents: []string{}}
		}
		return skillMap[name]
	}

	// Load lockfile for additional metadata
	lockFile, _ := config.LoadLockFile() // Ignore error, file might not exist
	lockMap := make(map[string]string)   // map name -> url
	if lockFile != nil {
		for _, entry := range lockFile.Skills {
			lockMap[entry.Name] = entry.URL
		}
	}

	// 1. Scan Configured/Legacy Skills first (base metadata)
	for _, si := range cfg.SkillsInfo {
		info := getOrCreate(si.Name)
		info.Description = si.Description
		info.URL = si.URL
		// Try to deduce repo from URL if present
		if si.URL != "" && strings.Contains(si.URL, "github.com") {
			parts := strings.Split(si.URL, "github.com/")
			if len(parts) > 1 {
				repoName := strings.TrimSuffix(parts[1], ".git")
				info.Repo = repoName
			}
		}
	}
	for _, name := range cfg.Skills {
		getOrCreate(name)
	}

	// 2. Scan each Agent directory for installed skills
	// We want to verify which agents actually have the skill installed
	toolTargets := cfg.GetEnabledToolTargets()

	for _, target := range toolTargets {
		// List subdirectories in the agent's skills dir
		entries, err := os.ReadDir(target.SkillsDir)
		if err != nil {
			continue // Skip if dir doesn't exist or is unreadable
		}

		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				name := entry.Name()
				skillPath := filepath.Join(target.SkillsDir, name)

				// Only consider it a skill if it contains SKILL.md
				// This prevents repo root directories from being mistakenly treated as skills
				if !skill.FindSkillMD(skillPath) {
					continue
				}

				info := getOrCreate(name)

				// Add this agent to the list (deduplicate just in case)
				found := false
				for _, a := range info.Agents {
					if a == target.Name {
						found = true
						break
					}
				}
				if !found {
					info.Agents = append(info.Agents, target.Name)
				}

				// If we don't have a path yet (or if this is the first one found), set it
				// This might be ambiguous if installed in multiple locations, but just picking one for "Path" is fine for basic file ops
				if info.Path == "" {
					info.Path = skillPath
				}

				// Try to deduce repo from .git/config if not present
				if info.Repo == "" {
					info.Repo = getRepoFromGitConfig(skillPath)
				}

				// Try to deduce repo from lockfile if still missing
				if info.Repo == "" && info.URL == "" {
					if url, ok := lockMap[name]; ok {
						info.URL = url
					}
				}

				// Ensure info.Repo is populated if URL is present (from lockfile or other source)
				if info.Repo == "" && info.URL != "" && strings.Contains(info.URL, "github.com") {
					// Cleaner logic:
					// If contains "github.com/", get everything after the LAST occurrence.
					lastIdx := strings.LastIndex(info.URL, "github.com/")
					if lastIdx != -1 {
						repoName := info.URL[lastIdx+11:] // len("github.com/") = 11
						repoName = strings.TrimSuffix(repoName, ".git")
						// Clean up optional /tree/... path if present (deep link)
						if idx := strings.Index(repoName, "/tree/"); idx != -1 {
							repoName = repoName[:idx]
						}
						info.Repo = repoName
					}
				}

				// Try to read SKILL.md for metadata if missing
				if info.Version == "" || info.Description == "" || info.Repo == "" {
					if meta, err := skill.ParseSkillMD(skillPath); err == nil && meta != nil {
						if meta.Version != "" {
							info.Version = meta.Version
						}
						if info.Description == "" && meta.Description != "" {
							info.Description = meta.Description
						}
						// If SKILL.md has repository info (hypothetically), we could use it.
						// Currently skill.Metadata might not have it, but we can assume descriptions are meaningful.
					}
				}
			}
		}
	}

	// Convert map to slice
	skills := make([]SkillInfo, 0, len(skillMap))
	for _, info := range skillMap {
		// Normalize Repo Name to match configured aliases
		// If info.Repo matches a configured repo URL or Name, use the Name.
		if info.Repo != "" {
			for _, repo := range cfg.Repos {
				// Check specific name match
				if strings.EqualFold(info.Repo, repo.Name) {
					info.Repo = repo.Name
					break
				}
				// Check URL match
				// Repo URL might be "https://github.com/owner/repo" or "owner/repo"
				// info.Repo (deduced) is usually "owner/repo"

				// Normalize repo.URL to owner/repo for comparison if possible
				repoURL := repo.URL
				if strings.HasPrefix(repoURL, "https://github.com/") {
					repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
					repoURL = strings.TrimSuffix(repoURL, ".git")
				}
				// Also strip /tree/... if present
				if idx := strings.Index(repoURL, "/tree/"); idx != -1 {
					repoURL = repoURL[:idx]
				}

				// Check matches
				if strings.EqualFold(info.Repo, repoURL) {
					info.Repo = repo.Name
					break
				}
			}
		}

		// Only include skills that are either in config OR installed in at least one agent
		// Actually, if it's in config but directory not found, it might be "broken" or uninstalled properly
		// But let's show everything we know about.
		skills = append(skills, *info)
	}

	jsonResponse(w, skills)
}

// SearchResult represents a search result
type SearchResult struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Stars       int      `json:"stars"`
	URL         string   `json:"url"`
	Source      string   `json:"source"`
	RepoName    string   `json:"repo,omitempty"`
	Agents      []string `json:"agents,omitempty"`
}

func (s *Server) handleSkillSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	repoFilter := r.URL.Query().Get("repo")
	topic := "agent-skill"

	results := make([]SearchResult, 0)

	// 1. Search Local/Configured Repos via Cache
	reposCache, err := cache.NewReposCache()
	if err == nil {
		repoInfos, _ := reposCache.LoadIndex()
		repoMap := make(map[string]cache.RepoInfo)
		for _, info := range repoInfos {
			repoMap[info.Name] = info
		}

		skills, err := reposCache.SearchSkills(query)
		if err == nil {
			for _, skill := range skills {
				repo := repoMap[skill.RepoName]
				skillURL := repo.URL
				if skillURL != "" && !strings.HasSuffix(skillURL, ".git") {
					skillURL = fmt.Sprintf("%s/tree/HEAD/%s", strings.TrimSuffix(skillURL, "/"), skill.Path)
				}

				if repoFilter != "" && !strings.EqualFold(skill.RepoName, repoFilter) {
					continue
				}

				results = append(results, SearchResult{
					Name:        skill.Name,
					FullName:    skill.Name,
					Description: skill.Description,
					Stars:       repo.Stars,
					URL:         skillURL,
					Source:      "repo",
					RepoName:    skill.RepoName,
				})
			}
		} else {
			fmt.Printf("Error searching local cache: %v\n", err)
		}
	}

	// 2. Search GitHub
	if query == "" {
		query = ""
	}

	ghRepos, err := github.SearchTopic(topic, query)
	if err != nil {
		if len(results) > 0 {
			jsonResponse(w, results)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, repo := range ghRepos {
		results = append(results, SearchResult{
			Name:        repo.Name,
			FullName:    repo.FullName, // e.g. owner/repo
			Description: repo.Description,
			Stars:       repo.StargazersCount,
			URL:         repo.HTMLURL,
			Source:      "github",
			// For GitHub results, we don't have a local "RepoName" alias.
			// We could try to match against config, but for discovery it's fine to leave empty or use FullName?
			// app.js filters use state.repos (configured). If this result is NOT configured, it won't appear in filter dropdown anyway.
			// But if we want it to show up as "anthropics" if configured:
			// Let's leave it empty for now or use FullName as fallback if we want to filter search results by unconfigured repos?
			// The user issue is about dropdown mismatch.
			// If we leave it empty, it won't pollute the dropdown.
			RepoName: repo.FullName,
		})
	}

	jsonResponse(w, results)
}

// InstallRequest represents an install request
type InstallRequest struct {
	Name  string `json:"name"`
	Agent string `json:"agent,omitempty"`
}

func (s *Server) handleSkillInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "Skill name is required", http.StatusBadRequest)
		return
	}

	// Execute install command
	exe, err := os.Executable()
	if err != nil {
		jsonError(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}
	args := []string{"skill", "install", req.Name}
	if req.Agent != "" {
		args = append(args, "--agent", req.Agent)
	}

	cmd := exec.Command(exe, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		jsonError(w, fmt.Sprintf("Install failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Installed %s", req.Name),
		"output":  string(output),
	})
}

func (s *Server) handleSkillUninstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "Skill name is required", http.StatusBadRequest)
		return
	}

	// Execute uninstall command with --all to fully remove
	exe, err := os.Executable()
	if err != nil {
		jsonError(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}
	cmd := exec.Command(exe, "skill", "uninstall", "--all", req.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		jsonError(w, fmt.Sprintf("Uninstall failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Uninstalled %s", req.Name),
		"output":  string(output),
	})
}

// RepoInfo represents a repository for API responses
type RepoInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Stars int    `json:"stars"`
}

func (s *Server) handleRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			def := config.DefaultConfig()
			cfg = &def
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Load cached star counts
	starCache := make(map[string]int)
	reposCache, err := cache.NewReposCache()
	if err == nil {
		// Ignore error if index doesn't exist yet
		repos, _ := reposCache.LoadIndex()
		for _, repo := range repos {
			starCache[repo.Name] = repo.Stars
		}
	}

	repos := make([]RepoInfo, 0, len(cfg.Repos))
	for _, repo := range cfg.Repos {
		// Convert owner/repo or owner/repo/path format to full GitHub URL
		displayURL := repo.URL
		if !strings.HasPrefix(repo.URL, "http") && strings.Contains(repo.URL, "/") {
			parts := strings.SplitN(repo.URL, "/", 3)
			if len(parts) >= 2 {
				// Use just owner/repo for the display URL
				displayURL = fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1])
			}
		}
		info := RepoInfo{
			Name: repo.Name,
			Type: repo.Type,
			URL:  displayURL,
		}
		if stars, ok := starCache[repo.Name]; ok {
			info.Stars = stars
		}
		repos = append(repos, info)
	}

	jsonResponse(w, repos)
}

func (s *Server) handleRepoAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL  string `json:"url"`
		Sync bool   `json:"sync"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		jsonError(w, "Repository URL is required", http.StatusBadRequest)
		return
	}

	// Execute repo add command
	exe, err := os.Executable()
	if err != nil {
		jsonError(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}
	args := []string{"repo", "add", req.URL}
	if req.Sync {
		args = append(args, "--sync")
	}

	cmd := exec.Command(exe, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		jsonError(w, fmt.Sprintf("Add repo failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Added repository %s", req.URL),
		"output":  string(output),
	})
}

func (s *Server) handleRepoRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "Repository name is required", http.StatusBadRequest)
		return
	}

	// Execute repo remove command
	exe, err := os.Executable()
	if err != nil {
		jsonError(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}
	cmd := exec.Command(exe, "repo", "remove", req.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		jsonError(w, fmt.Sprintf("Remove repo failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Removed repository %s", req.Name),
		"output":  string(output),
	})
}

func (s *Server) handleRepoSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse optional body for specific repo name
	var req struct {
		Name string `json:"name"`
	}
	// Ignore decode error as body might be empty (sync all)
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Execute repo sync command
	exe, err := os.Executable()
	if err != nil {
		jsonError(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}

	args := []string{"repo", "sync"}
	if req.Name != "" {
		args = append(args, req.Name)
	}

	cmd := exec.Command(exe, args...)
	output, _ := cmd.CombinedOutput() // Ignore error for output return

	msg := "Repositories synced"
	if req.Name != "" {
		msg = fmt.Sprintf("Repository '%s' synced", req.Name)
	}

	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": msg,
		"output":  string(output),
	})
}

// ConfigInfo represents configuration for API responses
type ConfigInfo struct {
	Version     string              `json:"version"`
	SkillsDir   string              `json:"skills_dir"`
	Agents      []string            `json:"agents"`
	ToolTargets []config.ToolTarget `json:"tool_targets"`
	GlobalDir   string              `json:"global_dir"`
	ProjectRoot string              `json:"project_root"`
	Initialized bool                `json:"initialized"`
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadConfig()
	initialized := true
	if err != nil {
		if os.IsNotExist(err) {
			initialized = false
			def := config.DefaultConfig()
			cfg = &def
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	info := ConfigInfo{
		Version:     s.version,
		SkillsDir:   cfg.GetSkillsDir(),
		Agents:      config.GetSupportedAgentNames(),
		ToolTargets: cfg.GetToolTargets(),

		GlobalDir:   config.GetGlobalSkillsDir(),
		Initialized: initialized,
	}

	// Get current working directory for ProjectRoot
	if cwd, err := os.Getwd(); err == nil {
		info.ProjectRoot = cwd
	}

	jsonResponse(w, info)
}

func (s *Server) handleCacheClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear Repos Cache
	reposCache, err := cache.New(cache.GetReposCacheDir(), 0)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to access repos cache: %v", err), http.StatusInternalServerError)
		return
	}
	if err := reposCache.Clear(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to clear repos cache: %v", err), http.StatusInternalServerError)
		return
	}

	// Assume skills cache might be similar if needed, but for now repos cache is the main one
	// or we can clear the whole cache directory if preferred.
	// The cache package allows creating a cache instance on a dir.
	// Let's assume verifying cache clearing via repos cache is sufficient or we create a general one.

	jsonResponse(w, map[string]string{"status": "success", "message": "Cache cleared"})
}

func (s *Server) handleConfigUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Agent       string `json:"agent"`
		Enabled     bool   `json:"enabled"`
		SkillsDir   string `json:"skills_dir"`
		ProjectRoot string `json:"project_root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize default if missing
			def := config.DefaultConfig()
			cfg = &def
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	updated := false

	// Handle project_root update first, as it changes the context for everything else
	if req.ProjectRoot != "" {
		if err := os.Chdir(req.ProjectRoot); err != nil {
			jsonError(w, "Failed to change directory: "+err.Error(), http.StatusBadRequest)
			return
		}
		// RELOAD config from the new directory to ensure we are editing the correct file
		// and not overwriting the new directory's config with the old one.
		newCfg, err := config.LoadConfig()
		if err != nil {
			if os.IsNotExist(err) {
				// Initialize default if missing in new location
				def := config.DefaultConfig()
				cfg = &def
			} else {
				jsonError(w, "Failed to load config in new root: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			cfg = newCfg
		}
		// We successfully switched context. Verification of initialized state will happen on next config fetch.
		updated = true

		// Update persistent global config
		if globalCfg, err := config.LoadGlobalConfig(); err == nil {
			globalCfg.LastProjectRoot = req.ProjectRoot
			if err := globalCfg.SaveGlobal(); err != nil {
				fmt.Printf("Failed to save global config: %v\n", err)
			}
		}
	}

	// Handle skills_dir update
	if req.SkillsDir != "" {
		cfg.SkillsDir = req.SkillsDir
		updated = true
	}

	// Handle agent toggle
	if req.Agent != "" {
		// Normalize agent name
		req.Agent = strings.ToLower(req.Agent)

		// Update or add tool target
		found := false
		for i, t := range cfg.ToolTargets {
			if t.Name == req.Agent {
				cfg.ToolTargets[i].Enabled = req.Enabled
				found = true
				break
			}
		}
		if !found {
			// If not in explicit config, check defaults
			defaultTargets := config.DefaultToolTargets()
			for _, t := range defaultTargets {
				if t.Name == req.Agent {
					t.Enabled = req.Enabled
					cfg.ToolTargets = append(cfg.ToolTargets, t)
					found = true
					break
				}
			}
		}

		// If still not found (shouldn't happen for supported agents), create entry
		if !found {
			// Just try to find default settings for this agent type
			agentType, ok := config.ResolveAgentType(req.Agent)
			if ok {
				ac, ok := config.SupportedAgents[agentType]
				if ok {
					cfg.ToolTargets = append(cfg.ToolTargets, config.ToolTarget{
						Name:      req.Agent,
						SkillsDir: ac.ProjectDir,
						Enabled:   req.Enabled,
					})
				}
			}
		}
		updated = true
	}

	if updated {
		if err := cfg.Save(); err != nil {
			jsonError(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	jsonResponse(w, map[string]string{"status": "success", "message": "Configuration updated"})
}

// StatsInfo represents dashboard statistics
type StatsInfo struct {
	InstalledSkills int `json:"installed_skills"`
	ConfiguredRepos int `json:"configured_repos"`
	SyncedRepos     int `json:"synced_repos"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			def := config.DefaultConfig()
			cfg = &def
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Count installed skills
	skillCount := len(cfg.SkillsInfo)
	shown := make(map[string]bool)
	for _, si := range cfg.SkillsInfo {
		shown[si.Name] = true
	}
	for _, name := range cfg.Skills {
		if !shown[name] {
			skillCount++
		}
	}

	// Count synced repos
	syncedCount := 0
	reposDir := cache.GetReposCacheDir()
	if entries, err := os.ReadDir(reposDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				syncedCount++
			}
		}
	}

	stats := StatsInfo{
		InstalledSkills: skillCount,
		ConfiguredRepos: len(cfg.Repos),
		SyncedRepos:     syncedCount,
	}

	jsonResponse(w, stats)
}

func (s *Server) handleSkillReadme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		jsonError(w, "Skill name is required", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		jsonError(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	skillsDir := cfg.GetSkillsDir()
	skillPath := filepath.Join(skillsDir, name)

	// Check if skill exists
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		jsonError(w, "Skill not found", http.StatusNotFound)
		return
	}

	// Try to find SKILL.md (case insensitive)
	readmePath := ""
	entries, err := os.ReadDir(skillPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(entry.Name(), "SKILL.md") {
				readmePath = filepath.Join(skillPath, entry.Name())
				break
			}
		}
	}

	if readmePath == "" {
		jsonError(w, "Documentation not found (SKILL.md)", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		jsonError(w, "Failed to read documentation", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{
		"content": string(content),
	})
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/yeasy/ask/internal/cache"
	"github.com/yeasy/ask/internal/config"
)

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
	limitRequestBody(w, r)

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

	if strings.HasPrefix(req.URL, "-") {
		jsonError(w, "Invalid repository URL", http.StatusBadRequest)
		return
	}

	// Validate URL format: must be HTTPS URL or owner/repo shorthand
	if !strings.HasPrefix(req.URL, "https://") && !strings.HasPrefix(req.URL, "http://") {
		// Allow owner/repo shorthand (e.g., "yeasy/awesome-agent-skills")
		parts := strings.SplitN(req.URL, "/", 3)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			jsonError(w, "Repository URL must be an HTTP(S) URL or owner/repo format", http.StatusBadRequest)
			return
		}
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
	limitRequestBody(w, r)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateSkillName(req.Name); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
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
	limitRequestBody(w, r)

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
		if err := validateSkillName(req.Name); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		args = append(args, req.Name)
	}

	cmd := exec.Command(exe, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		jsonError(w, fmt.Sprintf("Sync failed: %s", string(output)), http.StatusInternalServerError)
		return
	}

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

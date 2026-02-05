// Package server provides an embedded HTTP server for the ask web UI.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/yeasy/ask/internal/ui"
)

//go:embed web/*
var webFS embed.FS

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

	// New SkillsLM API routes
	mux.HandleFunc("/api/skills/scan", s.handleSkillScan)
	mux.HandleFunc("/api/skills/import", s.handleSkillImport)
	mux.HandleFunc("/api/skills/files", s.handleSkillFiles) // ?path=...

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

// validateSkillName checks if a skill name is safe (alphanumeric, -, _, .)
func validateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("skill name is required")
	}
	// Allow alphanumeric, dash, underscore, dot, slash (for repo/path)
	// But disallow characters that could be used for shell injection like ; & | $ ` > <
	// Actually, strictly allow only a safe set.
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' && r != '/' && r != '@' {
			return fmt.Errorf("invalid character in skill name: %c", r)
		}
	}
	// Check for directory traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("directory traversal not allowed")
	}
	return nil
}

// corsMiddleware adds CORS headers for development, restricted to localhost
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		if origin != "" {
			// Allow localhost/127.0.0.1
			if strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.HasPrefix(origin, "app://") { // Allow wails/electron type apps if needed
				allowed = true
			}
		} else {
			// No origin usually means same origin or direct request
			allowed = true
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// maxRequestBodySize limits the maximum size of request bodies (1MB)
const maxRequestBodySize = 1 << 20 // 1MB

// limitRequestBody is a helper to limit request body size for POST handlers
func limitRequestBody(r *http.Request) {
	if r.Body != nil {
		r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
	}
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

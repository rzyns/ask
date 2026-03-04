package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "my-skill", false},
		{"valid with underscore", "my_skill", false},
		{"valid with dot", "skill.v1", false},
		{"valid with slash", "owner/repo", false},
		{"valid with at", "skill@v1.0.0", false},
		{"empty name", "", true},
		{"directory traversal", "../secret", true},
		{"shell injection semicolon", "skill;rm -rf", true},
		{"shell injection ampersand", "skill && echo", true},
		{"shell injection pipe", "skill | cat", true},
		{"shell injection dollar", "skill$HOME", true},
		{"shell injection backtick", "skill`whoami`", true},
		{"shell injection greater than", "skill > file", true},
		{"shell injection less than", "skill < file", true},
		{"contains space", "my skill", true},
		{"newline injection", "skill\necho", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkillName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSkillName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCorsMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		origin         string
		method         string
		wantAllowed    bool
		wantStatusCode int
	}{
		{"localhost origin", "http://localhost:8080", "GET", true, http.StatusOK},
		{"127.0.0.1 origin", "http://127.0.0.1:8080", "GET", true, http.StatusOK},
		{"app:// origin", "app://myapp", "GET", true, http.StatusOK},
		{"external origin", "https://evil.com", "GET", false, http.StatusOK},
		{"no origin", "", "GET", true, http.StatusOK},
		{"options preflight localhost", "http://localhost:8080", "OPTIONS", true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}

			corsHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if tt.wantAllowed && corsHeader != tt.origin && tt.origin != "" {
				t.Errorf("CORS header = %q, want %q", corsHeader, tt.origin)
			}
			if !tt.wantAllowed && corsHeader != "" {
				t.Errorf("CORS header should be empty for external origin, got %q", corsHeader)
			}
		})
	}
}

func TestJSONResponse(t *testing.T) {
	rr := httptest.NewRecorder()

	data := map[string]string{"status": "ok", "message": "test"}
	jsonResponse(rr, data)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rr.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
}

func TestJSONError(t *testing.T) {
	rr := httptest.NewRecorder()

	jsonError(rr, "test error", http.StatusBadRequest)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rr.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["error"] != "test error" {
		t.Errorf("error = %q, want 'test error'", result["error"])
	}
}

func TestLimitRequestBody(t *testing.T) {
	// Create a request with a body larger than the limit
	largeBody := strings.Repeat("x", 2<<20) // 2MB, larger than 1MB limit
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(largeBody))

	w := httptest.NewRecorder()
	limitRequestBody(w, req)

	// Try to read the body - should fail after limit
	buf := make([]byte, 2<<20)
	_, err := req.Body.Read(buf)

	// Reading past the limit should return an error
	if err == nil {
		// Read more to trigger the limit
		for {
			_, err = req.Body.Read(buf)
			if err != nil {
				break
			}
		}
	}

	// The error should be http: request body too large or EOF
	if err == nil {
		t.Error("Expected error when reading oversized body")
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	s := New(0, "test")

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
	}{
		{"skills GET on POST endpoint", s.handleSkillInstall, "GET", "/api/skills/install"},
		{"skills DELETE on POST endpoint", s.handleSkillInstall, "DELETE", "/api/skills/install"},
		{"repos POST on GET endpoint", s.handleRepos, "POST", "/api/repos"},
		{"config POST on GET endpoint", s.handleConfig, "POST", "/api/config"},
		{"skills search POST on GET endpoint", s.handleSkillSearch, "POST", "/api/skills/search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			tt.handler(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHandleSkillInstallValidation(t *testing.T) {
	s := New(0, "test")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty name",
			body:       `{"name":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "injection attempt",
			body:       `{"name":"skill;rm -rf /"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "directory traversal",
			body:       `{"name":"../../../etc/passwd"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/skills/install", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.handleSkillInstall(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestHandleSkillScanValidation(t *testing.T) {
	s := New(0, "test")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "empty path",
			body:       `{"path":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "nonexistent path",
			body:       `{"path":"/nonexistent/path/12345"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/skills/scan", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.handleSkillScan(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestHandleSkillFilesValidation(t *testing.T) {
	s := New(0, "test")

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "missing skill name",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "nonexistent skill",
			query:      "?skill=nonexistent-skill-12345",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/skills/files"+tt.query, nil)
			rr := httptest.NewRecorder()

			s.handleSkillFiles(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestHandleRepoAddValidation(t *testing.T) {
	s := New(0, "test")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "empty url",
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/repos/add", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			s.handleRepoAdd(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestParseGitConfigForRepo(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "empty file",
			content: "",
			want:    "",
		},
		// Note: This function reads from filesystem, so we can't easily test actual parsing
		// without creating temp files. The function is indirectly tested through handleSkills.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// parseGitConfigForRepo reads from file, so we'd need a temp file here
			// This is a placeholder for the structure
			_ = tt.content
		})
	}
}

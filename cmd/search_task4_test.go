package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

func captureStdoutForTask4(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDisplaySearchResultsShowsInstallRefTextAndJSON(t *testing.T) {
	repos := []github.Repository{{
		Name:            "grill-me",
		Description:     "roast your code",
		Source:          config.RepoTypeSkillsSH,
		InstallRef:      "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me",
		HTMLURL:         "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me",
		StargazersCount: 5,
		Supported:       true,
	}}

	text := captureStdoutForTask4(t, func() {
		displaySearchResults(repos, nil, "remote", 0, false)
	})
	for _, want := range []string{"INSTALL", "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text output missing %q:\n%s", want, text)
		}
	}

	jsonOut := captureStdoutForTask4(t, func() {
		displaySearchResults(repos, nil, "remote", 0, true)
	})
	for _, want := range []string{`"supported": true`, `"install_ref": "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me"`} {
		if !strings.Contains(jsonOut, want) {
			t.Fatalf("JSON output missing %s:\n%s", want, jsonOut)
		}
	}
}

func TestDisplaySearchResultsShowsUnsupportedReasonTextAndJSON(t *testing.T) {
	repos := []github.Repository{{
		Name:              "mintlify",
		Description:       "docs helper",
		Source:            config.RepoTypeSkillsSH,
		StargazersCount:   5,
		Supported:         false,
		UnsupportedReason: "no native ASK resolver for skills.sh entry",
		PageURL:           "https://skills.sh/mintlify",
	}}

	text := captureStdoutForTask4(t, func() {
		displaySearchResults(repos, nil, "remote", 0, false)
	})
	if !strings.Contains(text, "UNSUPPORTED") || !strings.Contains(text, repos[0].UnsupportedReason) {
		t.Fatalf("text output did not include unsupported reason:\n%s", text)
	}

	jsonOut := captureStdoutForTask4(t, func() {
		displaySearchResults(repos, nil, "remote", 0, true)
	})
	for _, want := range []string{`"supported": false`, `"unsupported_reason": "no native ASK resolver for skills.sh entry"`, `"page_url": "https://skills.sh/mintlify"`} {
		if !strings.Contains(jsonOut, want) {
			t.Fatalf("JSON output missing %s:\n%s", want, jsonOut)
		}
	}
}

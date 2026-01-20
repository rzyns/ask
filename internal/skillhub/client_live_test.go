package skillhub_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/skillhub"
)

func TestSkillHubLive(t *testing.T) {
	client := skillhub.NewClient()

	// Test Search
	fmt.Println("Searching for 'python'...")
	skills, err := client.Search("python")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(skills) == 0 {
		t.Fatal("No skills found for 'python'")
	}
	fmt.Printf("Found %d skills. First: %s (Slug: %s)\n", len(skills), skills[0].Name, skills[0].Slug)

	// Test Resolve
	slug := skills[0].Slug
	fmt.Printf("Resolving slug '%s'...\n", slug)
	url, err := client.Resolve(slug)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	fmt.Printf("Resolved URL: %s\n", url)
	if url == "" {
		t.Fatal("Resolved URL is empty")
	}
	if strings.Contains(url, "#") {
		t.Fatalf("Resolved URL should not contain fragment: %s", url)
	}
}

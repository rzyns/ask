package github

import (
	"encoding/json"
	"testing"
)

func TestSearchResultParsing(t *testing.T) {
	jsonResponse := `{
		"total_count": 1,
		"items": [
			{
				"name": "browser-use",
				"full_name": "browser-use/browser-use",
				"description": "Make websites accessible for AI agents",
				"stargazers_count": 1024,
				"html_url": "https://github.com/browser-use/browser-use"
			}
		]
	}`

	var result SearchResult
	err := json.Unmarshal([]byte(jsonResponse), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("Expected TotalCount 1, got %d", result.TotalCount)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.Name != "browser-use" {
		t.Errorf("Expected name browser-use, got %s", item.Name)
	}
	if item.StargazersCount != 1024 {
		t.Errorf("Expected 1024 stars, got %d", item.StargazersCount)
	}
}

package ui

import (
	"strings"
	"testing"
)

func TestNewSpinner(t *testing.T) {
	description := "Testing spinner..."
	spinner := NewSpinner(description)

	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}

	// Verify spinner was created (can't easily test actual rendering)
	// Just verify it doesn't panic
	_ = spinner.Finish()
}

func TestNewProgressBar(t *testing.T) {
	total := 100
	description := "Testing progress"
	bar := NewProgressBar(total, description)

	if bar == nil {
		t.Fatal("NewProgressBar returned nil")
	}

	// Test updating progress
	for i := 0; i < 10; i++ {
		_ = bar.Add(1)
	}

	_ = bar.Finish()
}

func TestUpdateDescription(t *testing.T) {
	bar := NewProgressBar(10, "Initial description")

	newDescription := "Updated description"
	UpdateDescription(bar, newDescription)

	// Verify it doesn't panic when updating description
	_ = bar.Add(1)
	_ = bar.Finish()
}

func TestProgressBarWithZeroTotal(t *testing.T) {
	// Test edge case with zero total
	bar := NewProgressBar(0, "Empty progress")

	if bar == nil {
		t.Fatal("NewProgressBar returned nil for zero total")
	}

	_ = bar.Finish()
}

func TestSpinnerWithEmptyDescription(t *testing.T) {
	// Test edge case with empty description
	spinner := NewSpinner("")

	if spinner == nil {
		t.Fatal("NewSpinner returned nil for empty description")
	}

	_ = spinner.Finish()
}

func TestProgressBarSequence(t *testing.T) {
	// Test a realistic progress bar usage sequence
	total := 5
	bar := NewProgressBar(total, "Processing items")

	// Simulate processing
	for i := 0; i < total; i++ {
		UpdateDescription(bar, strings.Repeat("Processing item ", i+1))
		_ = bar.Add(1)
	}

	_ = bar.Finish()
}

func TestMultipleProgressBars(t *testing.T) {
	// Test creating multiple progress bars
	bar1 := NewProgressBar(10, "First task")
	bar2 := NewProgressBar(20, "Second task")

	if bar1 == nil || bar2 == nil {
		t.Fatal("Failed to create multiple progress bars")
	}

	// Update both
	_ = bar1.Add(5)
	_ = bar2.Add(10)

	_ = bar1.Finish()
	_ = bar2.Finish()
}

func TestSpinnerAndProgressBar(t *testing.T) {
	// Test using both spinner and progress bar
	spinner := NewSpinner("Initializing...")
	_ = spinner.Add(1)
	_ = spinner.Finish()

	bar := NewProgressBar(3, "Processing")
	for i := 0; i < 3; i++ {
		_ = bar.Add(1)
	}
	_ = bar.Finish()
}

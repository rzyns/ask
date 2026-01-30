package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoctorCommand(t *testing.T) {
	// Test that doctor command is registered
	cmd := doctorCmd
	assert.Equal(t, "doctor", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestDoctorFlags(t *testing.T) {
	// Test --json flag exists
	flag := doctorCmd.Flags().Lookup("json")
	assert.NotNil(t, flag, "doctor command should have --json flag")
	assert.Equal(t, "false", flag.DefValue)
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"ok", "✓"},
		{"warning", "⚠"},
		{"error", "✗"},
		{"unknown", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := getStatusIcon(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckConfiguration(t *testing.T) {
	// Test that checkConfiguration returns a valid result
	result := checkConfiguration()
	assert.Equal(t, "Configuration", result.Category)
	assert.NotEmpty(t, result.Status)
	// Should have at least ask.yaml and ask.lock checks
	assert.GreaterOrEqual(t, len(result.Children), 1)
}

func TestCheckSystem(t *testing.T) {
	// Test that checkSystem returns a valid result
	result := checkSystem()
	assert.Equal(t, "System", result.Category)
	assert.NotEmpty(t, result.Status)
	// Should have git check at minimum
	assert.GreaterOrEqual(t, len(result.Children), 1)
}

func TestCheckSkillsDirectory(t *testing.T) {
	// Test that checkSkillsDirectory returns a valid result
	result := checkSkillsDirectory()
	assert.Equal(t, "Skills Directory", result.Category)
	assert.NotEmpty(t, result.Status)
}

func TestCheckRepositoryCache(t *testing.T) {
	// Test that checkRepositoryCache returns a valid result
	result := checkRepositoryCache()
	assert.Equal(t, "Repository Cache", result.Category)
	assert.NotEmpty(t, result.Status)
}

func TestCheckAgentDirectories(t *testing.T) {
	// Test that checkAgentDirectories returns a valid result
	result := checkAgentDirectories()
	assert.Equal(t, "Agent Directories", result.Category)
	assert.NotEmpty(t, result.Status)
}

func TestDoctorReportStructure(t *testing.T) {
	// Test DoctorReport structure
	report := DoctorReport{
		Version: "1.0",
		Results: []DoctorResult{
			{
				Category: "Test",
				Status:   "ok",
				Children: []CheckItem{
					{Name: "item1", Status: "ok", Message: "test"},
				},
			},
		},
		Summary: DoctorSummary{
			TotalChecks:   1,
			PassedChecks:  1,
			WarningChecks: 0,
			FailedChecks:  0,
		},
	}

	assert.Equal(t, "1.0", report.Version)
	assert.Len(t, report.Results, 1)
	assert.Equal(t, 1, report.Summary.TotalChecks)
	assert.Equal(t, 1, report.Summary.PassedChecks)
}

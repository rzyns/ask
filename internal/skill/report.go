package skill

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// GenerateReport generates a report in the specified format ("md", "html", or "json")
func GenerateReport(result *CheckResult, format string) (string, error) {
	switch strings.ToLower(format) {
	case "md", "markdown":
		return generateMarkdown(result), nil
	case "html", "htm":
		return generateHTML(result), nil
	case "json":
		return generateJSON(result), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func generateMarkdown(result *CheckResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Security Report: %s\n\n", result.SkillName))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	criticals, warnings, infos := countSeverities(result.Findings)
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Severity | Count |\n")
	sb.WriteString("| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| 🔴 Critical | %d |\n", criticals))
	sb.WriteString(fmt.Sprintf("| 🟡 Warning | %d |\n", warnings))
	sb.WriteString(fmt.Sprintf("| 🔵 Info | %d |\n", infos))
	sb.WriteString("\n")

	if len(result.Findings) == 0 {
		sb.WriteString("✅ **No issues found.** Skill appears safe.\n")
		return sb.String()
	}

	sb.WriteString("## Detailed Findings\n\n")
	for _, f := range result.Findings {
		var icon string
		switch f.Severity {
		case SeverityCritical:
			icon = "🔴"
		case SeverityWarning:
			icon = "🟡"
		default:
			icon = "🔵"
		}

		sb.WriteString(fmt.Sprintf("### %s %s\n", icon, f.Description))
		sb.WriteString(fmt.Sprintf("- **Rule ID:** `%s`\n", f.RuleID))
		sb.WriteString(fmt.Sprintf("- **Location:** `%s:%d`\n", f.File, f.Line))
		sb.WriteString("```\n")
		sb.WriteString(f.Match)
		sb.WriteString("\n```\n\n")
	}

	return sb.String()
}

func generateHTML(result *CheckResult) string {
	const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Security Report: {{.SkillName}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; max-width: 900px; margin: 0 auto; padding: 20px; }
        h1 { border-bottom: 2px solid #eee; padding-bottom: 10px; }
        .summary { display: flex; gap: 20px; margin-bottom: 30px; }
        .card { background: #f9f9f9; padding: 15px; border-radius: 8px; flex: 1; text-align: center; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .count { font-size: 2em; font-weight: bold; display: block; }
        .critical { color: #d73a49; }
        .warning { color: #b08800; }
        .info { color: #0366d6; }
        .finding { border: 1px solid #e1e4e8; border-radius: 6px; margin-bottom: 16px; overflow: hidden; }
        .finding-header { padding: 10px 15px; background: #f6f8fa; border-bottom: 1px solid #e1e4e8; display: flex; justify-content: space-between; align-items: center; }
        .badge { padding: 4px 8px; border-radius: 12px; font-size: 0.85em; font-weight: 600; color: white; }
        .bg-critical { background-color: #d73a49; }
        .bg-warning { background-color: #dbab09; color: #24292e; }
        .bg-info { background-color: #0366d6; }
        .finding-body { padding: 15px; }
        .location { color: #586069; font-size: 0.9em; margin-bottom: 10px; }
        code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace; }
        pre { background: #f6f8fa; padding: 10px; border-radius: 4px; overflow-x: auto; margin-top: 10px; }
    </style>
</head>
<body>
    <h1>Security Report: {{.SkillName}}</h1>
    <p>Generated on {{.Date}}</p>

    <div class="summary">
        <div class="card critical">
            <span class="count">{{.Stats.Critical}}</span>
            Critical
        </div>
        <div class="card warning">
            <span class="count">{{.Stats.Warning}}</span>
            Warning
        </div>
        <div class="card info">
            <span class="count">{{.Stats.Info}}</span>
            Info
        </div>
    </div>

    {{if not .Findings}}
    <div style="text-align: center; padding: 40px; background: #e6fffa; border-radius: 8px; color: #006644;">
        <h2>✅ No issues found</h2>
        <p>The skill appears to be safe based on current checks.</p>
    </div>
    {{end}}

    {{range .FormattedFindings}}
    <div class="finding">
        <div class="finding-header">
            <strong>{{.Description}}</strong>
            <span class="badge bg-{{.SeverityClass}}">{{.Severity}}</span>
        </div>
        <div class="finding-body">
            <div class="location">
                📍 {{.File}}:{{.Line}} <span style="margin-left: 10px; color: #999;">Rule: {{.RuleID}}</span>
            </div>
            <pre><code>{{.Match}}</code></pre>
        </div>
    </div>
    {{end}}
</body>
</html>`

	type FindingView struct {
		Finding
		SeverityClass string
	}

	criticals, warnings, infos := countSeverities(result.Findings)

	data := struct {
		SkillName         string
		Date              string
		Stats             struct{ Critical, Warning, Info int }
		Findings          []Finding
		FormattedFindings []FindingView
	}{
		SkillName: result.SkillName,
		Date:      time.Now().Format("2006-01-02 15:04:05"),
		Stats:     struct{ Critical, Warning, Info int }{criticals, warnings, infos},
		Findings:  result.Findings,
	}

	for _, f := range result.Findings {
		var class string
		switch f.Severity {
		case SeverityCritical:
			class = "critical"
		case SeverityWarning:
			class = "warning"
		default:
			class = "info"
		}
		data.FormattedFindings = append(data.FormattedFindings, FindingView{
			Finding:       f,
			SeverityClass: class,
		})
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Sprintf("Error generating HTML: %v", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return sb.String()
}

func countSeverities(findings []Finding) (int, int, int) {
	c, w, i := 0, 0, 0
	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			c++
		case SeverityWarning:
			w++
		case SeverityInfo:
			i++
		}
	}
	return c, w, i
}

func generateJSON(result *CheckResult) string {
	// Add timestamp to the result for JSON output since CheckResult doesn't have it by default
	type JSONResult struct {
		*CheckResult
		Timestamp string `json:"timestamp"`
	}

	output := JSONResult{
		CheckResult: result,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"failed to marshal json: %v\"}", err)
	}
	return string(jsonData)
}

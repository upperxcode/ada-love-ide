package commands

import (
	"context"
	"fmt"
	"strings"

	commands "github.com/upperxcode/ada-commands"
)

// HealthStatus representa o status de um componente
type HealthStatus struct {
	Component string
	Status    string // "ok", "warning", "critical"
	Details   string
	ItemCount int
}

// HealthCheck defines a health check function
type HealthCheck struct {
	Component string
	Check     func() (status, details string, count int)
}

// HealthCommand implements /health to check system health status.
type HealthCommand struct {
	checks []HealthCheck
}

// NewHealthCommand creates a new HealthCommand with health checks.
func NewHealthCommand(checks ...HealthCheck) *HealthCommand {
	return &HealthCommand{checks: checks}
}

func (c *HealthCommand) Name() string { return "health" }
func (c *HealthCommand) Description() string {
	return "Check system health status: /health"
}
func (c *HealthCommand) Category() commands.CommandCategory { return commands.CategoryChat }

func (c *HealthCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	var statuses []HealthStatus

	for _, check := range c.checks {
		status, details, count := check.Check()
		statuses = append(statuses, HealthStatus{
			Component: check.Component,
			Status:    status,
			Details:   details,
			ItemCount: count,
		})
	}

	// Build output with two-column layout
	var sb strings.Builder

	// Header
	sb.WriteString("## 🏥 System Health Check\n\n")
	sb.WriteString("---\n\n")

	// Two-column layout
	sb.WriteString("| Component | Status | Details |\n")
	sb.WriteString("|-----------|--------|--------|\n")

	criticalCount := 0
	warningCount := 0
	okCount := 0

	for _, s := range statuses {
		var statusEmoji string
		switch s.Status {
		case "critical":
			statusEmoji = "🔴"
			criticalCount++
		case "warning":
			statusEmoji = "🟡"
			warningCount++
		default:
			statusEmoji = "🟢"
			okCount++
		}

		// Escape pipe characters in details
		details := strings.ReplaceAll(s.Details, "|", "\\|")
		if s.ItemCount > 0 {
			details = fmt.Sprintf("%s (%d)", details, s.ItemCount)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s %s | %s |\n",
			s.Component,
			statusEmoji,
			s.Status,
			details,
		))
	}

	// Summary section
	sb.WriteString("\n---\n\n")
	sb.WriteString("## 📊 Summary\n\n")

	if criticalCount > 0 {
		sb.WriteString(fmt.Sprintf("🔴 **Critical**: %d component(s) need attention\n", criticalCount))
	}
	if warningCount > 0 {
		sb.WriteString(fmt.Sprintf("🟡 **Warning**: %d component(s) have recommendations\n", warningCount))
	}
	if okCount > 0 {
		sb.WriteString(fmt.Sprintf("🟢 **Healthy**: %d component(s) OK\n", okCount))
	}

	// Overall status with color
	sb.WriteString("\n### Overall Status\n\n")
	if criticalCount > 0 {
		sb.WriteString("> **<span style=\"color:#ef4444\">🔴 CRITICAL</span>**\n")
	} else if warningCount > 0 {
		sb.WriteString("> **<span style=\"color:#eab308\">🟡 WARNING</span>**\n")
	} else {
		sb.WriteString("> **<span style=\"color:#22c55e\">🟢 HEALTHY</span>**\n")
	}

	return sb.String(), nil
}

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	commands "github.com/upperxcode/ada-commands"
	executor "github.com/upperxcode/ada-executor"
)

// ReadCommand implements /read for reading file contents.
type ReadCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewReadCommand creates a new ReadCommand.
func NewReadCommand(executor *executor.TaskExecutor, workspace string) *ReadCommand {
	return &ReadCommand{executor: executor, workspace: workspace}
}

func (c *ReadCommand) Name() string { return "read" }
func (c *ReadCommand) Description() string {
	return "Read a file: /read <path> [lines] - e.g., /read main.go 10-50"
}
func (c *ReadCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *ReadCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	if len(cmdCtx.Args) == 0 {
		return "Usage: /read <path> [lines]", nil
	}

	path := cmdCtx.Args[0]

	// Resolve path relative to workspace
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.workspace, path)
	}

	// Check for line range specification
	var lineRange string
	if len(cmdCtx.Args) > 1 {
		lineRange = cmdCtx.Args[1]
	}

	content, err := c.executor.ReadFile(ctx, path)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	// Apply line range if specified
	result := string(content)
	if lineRange != "" {
		result, err = filterLines(result, lineRange)
		if err != nil {
			return fmt.Sprintf("Error filtering lines: %v", err), nil
		}
	}

	// Format output with line numbers
	lines := strings.Split(result, "\n")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s (%d lines)\n", path, len(lines)))
	for i, line := range lines {
		sb.WriteString(fmt.Sprintf("%4d: %s\n", i+1, line))
	}

	return sb.String(), nil
}

// WriteCommand implements /write for writing file contents.
type WriteCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewWriteCommand creates a new WriteCommand.
func NewWriteCommand(executor *executor.TaskExecutor, workspace string) *WriteCommand {
	return &WriteCommand{executor: executor, workspace: workspace}
}

func (c *WriteCommand) Name() string { return "write" }
func (c *WriteCommand) Description() string {
	return "Write to a file: /write <path> <content> - Content can be quoted"
}
func (c *WriteCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *WriteCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	if len(cmdCtx.Args) < 2 {
		return "Usage: /write <path> <content>\nUse quotes for multi-word content: /write file.txt \"Hello World\"", nil
	}

	path := cmdCtx.Args[0]
	// Join remaining args as content (handles quoted strings)
	content := strings.Join(cmdCtx.Args[1:], " ")

	// Resolve path relative to workspace
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.workspace, path)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err), nil
	}

	result, err := c.executor.WriteFile(ctx, path, []byte(content))
	if err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File written: %s\n", path))
	if result.BackupPath != "" {
		sb.WriteString(fmt.Sprintf("Backup: %s\n", result.BackupPath))
	}
	sb.WriteString(fmt.Sprintf("Modified: %v\n", result.Modified))
	sb.WriteString(fmt.Sprintf("Duration: %v", result.Duration))

	return sb.String(), nil
}

// ShellCommand implements /shell for executing shell commands.
type ShellCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewShellCommand creates a new ShellCommand.
func NewShellCommand(executor *executor.TaskExecutor, workspace string) *ShellCommand {
	return &ShellCommand{executor: executor, workspace: workspace}
}

func (c *ShellCommand) Name() string { return "shell" }
func (c *ShellCommand) Description() string {
	return "Execute a shell command: /shell <command> [args...] - e.g., /shell ls -la"
}
func (c *ShellCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *ShellCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	if len(cmdCtx.Args) == 0 {
		return "Usage: /shell <command> [args...]", nil
	}

	// Set default timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := cmdCtx.Args[0]
	args := cmdCtx.Args[1:]

	result, err := c.executor.ExecuteCommand(ctx, c.workspace, cmd, args)
	if err != nil {
		return fmt.Sprintf("Command error:\n%s\nStderr: %s", result.Stdout, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString("Command output:\n")
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		if result.Stdout != "" {
			sb.WriteString("\n")
		}
		sb.WriteString("Stderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nExit code: %d\n", result.ExitCode))
	sb.WriteString(fmt.Sprintf("Duration: %v", result.Duration))

	return sb.String(), nil
}

// PlanCommand implements /plan for generating development plans.
type PlanCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewPlanCommand creates a new PlanCommand.
func NewPlanCommand(executor *executor.TaskExecutor, workspace string) *PlanCommand {
	return &PlanCommand{executor: executor, workspace: workspace}
}

func (c *PlanCommand) Name() string { return "plan" }
func (c *PlanCommand) Description() string {
	return "Generate a development plan: /plan <requirement> - e.g., /plan create user auth system"
}
func (c *PlanCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *PlanCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	if len(cmdCtx.Args) == 0 {
		return "Usage: /plan <requirement>\nExample: /plan create a REST API for user management", nil
	}

	requirement := strings.Join(cmdCtx.Args, " ")

	// Generate a structured plan based on the requirement
	plan := c.generatePlan(requirement)

	return plan, nil
}

// generatePlan creates a structured development plan.
func (c *PlanCommand) generatePlan(requirement string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Development Plan: %s\n\n", requirement))
	sb.WriteString("## Overview\n")
	sb.WriteString(fmt.Sprintf("This plan addresses the requirement: %s\n\n", requirement))

	sb.WriteString("## Implementation Steps\n\n")
	sb.WriteString("### 1. Planning Phase\n")
	sb.WriteString("- Define scope and boundaries\n")
	sb.WriteString("- Identify dependencies and prerequisites\n")
	sb.WriteString("- Create technical specification\n\n")

	sb.WriteString("### 2. Architecture Design\n")
	sb.WriteString("- Design system architecture\n")
	sb.WriteString("- Define data models and schemas\n")
	sb.WriteString("- Plan API endpoints (if applicable)\n")
	sb.WriteString("- Security considerations\n\n")

	sb.WriteString("### 3. Implementation Phase\n")
	sb.WriteString("- Set up project structure\n")
	sb.WriteString("- Implement core components\n")
	sb.WriteString("- Write tests\n")
	sb.WriteString("- Code review and refactoring\n\n")

	sb.WriteString("### 4. Testing Phase\n")
	sb.WriteString("- Unit tests\n")
	sb.WriteString("- Integration tests\n")
	sb.WriteString("- Manual QA\n\n")

	sb.WriteString("### 5. Deployment Phase\n")
	sb.WriteString("- Build production artifacts\n")
	sb.WriteString("- Deploy to staging\n")
	sb.WriteString("- Production deployment\n")
	sb.WriteString("- Post-deployment monitoring\n\n")

	sb.WriteString("## Files to Create/Modify\n")
	sb.WriteString("- Project configuration files\n")
	sb.WriteString("- Source code files\n")
	sb.WriteString("- Test files\n")
	sb.WriteString("- Documentation\n\n")

	sb.WriteString("## Estimated Timeline\n")
	sb.WriteString("- Planning: 1-2 hours\n")
	sb.WriteString("- Implementation: 1-3 days (depending on complexity)\n")
	sb.WriteString("- Testing: 1 day\n")
	sb.WriteString("- Deployment: 1-2 hours\n\n")

	sb.WriteString("## Commands to Use\n")
	sb.WriteString("- `/shell` - Execute build/test commands\n")
	sb.WriteString("- `/write` - Create new files\n")
	sb.WriteString("- `/read` - Review existing files\n")
	sb.WriteString("- `/shell git` - Version control operations\n")

	return sb.String()
}

// filterLines filters content by line range (e.g., "10-50" or "1-10").
func filterLines(content, lineRange string) (string, error) {
	lines := strings.Split(content, "\n")
	parts := strings.Split(lineRange, "-")

	if len(parts) != 2 {
		return "", fmt.Errorf("invalid line range format, use start-end (e.g., 10-50)")
	}

	start := 0
	end := len(lines)

	fmt.Sscanf(parts[0], "%d", &start)
	start = start - 1 // Convert to 0-indexed

	var endVal int
	fmt.Sscanf(parts[1], "%d", &endVal)
	end = endVal

	if start < 0 || start >= len(lines) {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start >= end {
		return "", fmt.Errorf("invalid range: start must be less than end")
	}

	return strings.Join(lines[start:end], "\n"), nil
}

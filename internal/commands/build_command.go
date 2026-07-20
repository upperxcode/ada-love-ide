package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	commands "github.com/upperxcode/ada-commands"
	executor "github.com/upperxcode/ada-executor"
)

// BuildCommand implements /build for compiling Go projects.
type BuildCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewBuildCommand creates a new BuildCommand with workspace directory.
func NewBuildCommand(executor *executor.TaskExecutor, workspace string) *BuildCommand {
	return &BuildCommand{executor: executor, workspace: workspace}
}

func (c *BuildCommand) Name() string { return "build" }
func (c *BuildCommand) Description() string {
	return "Build the current project: /build [args] - e.g., /build -o app"
}
func (c *BuildCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *BuildCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	args := cmdCtx.Args
	if len(args) == 0 {
		args = []string{"./..."}
	}

	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "go", args)
	if err != nil {
		return fmt.Sprintf("Build error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString("Build completed\n")
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

// TestCommand implements /test for running Go tests.
type TestCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewTestCommand creates a new TestCommand with workspace directory.
func NewTestCommand(executor *executor.TaskExecutor, workspace string) *TestCommand {
	return &TestCommand{executor: executor, workspace: workspace}
}

func (c *TestCommand) Name() string { return "test" }
func (c *TestCommand) Description() string {
	return "Run tests: /test [args] - e.g., /test -v ./..."
}
func (c *TestCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *TestCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	args := cmdCtx.Args
	if len(args) == 0 {
		args = []string{"./..."}
	}

	// Set timeout for tests
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "go", args)
	if err != nil {
		return fmt.Sprintf("Test error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString("Tests completed\n")
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

// RunCommand implements /run for running Go programs.
type RunCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewRunCommand creates a new RunCommand with workspace directory.
func NewRunCommand(executor *executor.TaskExecutor, workspace string) *RunCommand {
	return &RunCommand{executor: executor, workspace: workspace}
}

func (c *RunCommand) Name() string { return "run" }
func (c *RunCommand) Description() string {
	return "Run the main program: /run [args] - e.g., /run main.go"
}
func (c *RunCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *RunCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	args := cmdCtx.Args
	if len(args) == 0 {
		args = []string{"main.go"}
	}

	// Set timeout for running
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "go", append([]string{"run"}, args...))
	if err != nil {
		return fmt.Sprintf("Run error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString("Program output\n")
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

// IndexCommand implements /index for indexing code directories using ada-code-indexer binary.
type IndexCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewIndexCommand creates a new IndexCommand with workspace directory.
func NewIndexCommand(executor *executor.TaskExecutor, workspace string) *IndexCommand {
	return &IndexCommand{executor: executor, workspace: workspace}
}

func (c *IndexCommand) Name() string { return "index" }
func (c *IndexCommand) Description() string {
	return "Index a code directory: /index <path> [max-files] - e.g., /index . 1000"
}
func (c *IndexCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *IndexCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	if len(cmdCtx.Args) == 0 {
		return "Usage: /index <path> [max-files]", nil
	}

	path := cmdCtx.Args[0]
	maxFiles := "10000"
	if len(cmdCtx.Args) > 1 {
		maxFiles = cmdCtx.Args[1]
	}

	// Use ada-executor to run the indexer binary
	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "./indexer", []string{
		"--watch", path,
		"--max-files", maxFiles,
		"--no-server",
	})

	if err != nil {
		return fmt.Sprintf("Index error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString("Index completed\n")
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

// SearchCommand implements /search for code symbol search using ada-code-indexer binary.
type SearchCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewSearchCommand creates a new SearchCommand with workspace directory.
func NewSearchCommand(executor *executor.TaskExecutor, workspace string) *SearchCommand {
	return &SearchCommand{executor: executor, workspace: workspace}
}

func (c *SearchCommand) Name() string { return "search" }
func (c *SearchCommand) Description() string {
	return "Search symbols: /search <term> - e.g., /search skill"
}
func (c *SearchCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *SearchCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	if len(cmdCtx.Args) == 0 {
		return "Usage: /search <term>", nil
	}

	term := strings.Join(cmdCtx.Args, " ")

	// Use ada-executor to run the indexer binary with search
	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "./indexer", []string{
		"--no-server",
		"--json",
		"--query", term,
	})

	if err != nil {
		return fmt.Sprintf("Search error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for '%s':\n\n", term))
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

// DocCommand implements /doc for API documentation.
type DocCommand struct {
	executor  *executor.TaskExecutor
	workspace string
}

// NewDocCommand creates a new DocCommand with workspace directory.
func NewDocCommand(executor *executor.TaskExecutor, workspace string) *DocCommand {
	return &DocCommand{executor: executor, workspace: workspace}
}

func (c *DocCommand) Name() string { return "doc" }
func (c *DocCommand) Description() string {
	return "Get documentation: /doc <package|all> [kind] - e.g., /doc backend function"
}
func (c *DocCommand) Category() commands.CommandCategory { return commands.CategoryTool }

func (c *DocCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Check if workspace is configured
	if c.workspace == "" || c.workspace == "." {
		return "Error: No workspace configured. Please select a workspace first.", nil
	}

	if len(cmdCtx.Args) == 0 {
		return "Usage: /doc <package|all> [kind]", nil
	}

	pkg := cmdCtx.Args[0]

	// Use ada-executor to run the indexer binary with doc
	result, err := c.executor.ExecuteCommand(ctx, c.workspace, "./indexer", []string{
		"--no-server",
		"--json",
		"--query", pkg,
	})

	if err != nil {
		return fmt.Sprintf("Doc error: %v\nStderr: %s", err, result.Stderr), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Documentation for '%s':\n\n", pkg))
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		sb.WriteString("\nStderr: ")
		sb.WriteString(result.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nDuration: %v", result.Duration))
	return sb.String(), nil
}

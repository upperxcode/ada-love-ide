package commands_test

import (
	"context"
	"testing"

	commands "github.com/upperxcode/ada-commands"
)

func TestCommandCategories(t *testing.T) {
	router := commands.NewCommandRouter()

	// Register sample commands
	router.Register(&testChatCommand{name: "test_chat"})
	router.Register(&testToolCommand{name: "test_tool"})
	router.Register(&testInternalCommand{name: "test_internal"})

	// Test ListChatCommands
	chatCmds := router.ListChatCommands()
	if len(chatCmds) != 1 {
		t.Errorf("expected 1 chat command, got %d", len(chatCmds))
	}
	if chatCmds[0].Name() != "test_chat" {
		t.Errorf("expected test_chat, got %s", chatCmds[0].Name())
	}

	// Test ListCommandsByCategory
	toolCmds := router.ListCommandsByCategory(commands.CategoryTool)
	if len(toolCmds) != 1 {
		t.Errorf("expected 1 tool command, got %d", len(toolCmds))
	}

	internalCmds := router.ListCommandsByCategory(commands.CategoryInternal)
	if len(internalCmds) != 1 {
		t.Errorf("expected 1 internal command, got %d", len(internalCmds))
	}
}

// Test command implementations
type testChatCommand struct{ name string }

func (c *testChatCommand) Name() string                       { return c.name }
func (c *testChatCommand) Description() string                { return "chat command" }
func (c *testChatCommand) Category() commands.CommandCategory { return commands.CategoryChat }
func (c *testChatCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	return "", nil
}

type testToolCommand struct{ name string }

func (c *testToolCommand) Name() string                       { return c.name }
func (c *testToolCommand) Description() string                { return "tool command" }
func (c *testToolCommand) Category() commands.CommandCategory { return commands.CategoryTool }
func (c *testToolCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	return "", nil
}

type testInternalCommand struct{ name string }

func (c *testInternalCommand) Name() string                       { return c.name }
func (c *testInternalCommand) Description() string                { return "internal command" }
func (c *testInternalCommand) Category() commands.CommandCategory { return commands.CategoryInternal }
func (c *testInternalCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	return "", nil
}

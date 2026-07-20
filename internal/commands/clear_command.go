package commands

import (
	"context"
	"fmt"

	"ada-love-ide/internal/db"

	commands "github.com/upperxcode/ada-commands"
)

// ClearCommand implements /clear with access to the database store.
type ClearCommand struct {
	db        *db.Store
	sessionID string
}

// NewClearCommand creates a new ClearCommand with database access.
func NewClearCommand(db *db.Store, sessionID string) *ClearCommand {
	return &ClearCommand{
		db:        db,
		sessionID: sessionID,
	}
}

func (c *ClearCommand) Name() string { return "clear" }
func (c *ClearCommand) Description() string {
	return "Clears the current chat session history"
}
func (c *ClearCommand) Category() commands.CommandCategory { return commands.CategoryChat }

func (c *ClearCommand) Execute(ctx context.Context, cmdCtx *commands.CommandContext) (string, error) {
	// Use the sessionID from the command context (passed by the orchestrator)
	sessionID := cmdCtx.SessionID

	// Also check if we have a fallback sessionID stored
	if sessionID == "" && c.sessionID != "" {
		sessionID = c.sessionID
	}

	if sessionID == "" {
		return "Error: no session ID provided. Please start a chat first.", nil
	}

	err := c.db.Sessions().DeleteMessages(ctx, sessionID)
	if err != nil {
		return fmt.Sprintf("Error clearing chat: %v", err), nil
	}

	return "Chat cleared", nil
}

package commands

import (
	"context"
	"fmt"

	adaCommands "github.com/upperxcode/ada-commands"
)

// SyncSpecCommand regenera .AGENTS.md e .spec-wizard/ do Spec Wizard no workspace.
type SyncSpecCommand struct {
	SyncFn func(workspacePath string) error
}

func (c *SyncSpecCommand) Name() string        { return "sync-spec" }
func (c *SyncSpecCommand) Description() string { return "Regenera .AGENTS.md e .spec-wizard/ do Spec Wizard no workspace atual" }
func (c *SyncSpecCommand) Category() adaCommands.CommandCategory { return adaCommands.CategoryChat }
func (c *SyncSpecCommand) Execute(_ context.Context, cmdCtx *adaCommands.CommandContext) (string, error) {
	if c.SyncFn == nil {
		return "❌ Sync function not configured", nil
	}
	if cmdCtx.Workspace == "" {
		return "❌ Nenhum workspace ativo", nil
	}
	if err := c.SyncFn(cmdCtx.Workspace); err != nil {
		return fmt.Sprintf("❌ %v", err), nil
	}
	return "✅ .AGENTS.md e .spec-wizard/ atualizados!", nil
}

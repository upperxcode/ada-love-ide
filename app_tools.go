package main

import (
	"context"
	"fmt"

	"ada-love-ide/internal/config/tool"

	storage "github.com/ada-love-ai/storage/storage"
	adaCommands "github.com/upperxcode/ada-commands"
)

// GetToolProfiles retorna todos os perfis de ferramentas do banco SQL.
func (a *App) GetToolProfiles() []tool.ToolProfile {
	return a.eng.DB.ListProfiles()
}

// CreateToolProfile cria um novo perfil no banco SQL.
func (a *App) CreateToolProfile(name, color, icon string) tool.ToolProfile {
	ctx := context.Background()
	sp := &storage.ToolProfile{
		Name:  name,
		Color: color,
		Icon:  icon,
	}
	if err := a.eng.DB.ToolProfiles().CreateProfile(ctx, sp); err != nil {
		fmt.Printf("[Backend] CreateToolProfile error: %v\n", err)
		return tool.ToolProfile{}
	}
	return tool.ToolProfile{
		ID:    int(sp.ID),
		Name:  sp.Name,
		Color: sp.Color,
		Icon:  sp.Icon,
		Tools: []string{},
	}
}

// DeleteToolProfile remove um perfil do banco SQL.
func (a *App) DeleteToolProfile(id int) bool {
	ctx := context.Background()
	if err := a.eng.DB.ToolProfiles().DeleteProfile(ctx, int64(id)); err != nil {
		fmt.Printf("[Backend] DeleteToolProfile error: %v\n", err)
		return false
	}
	return true
}

// ToggleProfileTool liga/desliga uma ferramenta em um perfil no banco SQL.
func (a *App) ToggleProfileTool(profileID int, toolName string, enabled bool) bool {
	ctx := context.Background()
	if enabled {
		if err := a.eng.DB.ToolProfiles().AddTool(ctx, &storage.ToolProfileTool{
			ProfileID: int64(profileID),
			ToolName:  toolName,
		}); err != nil {
			fmt.Printf("[Backend] ToggleProfileTool add error: %v\n", err)
			return false
		}
	} else {
		if err := a.eng.DB.ToolProfiles().RemoveTool(ctx, int64(profileID), toolName); err != nil {
			fmt.Printf("[Backend] ToggleProfileTool remove error: %v\n", err)
			return false
		}
	}
	return true
}

// GetAvailableTools lista ferramentas disponíveis via ada-commands router.
func (a *App) GetAvailableTools() []tool.ToolUIInfo {
	// Use the router to get all registered commands as tools
	cmds := a.eng.Router.ListCommandsByCategory(adaCommands.CategoryTool)

	out := make([]tool.ToolUIInfo, 0, len(cmds))
	for _, cmd := range cmds {
		out = append(out, tool.ToolUIInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Category:    "tool",
			Enabled:     true, // All registered commands are enabled
		})
	}
	return out
}

// ToggleTool registra ou remove um comando do router.
func (a *App) ToggleTool(toolName string, enabled bool) bool {
	router := a.eng.Router
	if router == nil {
		return false
	}

	// Check if command exists in the list
	cmds := router.ListCommands()
	var exists bool
	for _, cmd := range cmds {
		if cmd.Name() == toolName {
			exists = true
			break
		}
	}

	if enabled && !exists {
		// For now, we cannot dynamically create commands without executor
		// This is a limitation - tools need to be pre-registered in engine.go
		fmt.Printf("[Backend] Cannot enable unregistered tool: %s\n", toolName)
		return false
	}

	if !enabled && exists {
		// Note: CommandRouter doesn't support unregistering
		// In a real implementation, we would need to add Unregister method
		fmt.Printf("[Backend] Cannot disable tool (unregister not supported): %s\n", toolName)
		return false
	}

	return true
}

// ListChatCommands retorna apenas comandos visíveis no chat.
func (a *App) ListChatCommands() []tool.ToolUIInfo {
	cmds := a.eng.Router.ListCommandsByCategory(adaCommands.CategoryChat)

	out := make([]tool.ToolUIInfo, 0, len(cmds))
	for _, cmd := range cmds {
		out = append(out, tool.ToolUIInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Category:    "chat",
			Enabled:     true,
		})
	}
	return out
}

// ListAllCommands retorna todos os comandos registrados.
func (a *App) ListAllCommands() []tool.ToolUIInfo {
	cmds := a.eng.Router.ListCommands()

	out := make([]tool.ToolUIInfo, 0, len(cmds))
	for _, cmd := range cmds {
		cat := "tool"
		if cmd.Category() == adaCommands.CategoryChat {
			cat = "chat"
		} else if cmd.Category() == adaCommands.CategoryInternal {
			cat = "internal"
		}
		out = append(out, tool.ToolUIInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Category:    cat,
			Enabled:     true,
		})
	}
	return out
}

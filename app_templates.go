package main

import (
	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/workspace"
)

// Workspace templates

// GetWorkspaceTemplates lista templates mockados.
func (a *App) GetWorkspaceTemplates() []workspace.WorkspaceTemplate {
	return a.eng.DB.ListTemplates()
}

// SaveWorkspaceTemplate cria/atualiza template.
func (a *App) SaveWorkspaceTemplate(t workspace.WorkspaceTemplate) {
	if t.ID == 0 {
		t.ID = a.eng.DB.NextTemplateID()
	}
	a.eng.DB.PutTemplate(t)
}

// DeleteWorkspaceTemplate remove template.
func (a *App) DeleteWorkspaceTemplate(id int) { a.eng.DB.DeleteTemplate(id) }

// ── MCP ──────────────────────────────────────────────────────────

// GetPredefinedConnections lista conexões MCP predefinidas.
func (a *App) GetPredefinedConnections() []mcp.ConnectionDefinition {
	return []mcp.ConnectionDefinition{
		{Name: "Ada", Type: "ada", Command: "ada", Description: "Ada engine local", Icon: "🤖"},
		{Name: "CLI", Type: "cli", Command: "bash", Description: "Shell local", Icon: "💻"},
		{Name: "REST", Type: "rest", Command: "", Description: "API REST genérica", Icon: "🌐"},
		{Name: "MCP", Type: "mcp", Command: "npx", Description: "Servidor MCP", Icon: "🔌"},
	}
}

// TestConnection testa conexão (mock ok).
func (a *App) TestConnection(connectionType, connectionName, connectionConfig string) mcp.ConnectionTestResult {
	return mcp.ConnectionTestResult{
		Success:   true,
		Message:   "mock: conexão OK",
		LatencyMS: 12,
	}
}

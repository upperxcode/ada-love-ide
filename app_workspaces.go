package main

import (
	"ada-love-ide/internal/config/workspace"
	"time"
)

// GetWorkspaces retorna todos os workspaces.
func (a *App) GetWorkspaces() []workspace.WorkspaceConfig {
	return a.eng.DB.ListWorkspaces()
}

// GetWorkspace recupera um workspace por path.
func (a *App) GetWorkspace(path string) (*workspace.WorkspaceConfig, error) {
	ws, err := a.eng.DB.GetWorkspace(path)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

// SaveWorkspace cria/atualiza um workspace.
func (a *App) SaveWorkspace(ws workspace.WorkspaceConfig) {
	if ws.Path == "" {
		ws.Path = "workspace-" + time.Now().Format("20060102150405")
	}
	a.eng.DB.AddWorkspace(ws)
}

// DeleteWorkspace remove um workspace por path.
func (a *App) DeleteWorkspace(path string) { a.eng.DB.DeleteWorkspace(path) }

// SetActiveWorkspace marca o workspace atual.
func (a *App) SetActiveWorkspace(path string) { a.eng.DB.SetActiveWorkspace(path) }

// ToggleWorkspace alterna o flag enabled.
func (a *App) ToggleWorkspace(path string) {
	// Reflete uma troca de flag na memória.
	list := a.eng.DB.ListWorkspaces()
	for _, w := range list {
		if w.Path == path {
			w.Enabled = !w.Enabled
			a.eng.DB.AddWorkspace(w)
			return
		}
	}
}

// SetWorkspaces sobrescreve todos os workspaces.
func (a *App) SetWorkspaces(list []workspace.WorkspaceConfig) {
	a.eng.DB.SetWorkspaces(list)
}

// AddWorkspace adiciona um novo workspace.
func (a *App) AddWorkspace(title, path, personality, routingRules string) {
	ws := workspace.New(title, path)
	ws.Personality = personality
	ws.RoutingRules = routingRules
	a.eng.DB.AddWorkspace(ws)
}

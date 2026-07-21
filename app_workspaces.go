package main

import (
	"context"
	"fmt"
	"time"

	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/worker"
	"ada-love-ide/internal/config/workspace"
	"ada-love-ide/internal/db"
	"ada-love-ide/internal/specwizardmgr"

	storage "github.com/ada-love-ai/storage/storage"
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

	if ws.SpecWizardID != "" {
		ensureSWMCP(a.eng.DB, ws)
		SyncSpecToWorkspace(a.eng.DB, ws)
	}
}

func ensureSWMCP(store *db.Store, ws workspace.WorkspaceConfig) {
	mcpName := "sw-" + ws.Title
	if _, ok := store.ListMCPServers()[mcpName]; ok {
		return
	}

	dir := ws.Path
	if len(ws.Folders) > 0 && ws.Folders[0] != "" {
		dir = ws.Folders[0]
	}

	store.SaveMCPServer(mcpName, mcp.MCPServerUI{
		Command: "/home/john/.local/bin/sw",
		Args:    []string{"mcp"},
		Env: map[string]string{
			"WZ_PROJECT_PATH": dir,
		},
		Enabled: true,
		Icon:    "📋",
		Color:   "#8b5cf6",
	})
	fmt.Printf("[workspace] MCP server '%s' registered for workspace '%s'\n", mcpName, ws.Title)
}

// SyncSpecToWorkspace gera os arquivos .spec-wizard/ e .AGENTS.md no diretório
// do workspace com base nos dados do Spec Wizard vinculado.
// SyncSpecToWorkspace gera os arquivos .spec-wizard/ e .AGENTS.md via specwizardmgr.
func SyncSpecToWorkspace(store *db.Store, ws workspace.WorkspaceConfig) {
	specwizardmgr.SyncSpecToWorkspace(store, ws)
}

// SyncSpecToWorkspaceBySession é a versão que recebe sessionID para chamada do frontend.
func (a *App) SyncSpecToWorkspaceBySession(sessionID string) error {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return fmt.Errorf("sessão %s não encontrada", sessionID)
	}
	ws, err := a.eng.DB.GetWorkspace(sess.WorkspaceID)
	if err != nil {
		return fmt.Errorf("workspace %s não encontrado: %w", sess.WorkspaceID, err)
	}
	if ws.SpecWizardID == "" {
		return fmt.Errorf("workspace %s não possui Spec Wizard", ws.Title)
	}
	SyncSpecToWorkspace(a.eng.DB, ws)
	return nil
}

// ── Workspace Worker Management ──

// workspaceStorageID resolves the storage-level workspace ID from a path.
func (a *App) workspaceStorageID(path string) (int64, error) {
	ctx := context.Background()
	sw, err := a.eng.DB.WorkspaceStore().GetWorkspaceByPath(ctx, path)
	if err != nil {
		return 0, fmt.Errorf("workspace %s not found", path)
	}
	return sw.ID, nil
}

// AddWorkerToWorkspace vincula um worker a um workspace.
func (a *App) AddWorkerToWorkspace(workspacePath, workerName string) error {
	wsID, err := a.workspaceStorageID(workspacePath)
	if err != nil {
		return err
	}
	ctx := context.Background()
	w, err := a.eng.DB.WorkerStore().GetWorkerByName(ctx, workerName)
	if err != nil {
		return fmt.Errorf("worker %s not found", workerName)
	}
	return a.eng.DB.WorkspaceWorkers().AddWorker(ctx, &storage.WorkspaceWorker{
		WorkspaceID: wsID,
		WorkerID:    w.ID,
		Enabled:     true,
	})
}

// RemoveWorkerFromWorkspace remove um worker de um workspace.
func (a *App) RemoveWorkerFromWorkspace(workspacePath, workerName string) error {
	wsID, err := a.workspaceStorageID(workspacePath)
	if err != nil {
		return err
	}
	ctx := context.Background()
	w, err := a.eng.DB.WorkerStore().GetWorkerByName(ctx, workerName)
	if err != nil {
		return fmt.Errorf("worker %s not found", workerName)
	}
	return a.eng.DB.WorkspaceWorkers().RemoveWorker(ctx, wsID, w.ID)
}

// ListWorkspaceWorkers retorna os workers vinculados a um workspace.
func (a *App) ListWorkspaceWorkers(workspacePath string) ([]worker.WorkerConfig, error) {
	wsID, err := a.workspaceStorageID(workspacePath)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	links, err := a.eng.DB.WorkspaceWorkers().ListWorkers(ctx, wsID)
	if err != nil {
		return nil, err
	}
	all := a.eng.DB.ListWorkers()
	linked := make([]worker.WorkerConfig, 0, len(links))
	for _, link := range links {
		for _, w := range all {
			// Resolve worker ID via storage to match names
			sw, err := a.eng.DB.WorkerStore().GetWorkerByName(ctx, w.Name)
			if err == nil && sw.ID == link.WorkerID {
				linked = append(linked, w)
			}
		}
	}
	return linked, nil
}

// CountWorkerSessions retorna quantas sessões um worker tem em um workspace.
func (a *App) CountWorkerSessions(workspacePath, workerName string) int {
	sessions := a.eng.DB.ListSessions(workspacePath)
	count := 0
	for _, s := range sessions {
		if s.WorkerName == workerName {
			count++
		}
	}
	return count
}

// NextChatName retorna o próximo nome de chat disponível para um worker no workspace.
func (a *App) NextChatName(workspacePath, workerName string) string {
	sessions := a.eng.DB.ListSessions(workspacePath)
	used := make(map[string]bool)
	for _, s := range sessions {
		if s.WorkerName == workerName {
			used[s.Title] = true
		}
	}
	for i := 1; i <= 999; i++ {
		name := fmt.Sprintf("chat%d", i)
		if !used[name] {
			return name
		}
	}
	return fmt.Sprintf("chat%d", len(used)+1)
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

// GetWorkspaceDir retorna o diretório de trabalho do workspace pelo path.
func (a *App) GetWorkspaceDir(workspacePath string) string {
	return a.eng.ResolveWorkspaceDir(workspacePath)
}

// GetSessionDir retorna o diretório de trabalho da sessão.
func (a *App) GetSessionDir(sessionID string) string {
	return a.eng.ResolveSessionDir(sessionID)
}

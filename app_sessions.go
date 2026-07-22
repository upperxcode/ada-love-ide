package main

import (
	"errors"
	"fmt"

	"ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/config/workspace"
	"ada-love-ide/internal/engine"
	core "ada-love-core"
)

// Sessions / Chat ─────────────────────────────────────────────────

// CreateSession cria um novo chat no workspace/worker indicados.
func (a *App) CreateSession(workspaceID, workerName string) core.Session {
	return a.eng.Saver.Create(workspaceID, workerName)
}

// CreateSessionWithConfig cria um novo chat com nome único e copia a config (model, provider, mode, thinking) de uma sessão existente.
func (a *App) CreateSessionWithConfig(workspaceID, workerName, sourceSessionID string) (core.Session, error) {
	sess := a.eng.Saver.Create(workspaceID, workerName)

	// Generate unique chat name: "New Chat 1", "New Chat 2", etc.
	sessions := a.eng.DB.ListSessions(workspaceID)
	used := make(map[string]bool)
	for _, s := range sessions {
		if s.WorkerName == workerName {
			used[s.Title] = true
		}
	}
	for i := 1; i <= 999; i++ {
		name := fmt.Sprintf("New Chat %d", i)
		if !used[name] {
			sess.Title = name
			break
		}
	}

	if sourceSessionID != "" {
		src, ok := a.eng.DB.GetSession(sourceSessionID)
		if ok {
			sess.Model = src.Model
			sess.Provider = src.Provider
			sess.Mode = src.Mode
			sess.Thinking = src.Thinking
		}
	}

	a.eng.DB.PutSession(&sess)
	return sess, nil
}

// CreateSummarizedSession cria um chat filho de outro.
func (a *App) CreateSummarizedSession(workspaceID, workerName, sourceSessionID string) (core.Session, error) {
	return a.eng.Saver.CreateSummarized(workspaceID, workerName, sourceSessionID)
}

// GetSessions lista as sessões de um workspace.
func (a *App) GetSessions(workspaceID string) []core.Session {
	return a.eng.Fetcher.List(workspaceID)
}

// GetSessionByID retorna uma sessão pelo ID.
func (a *App) GetSessionByID(id string) (core.Session, error) {
	sess, ok := a.eng.DB.GetSession(id)
	if !ok {
		return core.Session{}, fmt.Errorf("sessão %s não encontrada", id)
	}
	return *sess, nil
}

// GetSessionMessages retorna as mensagens de uma sessão.
func (a *App) GetSessionMessages(sessionID string) []core.RawMessage {
	return a.eng.DB.GetMessages(sessionID)
}

// DeleteSession remove a sessão.
func (a *App) DeleteSession(id string) { a.eng.Saver.Delete(id) }

// RenameSession troca o título.
func (a *App) RenameSession(id, newTitle string) (core.Session, error) {
	return a.eng.Saver.Rename(id, newTitle)
}

// TogglePin inverte o estado fixado.
func (a *App) TogglePin(id string) error { return a.eng.Saver.TogglePin(id) }

// SetSessionConfig sobrescreve model/provider/mode/thinking.
func (a *App) SetSessionConfig(id, model, provider, mode, thinking string) error {
	return a.eng.Saver.SetConfig(id, model, provider, mode, thinking)
}

// ErrSessionNotFound re-export para que o frontend possa detectar.
var ErrSessionNotFound = errors.New("sessão não encontrada")

// GetSessionWorkspaceSpec retorna o Spec Wizard vinculado ao workspace da sessão.
func (a *App) GetSessionWorkspaceSpec(sessionID string) (*specwizard.SpecWizardConfig, error) {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return nil, fmt.Errorf("sessão %s não encontrada", sessionID)
	}
	ws, err := a.eng.DB.GetWorkspace(sess.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace %s não encontrado: %w", sess.WorkspaceID, err)
	}
	if ws.SpecWizardID == "" {
		return nil, fmt.Errorf("workspace %s não possui Spec Wizard configurado", ws.Title)
	}
	wiz, ok := a.eng.DB.GetWizard(ws.SpecWizardID)
	if !ok {
		return nil, fmt.Errorf("Spec Wizard %s não encontrado", ws.SpecWizardID)
	}
	return &wiz, nil
}

// GetWorkspaceSpec retorna o Spec Wizard configurado em um workspace pelo path.
func (a *App) GetWorkspaceSpec(workspacePath string) (*specwizard.SpecWizardConfig, error) {
	ws, err := a.eng.DB.GetWorkspace(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("workspace %s não encontrado: %w", workspacePath, err)
	}
	if ws.SpecWizardID == "" {
		return nil, fmt.Errorf("workspace %s não possui Spec Wizard configurado", ws.Title)
	}
	wiz, ok := a.eng.DB.GetWizard(ws.SpecWizardID)
	if !ok {
		return nil, fmt.Errorf("Spec Wizard %s não encontrado", ws.SpecWizardID)
	}
	return &wiz, nil
}

var _ = workspace.WorkspaceConfig{}

// ContextInfo retorna o uso de contexto de uma sessão.
func (a *App) GetSessionContextInfo(sessionID string) engine.ContextInfo {
	return a.eng.GetSessionContextInfo(sessionID)
}

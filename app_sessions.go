package main

import (
	"errors"

	"ada-love-ide/internal/core"
)

// Sessions / Chat ─────────────────────────────────────────────────

// CreateSession cria um novo chat no workspace/worker indicados.
func (a *App) CreateSession(workspaceID, workerName string) core.Session {
	return a.eng.Saver.Create(workspaceID, workerName)
}

// CreateSummarizedSession cria um chat filho de outro.
func (a *App) CreateSummarizedSession(workspaceID, workerName, sourceSessionID string) (core.Session, error) {
	return a.eng.Saver.CreateSummarized(workspaceID, workerName, sourceSessionID)
}

// GetSessions lista as sessões de um workspace.
func (a *App) GetSessions(workspaceID string) []core.Session {
	return a.eng.Fetcher.List(workspaceID)
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

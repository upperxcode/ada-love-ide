package main

import (
	"context"

	"ada-love-ide/internal/engine"
)

// App é a struct exposta ao frontend via Wails Bind. Todos os
// métodos públicos são automaticamente acessíveis via `window.go.main.App`.
type App struct {
	eng *engine.Engine
	ctx context.Context
}

// NewApp cria a App recebendo o engine injetado.
func NewApp(eng *engine.Engine) *App { return &App{eng: eng} }

// startup é chamado pelo Wails quando o runtime está pronto.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.eng.SetContext(ctx)
	a.eng.Chat.SetEmitter(engine.NewEmitter(ctx))
}

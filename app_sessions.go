package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ada-love-ide/internal/chat"
	"ada-love-ide/internal/cliworker"
	"ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/config/worker"
	"ada-love-ide/internal/config/workspace"
	"ada-love-ide/internal/contextprovider"
	"ada-love-ide/internal/engine"
	"ada-love-ide/internal/urlworker"
	core "ada-love-core"
)

// Sessions / Chat ─────────────────────────────────────────────────

// CreateSession cria um novo chat no workspace/worker indicados.
func (a *App) CreateSession(workspaceID, workerName string) core.Session {
	go a.ensureWorkspaceSummary(workspaceID)
	return a.eng.Saver.Create(workspaceID, workerName)
}

// CreateSessionWithConfig cria um novo chat com nome único e copia a config (model, provider, mode, thinking) de uma sessão existente.
func (a *App) CreateSessionWithConfig(workspaceID, workerName, sourceSessionID string) (core.Session, error) {
	sess := a.eng.Saver.Create(workspaceID, workerName)

	// Auto-generate workspace summary if needed (background, non-blocking)
	go a.ensureWorkspaceSummary(workspaceID)

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
// Se o mode mudar, notifica o PermissionStore para limpar grants em downgrade.
func (a *App) SetSessionConfig(id, model, provider, mode, thinking string) error {
	// Normaliza o mode e detecta mudança
	oldModeRaw := ""
	if sess, ok := a.eng.DB.GetSession(id); ok {
		oldModeRaw = sess.Mode
	}
	modeNorm := chat.NormalizeMode(mode)

	if mode != oldModeRaw && modeNorm != "" && a.eng.Chat != nil {
		a.eng.Chat.HandleModeChange(id, chat.ChatMode(modeNorm))
	}

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

// GetSessionWorker retorna a configuração do worker vinculado a uma sessão.
func (a *App) GetSessionWorker(sessionID string) (worker.WorkerConfig, error) {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return worker.WorkerConfig{}, fmt.Errorf("sessão %s não encontrada", sessionID)
	}
	if sess.WorkerName == "" {
		return worker.WorkerConfig{}, errors.New("sessão não possui worker vinculado")
	}
	return a.eng.DB.GetWorker(sess.WorkerName)
}

// SendCLIMessage executa um worker CLI com a mensagem do usuário e retorna a resposta.
func (a *App) SendCLIMessage(sessionID, message, model string) (string, error) {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("SendCLIMessage: sessão %s não encontrada", sessionID)
	}

	cfg, err := a.eng.DB.GetWorker(sess.WorkerName)
	if err != nil {
		return "", fmt.Errorf("SendCLIMessage: worker %q não encontrado: %w", sess.WorkerName, err)
	}

	rt := cliworker.New(cfg).WithWorkspace(sess.WorkspaceID)
	cmd := rt.BuildCommand(message, model)
	fmt.Printf("[SendCLIMessage] exec: %s %s (dir=%s)\n", cmd.Path, strings.Join(cmd.Args, " "), cmd.Dir)

	emitter := engine.NewEmitter(a.ctx)
	resp, err := cliworker.ExecuteStream(cmd, emitter.Emit)
	if err != nil {
		return resp, fmt.Errorf("CLI worker error: %w", err)
	}

	// Persist the exchange (user + assistant)
	a.eng.DB.AppendMessage(sessionID, core.RawMessage{
		Role: "user", Content: message, Time: time.Now(),
	})
	if resp != "" {
		a.eng.DB.AppendMessage(sessionID, core.RawMessage{
			Role: "assistant", Content: resp, Time: time.Now(),
		})
	}

	return resp, nil
}

// GetCLIModels retorna a lista de modelos disponíveis para um worker CLI.
func (a *App) GetCLIModels(workerName string) ([]cliworker.ModelInfo, error) {
	cfg, err := a.eng.DB.GetWorker(workerName)
	if err != nil {
		return nil, fmt.Errorf("worker %q não encontrado: %w", workerName, err)
	}
	if cfg.ConnectionType != "cli" {
		return nil, fmt.Errorf("worker %q não é do tipo CLI", workerName)
	}

	rt := cliworker.New(cfg)
	models, err := cliworker.ListModels(rt)
	if err != nil {
		return nil, fmt.Errorf("GetCLIModels: %w", err)
	}
	return models, nil
}

// SendURLMessage envia uma mensagem via worker URL/API e retorna a resposta.
func (a *App) SendURLMessage(sessionID, message, model string) (string, error) {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("SendURLMessage: sessão %s não encontrada", sessionID)
	}

	cfg, err := a.eng.DB.GetWorker(sess.WorkerName)
	if err != nil {
		return "", fmt.Errorf("SendURLMessage: worker %q não encontrado: %w", sess.WorkerName, err)
	}

	rt := urlworker.New(cfg)
	req := rt.BuildChatRequest(message, model)
	fmt.Printf("[SendURLMessage] %s %s\n", req.Method, req.URL)

	emitter := engine.NewEmitter(a.ctx)
	resp, err := urlworker.ExecuteChat(req, emitter.Emit)
	if err != nil {
		return resp, fmt.Errorf("URL worker error: %w", err)
	}

	a.eng.DB.AppendMessage(sessionID, core.RawMessage{
		Role: "user", Content: message, Time: time.Now(),
	})
	if resp != "" {
		a.eng.DB.AppendMessage(sessionID, core.RawMessage{
			Role: "assistant", Content: resp, Time: time.Now(),
		})
	}

	return resp, nil
}

// GetURLModels retorna a lista de modelos disponíveis para um worker URL/API.
func (a *App) GetURLModels(workerName string) ([]urlworker.ModelInfo, error) {
	cfg, err := a.eng.DB.GetWorker(workerName)
	if err != nil {
		return nil, fmt.Errorf("worker %q não encontrado: %w", workerName, err)
	}
	if cfg.ConnectionType != "url" && cfg.ConnectionType != "opencode_serve" {
		return nil, fmt.Errorf("worker %q não é do tipo URL", workerName)
	}

	rt := urlworker.New(cfg)
	req := rt.BuildModelsRequest()
	fmt.Printf("[GetURLModels] fetching models from %s %s\n", req.Method, req.URL)
	models, err := urlworker.FetchModels(req)
	if err != nil {
		fmt.Printf("[GetURLModels] error: %v\n", err)
		return nil, fmt.Errorf("GetURLModels: %w", err)
	}
	fmt.Printf("[GetURLModels] received %d models\n", len(models))
	for _, m := range models {
		fmt.Printf("  - %s/%s\n", m.ProviderName, m.Name)
	}
	return models, nil
}

// StartURLWorkerServer inicia o servidor de um worker URL em background.
func (a *App) StartURLWorkerServer(workerName string) error {
	cfg, err := a.eng.DB.GetWorker(workerName)
	if err != nil {
		return fmt.Errorf("worker %q não encontrado: %w", workerName, err)
	}
	if cfg.ConnectionType != "url" && cfg.ConnectionType != "opencode_serve" {
		return fmt.Errorf("worker %q não é do tipo URL ou OpenCode Server", workerName)
	}

	rt := urlworker.New(cfg)
	return urlworker.DefaultServerManager.Start(*rt)
}

// StopURLWorkerServer para o servidor de um worker URL.
func (a *App) StopURLWorkerServer(workerName string) error {
	return urlworker.DefaultServerManager.Stop(workerName)
}

// GetURLWorkerStatus retorna o status atual do servidor de um worker URL.
func (a *App) GetURLWorkerStatus(workerName string) map[string]any {
	running, port, baseURL, uptime := urlworker.DefaultServerManager.Status(workerName)
	return map[string]any{
		"running":  running,
		"port":     port,
		"base_url": baseURL,
		"uptime":   uptime,
	}
}

// SendOpenCodeMessage envia uma mensagem via OpenCode Server (com gerenciamento de sessão).
func (a *App) SendOpenCodeMessage(sessionID, message, model string) (string, error) {
	sess, ok := a.eng.DB.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("SendOpenCodeMessage: sessão %s não encontrada", sessionID)
	}

	cfg, err := a.eng.DB.GetWorker(sess.WorkerName)
	if err != nil {
		return "", fmt.Errorf("SendOpenCodeMessage: worker %q não encontrado: %w", sess.WorkerName, err)
	}

	rt := urlworker.New(cfg)
	emitter := engine.NewEmitter(a.ctx)

	// Get or create opencode server session
	ocID, exists := urlworker.DefaultOpenCodeManager.Get(sessionID)
	contextSystem := ""
	var contextParts []string

	if !exists {
		fmt.Printf("[SendOpenCodeMessage] creating opencode session for Ada session %s\n", sessionID)
		ocID, err = rt.OpenCodeCreateSession()
		if err != nil {
			return "", fmt.Errorf("failed to create opencode session: %w", err)
		}
		urlworker.DefaultOpenCodeManager.Set(sessionID, ocID)
		fmt.Printf("[SendOpenCodeMessage] opencode session created: %s\n", ocID)

		// Load context from contextprovider for new sessions
		ws, wsErr := a.eng.DB.GetWorkspace(sess.WorkspaceID)
		if wsErr != nil {
			fmt.Printf("[SendOpenCodeMessage] GetWorkspace error: %v\n", wsErr)
		} else {
			fmt.Printf("[SendOpenCodeMessage] workspace loaded: %q summary_len=%d\n", ws.Title, len(ws.Summary))
			ctxParams := contextprovider.Params{
				WorkerType:   contextprovider.TypeOpenCodeServe,
				Worker:       cfg,
				Workspace:    ws,
				WorkspaceDir: sess.WorkspaceID,
				WorkspaceID:  a.eng.DB.WorkspaceIDByPath(sess.WorkspaceID),
				SessionID:    sessionID,
				CurrentMsg:   message,
				KnowledgeIdx: a.eng.KnowledgeIndex,
			}
			ctxResult, ctxErr := contextprovider.GetContext(context.Background(), ctxParams)
			if ctxErr != nil {
				fmt.Printf("[SendOpenCodeMessage] GetContext error: %v\n", ctxErr)
			} else if ctxResult != nil {
				fmt.Printf("[SendOpenCodeMessage] context: system=%d parts=%d history=%d\n",
					len(ctxResult.System), len(ctxResult.Parts), len(ctxResult.History))
				contextSystem = ctxResult.System
				contextParts = ctxResult.Parts
				if ctxResult.History != "" {
					contextParts = append(contextParts, ctxResult.History)
				}
			}
		}
	}

	fmt.Printf("[SendOpenCodeMessage] sending to opencode: system=%d context_parts=%d\n", len(contextSystem), len(contextParts))
	resp, thinking, err := rt.OpenCodeSendMessage(ocID, message, model, emitter.Emit, contextSystem, contextParts...)
	if err != nil {
		return resp, fmt.Errorf("OpenCode message error: %w", err)
	}

	// Persist the exchange (user → assistant → thinking)
	a.eng.DB.AppendMessage(sessionID, core.RawMessage{
		Role: "user", Content: message, Time: time.Now(),
	})
	if resp != "" {
		a.eng.DB.AppendMessage(sessionID, core.RawMessage{
			Role: "assistant", Content: resp, Time: time.Now(),
		})
	}
	if thinking != "" {
		a.eng.DB.AppendMessage(sessionID, core.RawMessage{
			Role: "thinking", Content: thinking, Time: time.Now(),
		})
	}

	return resp, nil
}

// GetOpenCodeModels retorna a lista de modelos de um servidor OpenCode.
func (a *App) GetOpenCodeModels(workerName string) ([]urlworker.ModelInfo, error) {
	cfg, err := a.eng.DB.GetWorker(workerName)
	if err != nil {
		return nil, fmt.Errorf("worker %q não encontrado: %w", workerName, err)
	}
	if cfg.ConnectionType != "opencode_serve" {
		return nil, fmt.Errorf("worker %q não é OpenCode Server", workerName)
	}

	rt := urlworker.New(cfg)
	models, err := rt.OpenCodeListModels()
	if err != nil {
		return nil, err
	}
	fmt.Printf("[GetOpenCodeModels] %d models for worker %s\n", len(models), workerName)
	for _, m := range models {
		fmt.Printf("  [%s] provider=%s model=%s\n", m.ID, m.ProviderName, m.Name)
	}
	return models, nil
}

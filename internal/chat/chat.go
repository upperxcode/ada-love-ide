package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"ada-love-ide/internal/adapters"
	core "ada-love-core"

	llm "github.com/upperxcode/ada-llm-client"
	commands "github.com/upperxcode/ada-commands"
	stream "github.com/upperxcode/ada-stream"
)

// Emitter is the interface for sending events to the frontend.
type Emitter interface {
	Emit(event string, data ...any)
}

// FrontendEmitter is a placeholder for emitting events to the frontend via Wails runtime.
type WailsEmitter struct {
	// In a real app, this would hold a reference to Wails runtime.EventsEmit
}

// NewWailsEmitter creates a new instance of WailsEmitter.
func NewWailsEmitter() *WailsEmitter {
	return &WailsEmitter{}
}

// Emit implements the chat.Emitter interface.
func (we *WailsEmitter) Emit(event string, data ...interface{}) {
	// In a Wails app, this would call: runtime.EventsEmit(runtime.Context(), event, data...)
	fmt.Printf("[DEBUG WailsEmitter] event=%s, data=%v\n", event, data)
	// In a real Wails app, you'd use something like:
	// runtime.EventsEmit(runtime.Context(), event, data...)
	fmt.Printf("[FrontendEmitter] Emitting event: %s, data: %v\n", event, data)
}

// streamingEmitter implementa a interface Emitter para streaming SSE
type streamingEmitter struct {
	tokenChan chan<- string
	doneChan  chan bool
}

func (e *streamingEmitter) Emit(event string, data ...any) {
	switch event {
	case "token-received":
		if len(data) > 1 {
			if m, ok := data[1].(map[string]interface{}); ok {
				if token, ok := m["token"].(string); ok {
					select {
					case e.tokenChan <- token:
					default:
					}
				}
			}
		}
	case "stream-finished":
		e.doneChan <- true
	}
}

// Chat manages the chat state and interactions.
type Chat struct {
	orch            *core.Orchestrator
	emitter         Emitter // This will be our FrontendEmitter
	streamingClient adapters.LLMStreamingClient
	permStore       *PermissionStore
	systemPrompt    string

	mu       sync.Mutex
	cancelFn map[string]context.CancelFunc
	pending  map[string]*pendingStream // accumulated response per session
}

func (c *Chat) SetPermissionStore(ps *PermissionStore) {
	c.permStore = ps
}

// NormalizeMode converte string de modo para o formato padronizado (ex: "execute" → "EXECUTE").
// Retorna "ASK" para valores não reconhecidos.
func NormalizeMode(mode string) string {
	switch strings.ToLower(mode) {
	case "ask":
		return string(ModeAsk)
	case "edit":
		return string(ModeEdit)
	case "plan":
		return string(ModePlan)
	case "execute", "exec", "test":
		return string(ModeExec)
	case "full":
		return string(ModeFull)
	case "admin", "config":
		return string(ModeAdmin)
	default:
		return string(ModeAsk)
	}
}

func (c *Chat) RespondPermission(requestID, decision string) {
	if c.permStore == nil {
		return
	}
	c.permStore.SendDecision(requestID, decision)
}

// HandleModeChange processa uma mudança de modo, limpando grants se for downgrade.
// Retorna true se limpou grants.
func (c *Chat) HandleModeChange(sessionID string, newMode ChatMode) bool {
	if c.permStore == nil {
		return false
	}
	oldMode := c.permStore.GetCurrentMode(sessionID)
	if oldMode == newMode {
		return false
	}
	result := c.permStore.SetCurrentMode(sessionID, newMode, c.emitter)
	if result {
		// Log completo no terminal
		fmt.Printf("\n=== MODO ALTERADO: %s → %s ===\n", oldMode, newMode)
		fmt.Println("Downgrade detectado — todos os grants foram limpos!")
		c.permStore.DumpGrants(sessionID)
	} else {
		fmt.Printf("[Chat] Mode change: %s → %s (upgrade/same, grants mantidos)\n", oldMode, newMode)
		c.permStore.DumpGrants(sessionID)
	}
	return result
}

type ThinkingSection struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type pendingStream struct {
	userInput        string
	response         strings.Builder
	thinkingContent  strings.Builder
	thinkingSections []ThinkingSection
	lastThinkingType string
}

// New creates a new Chat instance.
func New(orch *core.Orchestrator, frontendEmitter Emitter) *Chat {
	return &Chat{
		orch:     orch,
		emitter:  frontendEmitter,
		cancelFn: map[string]context.CancelFunc{},
		pending:  map[string]*pendingStream{},
	}
}

func (c *Chat) SetEmitter(e Emitter) { c.emitter = e }

func (c *Chat) SetStreamingClient(sc adapters.LLMStreamingClient) {
	c.streamingClient = sc
}

var ErrNoEmitter = errors.New("emitter não configurado")
var ErrSessionNotFound = errors.New("sessão não encontrada")

func (c *Chat) Send(ctx context.Context, sessionID, text, model, thinking, mode string, contextSize ...int) (string, error) {
	fmt.Printf("[Chat.Send] ENTER session=%s text=%q model=%s thinking=%s mode=%s\n", sessionID, text[:min(50, len(text))], model, thinking, mode)
	if c.emitter == nil {
		return "", ErrNoEmitter
	}

	// If thinking level not provided, get it from session
	if thinking == "" && c.orch != nil && c.orch.Storage != nil {
		if sess, ok := c.orch.Storage.GetSession(sessionID); ok && sess.HasThinking() {
			thinking = sess.Thinking
			fmt.Printf("[Chat.Send] Got thinking level from session: %s\n", thinking)
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancelFn[sessionID] = cancel
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.cancelFn, sessionID)
		c.mu.Unlock()
	}()

	c.emitter.Emit("chat:status", map[string]any{
		"session_id": sessionID,
		"stage":      "thinking",
	})

	// Check if thinking mode is enabled and emit orchestrator decision event
	if thinking == "high" || thinking == "medium" || thinking == "low" {
		c.emitter.Emit("orchestrator:decision", map[string]any{
			"session_id": sessionID,
			"reasoning":  "[Analisando...]",
			"next_agent": "",
			"task":       "Processando mensagem com raciocínio profundo",
			"sub_tasks":  0,
		})
	}

	// Check for command prefix and execute via ada-commands router
	if strings.HasPrefix(strings.TrimSpace(text), "/") {
		cmdName, args := commands.ParseCommand(text)
		if cmdName != "" {
			response, err := c.orch.ExecuteCommand(ctx, sessionID, cmdName, args)
			if err != nil {
				return "", err
			}
			// Emit command result event for snackbar notification
			// Emit BEFORE turnEnd to ensure frontend processes it first
			c.emitter.Emit("chat:commandResult", map[string]any{
				"session_id": sessionID,
				"command":    "/" + cmdName,
				"output":     response,
				"done":       true,
			})
			// Emit a short-lived notification for health command
			if cmdName == "health" {
				c.emitter.Emit("command:notification", map[string]any{
					"title":    "Health Check Complete",
					"response": response,
					"duration": 10000, // 10 seconds
				})
			}
			// Emit chat:cleared event for clear command
			if cmdName == "clear" {
				c.emitter.Emit("chat:cleared", map[string]any{
					"session_id": sessionID,
					"command":    "/" + cmdName,
					"response":   response,
				})
			}
			// Emit health:status event for health command
			if cmdName == "health" {
				c.emitter.Emit("health:status", map[string]any{
					"session_id": sessionID,
					"response":   response,
				})
			}
			// Emit turnEnd to reset loading state
			c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
			// Return empty string so command result doesn't appear in chat
			// The result is shown in CommandResultPanel via chat:commandResult event
			return "", nil
		}
	}

	// Check static response from ada-llm-client (greetings, etc.)
	if response, ok := llm.CheckStaticResponse(text); ok {
		fmt.Printf("[Chat.Send] Static response matched for %q\n", text)
		c.orch.SaveMessages(sessionID, text, response)
		c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
		return response, nil
	}

	// Build full context via Orchestrator (pass context size if known)
	prompt, err := c.orch.CompilePrompt(ctx, sessionID, text, contextSize...)
	if err != nil {
		return "", fmt.Errorf("failed to compile prompt: %w", err)
	}

	if c.streamingClient != nil {
		c.mu.Lock()
		c.pending[sessionID] = &pendingStream{userInput: text}
		c.mu.Unlock()

		emitter := func(eventName string, optionalData ...interface{}) {
			c.handleStreamEvent(sessionID, eventName, optionalData...)
		}

		c.streamingClient.SetEmitter(stream.EventEmitter(emitter))

		// Normaliza o mode para uppercase (ex: "execute" → "EXECUTE")
		mode = NormalizeMode(mode)
		cfg := GetModeConfig(ChatMode(mode))
		c.streamingClient.SetMode(mode)
		if c.systemPrompt != "" {
			c.streamingClient.SetSystemPrompt(c.systemPrompt)
		} else {
			c.streamingClient.SetSystemPrompt(cfg.SystemPrompt)
		}

		if c.permStore != nil {
			permGuard := c.permStore.MakeGuard(ctx, sessionID, ChatMode(mode), c.emitter)
			c.streamingClient.SetPermissionGuard(permGuard)
		}

		err := c.streamingClient.GenerateStream(ctx, sessionID, prompt, model)
		if err != nil {
			fmt.Printf("[Chat.Send] GenerateStream ERROR: %v\n", err)
			c.emitter.Emit("chat:error", map[string]any{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			return "", err
		}

		return "", nil
	}

	// Non-streaming fallback
	tokens, err := c.orch.ProcessMessageStream(ctx, sessionID, text, model)
	if err != nil {
		fmt.Printf("[Chat.Send] ProcessMessageStream ERROR: %v\n", err)
		return "", err
	}

	var fullStr strings.Builder
	for token := range tokens {
		if token.Token != "" {
			fullStr.WriteString(token.Token)
			c.emitter.Emit("chat:delta", map[string]any{
				"session_id": sessionID,
				"content":    token.Token,
			})
		}
		if token.Done {
			break
		}
	}

	reply := fullStr.String()
	c.orch.SaveMessages(sessionID, text, reply)
	c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
	return reply, nil
}

func (c *Chat) handleStreamEvent(sessionID, eventName string, data ...interface{}) {
	switch eventName {
	case "token-received":
		if len(data) > 0 {
			if m, ok := data[0].(map[string]interface{}); ok {
				if token, ok := m["token"].(string); ok {
					// token is the FULL accumulated text — store it directly
					c.mu.Lock()
					if p, ok := c.pending[sessionID]; ok {
						p.response.Reset()
						p.response.WriteString(token)
					}
					c.mu.Unlock()
					c.emitter.Emit("chat:delta", map[string]any{
						"session_id": sessionID,
						"content":    token,
					})
				}
			}
		}
	case "stream-finished":
		c.mu.Lock()
		p, ok := c.pending[sessionID]
		if ok {
			delete(c.pending, sessionID)
		}
		c.mu.Unlock()
		if ok {
			c.orch.SaveMessages(sessionID, p.userInput, p.response.String())
			if rawThinking := strings.TrimSpace(p.thinkingContent.String()); rawThinking != "" {
				var content string
				if len(p.thinkingSections) > 0 {
					payload := map[string]any{
						"text":     rawThinking,
						"sections": p.thinkingSections,
					}
					if b, err := json.Marshal(payload); err == nil {
						content = string(b)
					}
				}
				if content == "" {
					content = rawThinking
				}
				msg := core.Message{
					ID:        fmt.Sprintf("%d-think", time.Now().UnixNano()),
					SessionID: sessionID,
					Role:      "thinking",
					Content:   content,
					CreatedAt: time.Now().Format(time.RFC3339),
				}
				if c.orch.Storage != nil {
					_ = c.orch.Storage.SaveMessage(msg)
				}
			}
		}
		if c.permStore != nil {
			c.permStore.ClearSessionGrants(sessionID)
		}
		c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
	case "reasoning-received":
		if len(data) > 0 {
			if m, ok := data[0].(map[string]interface{}); ok {
				if reasoning, ok := m["reasoning"].(string); ok && reasoning != "" {
					evt := map[string]any{
						"session_id": sessionID,
						"content":    reasoning,
					}
					sectionType := "text"
					if t, ok := m["type"].(string); ok {
						evt["type"] = t
						sectionType = t
					}
					c.emitter.Emit("chat:thinking", evt)
					c.mu.Lock()
					if p, ok := c.pending[sessionID]; ok {
						p.thinkingContent.WriteString(reasoning)
						if p.lastThinkingType != sectionType || len(p.thinkingSections) == 0 {
							p.thinkingSections = append(p.thinkingSections, ThinkingSection{Type: sectionType, Content: reasoning})
							p.lastThinkingType = sectionType
						} else {
							p.thinkingSections[len(p.thinkingSections)-1].Content += reasoning
						}
					}
					c.mu.Unlock()
				}
			}
		}
	case "stream-interrupted":
		c.mu.Lock()
		delete(c.pending, sessionID)
		c.mu.Unlock()
		if c.permStore != nil {
			c.permStore.ClearSessionGrants(sessionID)
		}
		c.emitter.Emit("chat:status", map[string]any{
			"session_id": sessionID,
			"stage":      "interrupted",
		})
	}
}

func (c *Chat) Stop(sessionID string) {
	c.mu.Lock()
	fn, ok := c.cancelFn[sessionID]
	c.mu.Unlock()
	if ok {
		fn()
	}
}

// QuickGenerate faz uma inferência direta sem streaming, sessão ou eventos.
func (c *Chat) QuickGenerate(ctx context.Context, prompt, model string) (string, error) {
	if c.streamingClient == nil {
		return "", fmt.Errorf("streaming client not available")
	}
	return c.streamingClient.GenerateSimple(ctx, prompt, model, false)
}

// SendStreamingToChannel inicia o streaming e envia tokens para o canal
func (c *Chat) SendStreamingToChannel(ctx context.Context, sessionID string, tokenChan chan<- string) error {
	if c.emitter == nil {
		return ErrNoEmitter
	}

	// Criar emitter que encaminha tokens para o canal
	c.emitter = &streamingEmitter{
		tokenChan: tokenChan,
		doneChan:  make(chan bool, 1),
	}

	// Se tiver streamingClient, usar GenerateStream
	if c.streamingClient != nil {
		c.streamingClient.SetEmitter(stream.EventEmitter(func(eventName string, data ...interface{}) {
			switch eventName {
			case "token-received":
				if len(data) > 1 {
					if m, ok := data[1].(map[string]interface{}); ok {
						if token, ok := m["token"].(string); ok {
							select {
							case tokenChan <- token:
							case <-ctx.Done():
							}
						}
					}
				}
			case "stream-finished":
				select {
				case tokenChan <- "":
				case <-ctx.Done():
				}
			}
		}))

		return c.streamingClient.GenerateStream(ctx, sessionID, "", "")
	}

	return nil
}

package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"ada-love-ide/internal/core"

	commands "github.com/upperxcode/ada-commands"
	stream "github.com/upperxcode/ada-stream"
)

// Emitter is the interface for sending events to the frontend.
type Emitter interface {
	Emit(event string, data ...any)
}

// FrontendEmitter is a placeholder for emitting events to the frontend via Wails runtime.
type FrontendEmitter struct {
	// In a real app, this would hold a reference to Wails runtime.EventsEmit
}

// NewFrontendEmitter creates a new instance of FrontendEmitter.
func NewFrontendEmitter() *FrontendEmitter {
	return &FrontendEmitter{}
}

// Emit implements the chat.Emitter interface.
func (fe *FrontendEmitter) Emit(event string, data ...interface{}) {
	// Simulate sending to frontend via Wails runtime.
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
	streamingClient core.LLMStreamingClient

	mu       sync.Mutex
	cancelFn map[string]context.CancelFunc
	pending  map[string]*pendingStream // accumulated response per session
}

type pendingStream struct {
	userInput string
	response  strings.Builder
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

func (c *Chat) SetStreamingClient(sc core.LLMStreamingClient) {
	c.streamingClient = sc
}

var ErrNoEmitter = errors.New("emitter não configurado")
var ErrSessionNotFound = errors.New("sessão não encontrada")

func (c *Chat) Send(ctx context.Context, sessionID, text, model, thinking, mode string) (string, error) {
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

	if c.streamingClient != nil {
		// Register pending stream for this session
		c.mu.Lock()
		c.pending[sessionID] = &pendingStream{userInput: text}
		c.mu.Unlock()

		emitter := func(eventName string, optionalData ...interface{}) {
			c.handleStreamEvent(sessionID, eventName, optionalData...)
		}

		c.streamingClient.SetEmitter(stream.EventEmitter(emitter))

		err := c.streamingClient.GenerateStream(ctx, sessionID, text, model)
		if err != nil {
			fmt.Printf("[Chat.Send] GenerateStream ERROR: %v\n", err)
			c.emitter.Emit("chat:error", map[string]any{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			return "", err
		}

		// The actual streaming and event emission are handled by the adapter
		// via handleStreamEvent and c.emitter. We should not call synchronous
		// ProcessMessage here. The Send method should simply initiate the stream.
		// The frontend will receive updates via delta events.
		// The chat:turnEnd event is emitted by handleStreamEvent when the stream finishes.
		return "", nil // Return empty, as response is streamed via deltas.
	}

	// Fallback for non-streaming clients or when streamingClient is nil
	tokens, err := c.orch.ProcessMessageStream(ctx, sessionID, text, model)
	if err != nil {
		fmt.Printf("[Chat.Send] ProcessMessageStream ERROR: %v\n", err)
		return "", err
	}

	var full strings.Builder
	tokenCount := 0
	for token := range tokens {
		tokenCount++
		if token.Token != "" {
			full.WriteString(token.Token)
			c.emitter.Emit("chat:delta", map[string]any{
				"session_id": sessionID,
				"content":    token.Token,
			})
		}
		if token.Done {
			break
		}
	}
	fmt.Printf("[Chat.Send] Stream complete, tokens=%d, fullLen=%d\n", tokenCount, len(full.String()))

	reply := full.String()
	c.orch.SaveMessages(sessionID, text, reply)
	c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
	fmt.Printf("[Chat.Send] EXIT session=%s replyLen=%d\n", sessionID, len(reply))
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
		}
		c.emitter.Emit("chat:turnEnd", map[string]any{"session_id": sessionID})
	case "stream-interrupted":
		c.mu.Lock()
		delete(c.pending, sessionID)
		c.mu.Unlock()
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

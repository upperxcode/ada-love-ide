package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	core "ada-love-core"
	llm "github.com/upperxcode/ada-llm-client"
	stream "github.com/upperxcode/ada-stream"
)

type ProviderConfig struct {
	Type    string
	BaseURL string
	APIKey  string
}

type ToolHandler func(ctx context.Context, name string, argsJSON string) (string, error)

type PermissionGuard func(toolName, argsJSON string) (allowed bool, reason string, requestID string)

type MultiLLMAdapter struct {
	providers       map[string]llm.StreamingLLMClient
	defaultModel    string
	streamMgr       *stream.StreamManager
	emitter         stream.EventEmitter
	tools           []llm.ToolDefinition
	toolHandler     ToolHandler
	onModelError    func(providerName, modelID string, err error)
	mode            string
	systemPrompt    string
	permissionGuard PermissionGuard
}

func (a *MultiLLMAdapter) SetMode(mode string) {
	a.mode = mode
}

func (a *MultiLLMAdapter) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

func (a *MultiLLMAdapter) SetPermissionGuard(guard PermissionGuard) {
	a.permissionGuard = guard
}

const (
	modeAsk  = "ASK"
	modePlan = "PLAN"
	modeEdit = "EDIT"
	modeFull = "FULL"
)

func (a *MultiLLMAdapter) allowedTools() []llm.ToolDefinition {
	switch a.mode {
	case modeAsk, modePlan:
		return filterTools(a.tools, []string{"read", "search"})
	case modeEdit:
		return filterTools(a.tools, []string{"read", "search", "write"})
	case modeFull:
		return a.tools
	}
	return a.tools
}

func filterTools(tools []llm.ToolDefinition, allowed []string) []llm.ToolDefinition {
	if len(allowed) == 0 {
		return nil
	}
	var out []llm.ToolDefinition
	allowedSet := make(map[string]bool)
	for _, a := range allowed {
		allowedSet[a] = true
	}
	for _, t := range tools {
		if allowedSet[t.Function.Name] {
			out = append(out, t)
		}
	}
	return out
}

const maxToolIterations = 10

func (a *MultiLLMAdapter) SetTools(tools []llm.ToolDefinition) {
	a.tools = tools
}

func (a *MultiLLMAdapter) SetToolHandler(h ToolHandler) {
	a.toolHandler = h
}

func (a *MultiLLMAdapter) SetOnModelError(fn func(providerName, modelID string, err error)) {
	a.onModelError = fn
}

func (a *MultiLLMAdapter) resolveClient(model string) (llm.StreamingLLMClient, string) {
	if model == "" || !strings.Contains(model, "/") {
		return nil, model
	}
	providerName := model[:strings.Index(model, "/")]
	client, ok := a.providers[providerName]
	if !ok {
		fmt.Printf("[MultiLLMAdapter.resolveClient] provider %q not found, returning model as-is: %q\n", providerName, model)
		return nil, model
	}
	// Strip the provider prefix (everything before and including the first "/")
	// so that only the actual model ID is sent to the provider's API.
	// Example: "Openrouter/poolside/laguna-xs-2.1:free" → "poolside/laguna-xs-2.1:free"
	modelID := model[strings.Index(model, "/")+1:]
	fmt.Printf("[MultiLLMAdapter.resolveClient] provider=%q model=%q → resolvedModel=%q\n", providerName, model, modelID)
	return client, modelID
}

func NewMultiLLMAdapter(providers map[string]ProviderConfig, defaultModel string) *MultiLLMAdapter {
	return NewMultiLLMAdapterWithEmitter(providers, defaultModel, nil)
}

func NewMultiLLMAdapterWithEmitter(providers map[string]ProviderConfig, defaultModel string, emitter stream.EventEmitter) *MultiLLMAdapter {
	clientMap := make(map[string]llm.StreamingLLMClient)
	for name, cfg := range providers {
		lc := llm.ConnectionConfig{
			Type:    llm.ConnectionType(cfg.Type),
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		}
		clientMap[name] = llm.NewStreamingClient(lc)
	}
	return &MultiLLMAdapter{
		providers:    clientMap,
		defaultModel: defaultModel,
		streamMgr:    stream.NewStreamManager(),
		emitter:      emitter,
	}
}

func (a *MultiLLMAdapter) SetEmitter(emitter stream.EventEmitter) {
	a.emitter = emitter
}

func (a *MultiLLMAdapter) Generate(ctx context.Context, prompt string, model string) (<-chan core.LLMToken, error) {
	ch := make(chan core.LLMToken, 100)

	go func() {
		defer close(ch)

		if response, ok := llm.CheckStaticResponse(prompt); ok {
			ch <- core.LLMToken{Token: response, Done: true}
			return
		}

		client, resolvedModel := a.resolveClient(model)
		if client == nil {
			errMsg := fmt.Sprintf("Provider para modelo '%s' não encontrado.", model)
			ch <- core.LLMToken{Token: "[Error: " + errMsg + "]", Done: true}
			return
		}

		messages := []llm.Message{llm.NewUserMessage(prompt)}

		for iter := 0; iter < maxToolIterations; iter++ {
			req := llm.InferenceRequest{
				Messages: messages,
				Config: llm.InferenceConfig{
					Model:       resolvedModel,
					Temperature: 0.7,
					MaxTokens:   4096,
				},
			}
			filtered := a.allowedTools()
			if len(filtered) > 0 {
				req.Tools = filtered
			}

			resp, toolCalls, err := client.Generate(ctx, req)
			if err != nil {
				ch <- core.LLMToken{Token: "[Error: " + err.Error() + "]", Done: true}
				return
			}

			if len(toolCalls) == 0 {
				if resp != "" {
					ch <- core.LLMToken{Token: resp}
				}
				ch <- core.LLMToken{Done: true}
				return
			}

			astMsg := llm.NewAssistantMessage(resp)
			astMsg = llm.MessageWithToolCalls(astMsg, toolCalls)
			messages = append(messages, astMsg)

			for _, tc := range toolCalls {
				result := ""
				if a.toolHandler != nil {
					result, err = a.toolHandler(ctx, tc.Function.Name, tc.Function.Arguments)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
					}
				} else {
					result = fmt.Sprintf("Tool %s not available", tc.Function.Name)
				}
				messages = append(messages, llm.NewToolMessage(tc.ID, result))
			}
		}

		ch <- core.LLMToken{Token: "[Error: max tool iterations reached]", Done: true}
	}()

	return ch, nil
}

func (a *MultiLLMAdapter) GenerateStream(ctx context.Context, sessionID string, prompt string, model string) error {
	fmt.Printf("[MultiLLMAdapter.GenerateStream] ENTER session=%s model=%s\n", sessionID, model)

	if response, ok := llm.CheckStaticResponse(prompt); ok {
		if a.emitter != nil {
			a.emitter("stream:chunk", map[string]interface{}{
				"sessionID": sessionID,
				"type":      string(stream.ChunkContent),
				"payload":   response,
			})
			a.emitter("stream-finished", map[string]interface{}{
				"sessionID": sessionID,
			})
		}
		return nil
	}

	client, resolvedModel := a.resolveClient(model)
	if client == nil {
		errMsg := fmt.Sprintf("Provider '%s' não encontrado.", model)
		if a.emitter != nil {
			a.emitter("stream:chunk", map[string]interface{}{
				"sessionID": sessionID,
				"type":      string(stream.ChunkError),
				"payload":   errMsg,
			})
			a.emitter("stream-finished", map[string]interface{}{
				"sessionID": sessionID,
			})
		}
		return fmt.Errorf("%s", errMsg)
	}

	ctx, cancel := a.streamMgr.Register(ctx, sessionID)
	defer func() {
		a.streamMgr.Unregister(sessionID)
		cancel()
	}()

messages := []llm.Message{}
	if a.systemPrompt != "" {
		messages = append(messages, llm.NewSystemMessage(a.systemPrompt))
	}
	messages = append(messages, llm.NewUserMessage(prompt))
	chunkChan := make(chan stream.StreamChunk, 100)
	errChan := make(chan error, 1)

	sendChunk := func(chunk stream.StreamChunk) {
		select {
		case chunkChan <- chunk:
		case <-ctx.Done():
		}
	}

	go func() {
		defer close(chunkChan)

		for iter := 0; iter < maxToolIterations; iter++ {
			req := llm.InferenceRequest{
				Messages: messages,
				Config: llm.InferenceConfig{
					Model:       resolvedModel,
					Temperature: 0.7,
					MaxTokens:   4096,
				},
			}
			filtered := a.allowedTools()
			if len(filtered) > 0 {
				req.Tools = filtered
			}

			prevReasoningLen := 0
				reasoningParser := NewReasoningParser()
				lastReasoningType := ReasoningPlan
				resp, toolCalls, err := client.GenerateStream(ctx, req, func(accumulated, reasoning string) {
					fmt.Printf("[DEBUG:adapter] callback called — accumulated len=%d, reasoning len=%d\n", len(accumulated), len(reasoning))
					if accumulated != "" {
						sendChunk(stream.StreamChunk{Type: stream.ChunkContent, Payload: accumulated})
					}
					if len(reasoning) > prevReasoningLen {
						delta := reasoning[prevReasoningLen:]
						prevReasoningLen = len(reasoning)
						if delta != "" {
							fmt.Printf("[DEBUG:adapter] reasoning delta len=%d total=%d — preview=%q\n", len(delta), len(reasoning), delta[:min(len(delta), 80)])
							detectedType, changed := reasoningParser.Feed(delta)
							chunkType := stream.ChunkThought
							switch detectedType {
							case ReasoningPlan:
								chunkType = stream.ChunkPlan
							case ReasoningExplore:
								chunkType = stream.ChunkExplore
							case ReasoningExec:
								chunkType = stream.ChunkExec
							case ReasoningRead:
								chunkType = stream.ChunkRead
							case ReasoningDiff:
								chunkType = stream.ChunkDiff
							}
							if changed {
								fmt.Printf("[DEBUG:adapter] reasoning type change: %s -> %s\n", lastReasoningType, detectedType)
								lastReasoningType = detectedType
							}
							sendChunk(stream.StreamChunk{Type: chunkType, Payload: delta})
						}
					}
				})

			if err != nil {
				fmt.Printf("[MultiLLMAdapter.GenerateStream] ERROR: %v\n", err)
				if a.onModelError != nil {
					providerName := model[:strings.Index(model, "/")]
					a.onModelError(providerName, resolvedModel, err)
				}
				sendChunk(stream.StreamChunk{Type: stream.ChunkError, Payload: err.Error()})
				errChan <- err
				return
			}

			if len(toolCalls) == 0 {
				fmt.Printf("[MultiLLMAdapter.GenerateStream] Done (no tool calls)\n")
				errChan <- nil
				return
			}

			fmt.Printf("[MultiLLMAdapter.GenerateStream] Got %d tool calls, executing...\n", len(toolCalls))

			astMsg := llm.NewAssistantMessage(resp)
			astMsg = llm.MessageWithToolCalls(astMsg, toolCalls)
			messages = append(messages, astMsg)

			for i, tc := range toolCalls {
				toolID := fmt.Sprintf("tool-%d-%d", iter, i)

				if a.permissionGuard != nil {
					allowed, reason, requestID := a.permissionGuard(tc.Function.Name, tc.Function.Arguments)
					if !allowed {
						pendingLabel := formatToolPending(tc.Function.Name, tc.Function.Arguments)
						sendChunk(stream.StreamChunk{
							Type: stream.ChunkAction, ID: toolID,
							Payload: pendingLabel, Meta: "blocked",
						})
						if requestID != "" {
							sendChunk(stream.StreamChunk{
								Type: stream.ChunkAction, ID: toolID,
								Payload: "⏳ " + tc.Function.Name + " — aguardando permissão",
								Meta:    "permission_required:" + requestID,
							})
							fmt.Printf("[MultiLLMAdapter] Permission required: %s (%s)\n", tc.Function.Name, reason)
							messages = append(messages, llm.NewToolMessage(tc.ID, fmt.Sprintf("[Permissão necessária: %s]", reason)))
						} else {
							errMsg := fmt.Sprintf("Bloqueado pelo modo: %s", reason)
							sendChunk(stream.StreamChunk{
								Type: stream.ChunkAction, ID: toolID,
								Payload: "❌ " + tc.Function.Name + " — " + errMsg,
								Meta:    "error",
							})
							fmt.Printf("[MultiLLMAdapter] Tool blocked: %s (%s)\n", tc.Function.Name, reason)
							messages = append(messages, llm.NewToolMessage(tc.ID, errMsg))
						}
						continue
					}
				}

				// Emit pending action before execution
				pendingLabel := formatToolPending(tc.Function.Name, tc.Function.Arguments)
				sendChunk(stream.StreamChunk{
					Type: stream.ChunkAction, ID: toolID,
					Payload: pendingLabel, Meta: "pending",
				})

				result := ""
				if a.toolHandler != nil {
					result, err = a.toolHandler(ctx, tc.Function.Name, tc.Function.Arguments)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
					}
				} else {
					result = fmt.Sprintf("Tool %s not available", tc.Function.Name)
				}
				messages = append(messages, llm.NewToolMessage(tc.ID, result))
				fmt.Printf("[MultiLLMAdapter.GenerateStream] Tool %s executed (result len=%d)\n", tc.Function.Name, len(result))

				// Emit completed action with result
				doneLabel := formatToolDone(tc.Function.Name, tc.Function.Arguments, result, err)
				sendChunk(stream.StreamChunk{
					Type: stream.ChunkAction, ID: toolID,
					Payload: doneLabel, Meta: result,
				})
			}
		}

		fmt.Printf("[MultiLLMAdapter.GenerateStream] Max tool iterations reached\n")
		errChan <- nil
	}()

	if a.emitter != nil {
		if err := a.streamMgr.StreamToEvents(ctx, sessionID, chunkChan, stream.EventEmitter(a.emitter)); err != nil {
			return err
		}
		if goroutineErr := <-errChan; goroutineErr != nil {
			return goroutineErr
		}
		return nil
	}

	return <-errChan
}

func formatToolPending(name, argsJSON string) string {
	switch name {
	case "exec":
		cmd := extractArg(argsJSON, "command")
		if cmd != "" {
			return "\U0001F5A5\uFE0F  " + cmd
		}
		return "\U0001F5A5\uFE0F  exec"
	case "read":
		path := extractArg(argsJSON, "path")
		if path != "" {
			return "\U0001F4D6  Reading " + shortenPath(path)
		}
		return "\U0001F4D6  read"
	case "write":
		path := extractArg(argsJSON, "path")
		if path != "" {
			return "\u270F\uFE0F  Writing " + shortenPath(path)
		}
		return "\u270F\uFE0F  write"
	case "search":
		query := extractArg(argsJSON, "query")
		if query != "" {
			return "\U0001F50D  Explore \"" + query + "\""
		}
		return "\U0001F50D  search"
	case "plan":
		task := extractArg(argsJSON, "task")
		if task != "" {
			return "\U0001F4CB  Planning: " + task
		}
		return "\U0001F4CB  plan"
	default:
		return "\U0001F916  " + name
	}
}

func formatToolDone(name, argsJSON, result string, execErr error) string {
	firstLine := firstNonEmptyLine(result)
	switch name {
	case "exec":
		cmd := extractArg(argsJSON, "command")
		label := "\U0001F5A5\uFE0F  "
		if cmd != "" {
			label += cmd
		} else {
			label += "exec"
		}
		if execErr != nil {
			label += " \u274C"
		} else if firstLine != "" && len(firstLine) < 80 {
			label += "  \u2705 " + firstLine
		} else {
			label += " \u2705"
		}
		return label
	case "read":
		path := extractArg(argsJSON, "path")
		label := "\U0001F4D6  Read "
		if path != "" {
			label += shortenPath(path)
		}
		return label
	case "write":
		path := extractArg(argsJSON, "path")
		label := "\u270F\uFE0F  Edited "
		if path != "" {
			label += "\U0001F4E4 " + shortenPath(path)
		}
		// Try to extract diff stats from result
		if !execErrFmt(execErr) {
			diffStr := extractDiffStats(result)
			if diffStr != "" {
				label += " " + diffStr
			}
		}
		return label
	case "search":
		query := extractArg(argsJSON, "query")
		label := "\U0001F50D  Explore"
		if query != "" {
			label += " \"" + query + "\""
		}
		// Count results from output
		lines := countNonEmptyLines(result)
		if lines > 0 {
			label += fmt.Sprintf("  %d results", lines)
		}
		return label
	case "plan":
		return "\U0001F4CB  Planned"
	default:
		return "\U0001F916  " + name
	}
}

func extractArg(argsJSON, key string) string {
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ""
	}
	v, ok := args[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func shortenPath(path string) string {
	// Take just the filename from a full path
	if idx := strings.LastIndex(path, "/"); idx >= 0 && idx < len(path)-1 {
		return path[idx+1:]
	}
	return path
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func countNonEmptyLines(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func execErrFmt(err error) bool {
	return err != nil
}

func extractDiffStats(result string) string {
	// Look for diff-like patterns like "+5 -4" in the result
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "+") && strings.Contains(line, "-") {
			return line
		}
	}
	return ""
}

type LLMStreamingClient interface {
	GenerateStream(ctx context.Context, sessionID string, prompt string, model string) error
	SetEmitter(emitter stream.EventEmitter)
	SetMode(mode string)
	SetSystemPrompt(prompt string)
	SetPermissionGuard(guard PermissionGuard)
}

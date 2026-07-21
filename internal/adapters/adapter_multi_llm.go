package adapters

import (
	"context"
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

type MultiLLMAdapter struct {
	providers    map[string]llm.StreamingLLMClient
	defaultModel string
	streamMgr    *stream.StreamManager
	emitter      stream.EventEmitter
	tools        []llm.ToolDefinition
}

func (a *MultiLLMAdapter) SetTools(tools []llm.ToolDefinition) {
	a.tools = tools
}

// resolveClient picks the right provider client for the given model string.
// Model must be in "provider/model" format. Returns nil if not found.
func (a *MultiLLMAdapter) resolveClient(model string) (llm.StreamingLLMClient, string) {
	if model == "" || !strings.Contains(model, "/") {
		return nil, model
	}
	providerName := model[:strings.Index(model, "/")]
	client, ok := a.providers[providerName]
	if !ok {
		return nil, model
	}
	return client, model
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
			errMsg := fmt.Sprintf("Provider para modelo '%s' não encontrado. Configure o provedor em Settings > Models.", model)
			ch <- core.LLMToken{Token: "[Error: " + errMsg + "]", Done: true}
			return
		}

		req := llm.InferenceRequest{
			UserPrompt: prompt,
			Config: llm.InferenceConfig{
				Model:       resolvedModel,
				Temperature: 0.7,
				MaxTokens:   4096,
			},
		}
		if len(a.tools) > 0 {
			req.Tools = a.tools
		}

		resp, err := client.GenerateStream(ctx, req, func(accumulated string) {
			ch <- core.LLMToken{Token: accumulated}
		})

		if err != nil {
			ch <- core.LLMToken{Token: "[Error: " + err.Error() + "]", Done: true}
			return
		}

		if resp != "" {
			ch <- core.LLMToken{Token: resp}
		}
		ch <- core.LLMToken{Done: true}
	}()

	return ch, nil
}

func (a *MultiLLMAdapter) GenerateStream(ctx context.Context, sessionID string, prompt string, model string) error {
	fmt.Printf("[MultiLLMAdapter.GenerateStream] ENTER session=%s prompt=%q model=%s\n", sessionID, prompt[:min(50, len(prompt))], model)

	if response, ok := llm.CheckStaticResponse(prompt); ok {
		fmt.Printf("[MultiLLMAdapter.GenerateStream] static route matched for %q\n", prompt[:min(20, len(prompt))])
		if a.emitter != nil {
			a.emitter("token-received", map[string]interface{}{
				"sessionID": sessionID,
				"token":     response,
			})
			a.emitter("stream-finished", map[string]interface{}{
				"sessionID": sessionID,
			})
		}
		return nil
	}

	canAttemptLLM := a.emitter != nil && len(a.providers) > 0

	if canAttemptLLM {
		client, resolvedModel := a.resolveClient(model)

		if client == nil {
			errMsg := fmt.Sprintf("Provider '%s' não encontrado para o modelo '%s'. Configure o provedor em Settings > Models.", model, model)
			fmt.Printf("[MultiLLMAdapter.GenerateStream] %s\n", errMsg)
			if a.emitter != nil {
				a.emitter("token-received", map[string]interface{}{
					"sessionID": sessionID,
					"token":     "[Error: " + errMsg + "]",
				})
				a.emitter("stream-finished", map[string]interface{}{
					"sessionID": sessionID,
				})
			}
			return fmt.Errorf(errMsg)
		}

		{
			ctx, cancel := a.streamMgr.Register(ctx, sessionID)
			defer func() {
				a.streamMgr.Unregister(sessionID)
				cancel()
			}()

			req := llm.InferenceRequest{
				UserPrompt: prompt,
				Config: llm.InferenceConfig{
					Model:       resolvedModel,
					Temperature: 0.7,
					MaxTokens:   4096,
				},
			}
			if len(a.tools) > 0 {
				req.Tools = a.tools
			}

			tokenChan := make(chan string, 100)
			go func() {
				defer close(tokenChan)
				fmt.Printf("[MultiLLMAdapter.GenerateStream] go: calling client.GenerateStream\n")
				_, err := client.GenerateStream(ctx, req, func(accumulated string) {
					fmt.Printf("[MultiLLMAdapter.GenerateStream] go: token accumulated: %q\n", accumulated)
					select {
					case tokenChan <- accumulated:
					case <-ctx.Done():
						fmt.Printf("[MultiLLMAdapter.GenerateStream] go: ctx done, dropping token\n")
					}
				})
				if err != nil {
					fmt.Printf("[MultiLLMAdapter.GenerateStream] go: LLM ERROR: %v\n", err)
					select {
					case tokenChan <- "[Error: " + err.Error() + "]":
					case <-ctx.Done():
					}
				} else {
					fmt.Printf("[MultiLLMAdapter.GenerateStream] go: client.GenerateStream returned OK\n")
				}
			}()

			fmt.Printf("[MultiLLMAdapter.GenerateStream] calling StreamToEvents\n")
			err := a.streamMgr.StreamToEvents(ctx, sessionID, tokenChan, a.emitter)
			fmt.Printf("[MultiLLMAdapter.GenerateStream] StreamToEvents returned: %v\n", err)
			return err
		}
	}

	// Fallback: handle static or no client case
	if response, ok := llm.CheckStaticResponse(prompt); ok {
		if a.emitter != nil {
			a.emitter("token-received", map[string]interface{}{
				"sessionID": sessionID,
				"token":     response,
			})
			a.emitter("stream-finished", map[string]interface{}{
				"sessionID": sessionID,
			})
		}
		return nil
	}

	if a.emitter != nil {
		a.emitter("stream-interrupted", map[string]interface{}{
			"sessionID": sessionID,
			"reason":    "no LLM client or emitter configured",
		})
	}
	fmt.Printf("[MultiLLMAdapter.GenerateStream] EXIT - no client or static response\n")
	return nil
}

type LLMStreamingClient interface {
	GenerateStream(ctx context.Context, sessionID string, prompt string, model string) error
	SetEmitter(emitter stream.EventEmitter)
}

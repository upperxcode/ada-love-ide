package core

import (
	"context"
	"time"
)

// MockCompactor concatena histórico até o limite de tokens (simplificado).
type MockCompactor struct{}

func NewMockCompactor() *MockCompactor { return &MockCompactor{} }

func (m *MockCompactor) Compact(_ context.Context, systemPrompt string, history []string, limit int) (string, error) {
	if len(history) == 0 {
		return systemPrompt, nil
	}
	result := systemPrompt
	for _, msg := range history {
		if len(result)+len(msg) > limit {
			break
		}
		result += "\n" + msg
	}
	return result, nil
}

func (m *MockCompactor) CountTokens(text string) int {
	return len(text) / 4
}

func (m *MockCompactor) CompactWithOverhead(_ context.Context, history []Message, overheadTokens int) (string, error) {
	limit := 8000 - overheadTokens*4
	if limit < 0 {
		limit = 0
	}
	strs := make([]string, len(history))
	for i, h := range history {
		strs[i] = h.Role + ": " + h.Content
	}
	return m.Compact(context.Background(), "You are a helpful AI assistant.", strs, limit)
}

// MockExecutor retorna output mock para comandos.
type MockExecutor struct{}

func NewMockExecutor() *MockExecutor { return &MockExecutor{} }

func (m *MockExecutor) ExecuteCommand(_ context.Context, sessionID, cmd string, args []string) (string, error) {
	return "[mock] " + cmd, nil
}

// MockLLMClient stream de tokens mock com delay realista.
type MockLLMClient struct{}

func NewMockLLMClient() *MockLLMClient { return &MockLLMClient{} }

func (m *MockLLMClient) Generate(ctx context.Context, prompt string, model string) (<-chan LLMToken, error) {
	ch := make(chan LLMToken, 10)
	go func() {
		defer close(ch)
		response := "Esta é uma resposta mock. O cliente LLM real será injetido posteriormente."
		for _, r := range response {
			select {
			case <-ctx.Done():
				return
			case ch <- LLMToken{Token: string(r)}:
			case <-time.After(10 * time.Millisecond):
			}
		}
		ch <- LLMToken{Done: true}
	}()
	return ch, nil
}

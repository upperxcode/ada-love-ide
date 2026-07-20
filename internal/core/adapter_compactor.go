package core

import (
	"context"

	contextmanager "github.com/upperxcode/ada-context"
)

// CompactorAdapter adapta o ada-context para a interface core.Compactor.
type CompactorAdapter struct {
	compactor *contextmanager.ContextCompactor
}

// NewCompactorAdapter cria um novo adapter.
func NewCompactorAdapter(maxTokens, keepLatest int, systemPrompt string) *CompactorAdapter {
	tokenizer := contextmanager.NewFastTokenizer()
	cfg := contextmanager.CompactorConfig{
		MaxTokens:       maxTokens,
		SystemPrompt:    systemPrompt,
		KeepLatestCount: keepLatest,
	}
	return &CompactorAdapter{
		compactor: contextmanager.NewContextCompactor(tokenizer, cfg),
	}
}

// Compact implementa core.Compactor.
func (a *CompactorAdapter) Compact(ctx context.Context, systemPrompt string, history []string, limit int) (string, error) {
	return a.compactor.Compact(ctx, history)
}

// CountTokens retorna a contagem de tokens para um bloco de texto.
func (a *CompactorAdapter) CountTokens(text string) int {
	return a.compactor.CountTokens(text)
}

// CompactWithOverhead compacta o histórico descontando overheadTokens externos (ex: wiki).
func (a *CompactorAdapter) CompactWithOverhead(ctx context.Context, history []Message, overheadTokens int) (string, error) {
	msgs := make([]contextmanager.Message, len(history))
	for i, m := range history {
		msgs[i] = contextmanager.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return a.compactor.CompactWithOverhead(ctx, msgs, overheadTokens)
}

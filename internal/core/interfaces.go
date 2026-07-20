package core

import "context"

type StorageEngine interface {
	GetMessagesBySession(sessionID string) ([]Message, error)
	SaveMessage(msg Message) error
	GetGreetings() ([]Greeting, error)
	DeleteMessages(sessionID string) error
	GetSession(id string) (*Session, bool)
}

type LLMClient interface {
	Generate(ctx context.Context, prompt string, model string) (<-chan LLMToken, error)
}

type Compactor interface {
	Compact(ctx context.Context, systemPrompt string, history []string, limit int) (string, error)
	CountTokens(text string) int
	CompactWithOverhead(ctx context.Context, history []Message, overheadTokens int) (string, error)
}

type Executor interface {
	ExecuteCommand(ctx context.Context, sessionID, cmd string, args []string) (string, error)
}

// Emitter é a interface de emissão de eventos usada pelo Orchestrator.
// Definida em core para evitar o ciclo de importações com o pacote chat.
type Emitter interface {
	Emit(event string, data ...any)
}

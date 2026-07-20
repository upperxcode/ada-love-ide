// Package core define as interfaces e o orchestrator do pipeline de chat.
// Extraído do ada-love-core para integrar ao Wails.
package core

// Message representa uma mensagem de chat no formato do core.
type Message struct {
	ID        string
	SessionID string
	Role      string // "user" ou "assistant"
	Content   string
	CreatedAt string
}

// DecisionEvent representa um evento de decisão do orchestrator
// com a cadeia de pensamento do modelo.
type DecisionEvent struct {
	SessionID string `json:"session_id"`
	Reasoning string `json:"reasoning"`
	NextAgent string `json:"next_agent"`
	Task      string `json:"task"`
	SubTasks  int    `json:"sub_tasks"`
}

// LLMToken representa um token streaming do LLM.
type LLMToken struct {
	Token        string `json:"token"`
	Done         bool   `json:"done"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// Greeting representa uma saudação estática para evitar chamadas ao LLM.
type Greeting struct {
	Patterns string // padrões separados por vírgula
	Response string // resposta estática
}

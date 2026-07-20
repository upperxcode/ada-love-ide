package core

import "time"

// RawMessage representa uma mensagem no formato de persistência (banco).
// Diferente de Message (usado no pipeline do orchestrator), RawMessage
// contém campos de controle como ToolCalls e Time.
type RawMessage struct {
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolCalls  []any     `json:"tool_calls"`
	ToolCallID string    `json:"tool_call_id"`
	Time       time.Time `json:"time"`
}

// Session representa uma sessão de chat persistida.
type Session struct {
	ID              string       `json:"id"`
	WorkspaceID     string       `json:"workspace_id"`
	WorkerName      string       `json:"worker_name"`
	ParentSessionID string       `json:"parent_session_id"`
	Title           string       `json:"title"`
	Summary         string       `json:"summary"`
	Model           string       `json:"model"`
	Provider        string       `json:"provider"`
	Mode            string       `json:"mode"`
	Thinking        string       `json:"thinking"`
	Messages        []RawMessage `json:"messages"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Pinned          bool         `json:"pinned"`
}

// NewSession cria uma nova Session com valores padrão.
func NewSession(id, workspaceID, workerName string) Session {
	now := time.Now()
	return Session{
		ID:          id,
		WorkspaceID: workspaceID,
		WorkerName:  workerName,
		Title:       "Novo chat",
		Mode:        "ask",
		Messages:    []RawMessage{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// HasThinking retorna true se a sessão está em modo thinking.
func (s *Session) HasThinking() bool {
	return s.Thinking == "high" || s.Thinking == "medium" || s.Thinking == "low"
}

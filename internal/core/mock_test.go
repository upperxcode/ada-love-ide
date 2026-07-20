package core

import "sync"

// MockStorage implementa StorageEngine para testes.
type MockStorage struct {
	mu        sync.Mutex
	greetings []Greeting
	messages  []Message
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		greetings: []Greeting{
			{Patterns: "hello,hi,hey", Response: "Hello! How can I help?"},
			{Patterns: "bom dia", Response: "Bom dia! Como posso ajudar?"},
		},
	}
}

func (m *MockStorage) GetGreetings() ([]Greeting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Greeting, len(m.greetings))
	copy(out, m.greetings)
	return out, nil
}

func (m *MockStorage) GetMessagesBySession(sessionID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Message
	for _, msg := range m.messages {
		if msg.SessionID == sessionID {
			out = append(out, msg)
		}
	}
	return out, nil
}

func (m *MockStorage) SaveMessage(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

// Messages retorna cópia das mensagens salvas (para asserts).
func (m *MockStorage) Messages() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Message, len(m.messages))
	copy(out, m.messages)
	return out
}

// GetSession retorna nil, false (mock simples - testes não usam esse método atualmente).
func (m *MockStorage) GetSession(id string) (*Session, bool) {
	return nil, false
}

// DeleteMessages remove todas as mensagens de uma sessão.
func (m *MockStorage) DeleteMessages(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filtered []Message
	for _, msg := range m.messages {
		if msg.SessionID != sessionID {
			filtered = append(filtered, msg)
		}
	}
	m.messages = filtered
	return nil
}

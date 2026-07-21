package db

import (
	"context"
	"fmt"
	"time"

	core "ada-love-core"
)

// StorageAdapter envolve *Store e implementa core.StorageEngine.
type StorageAdapter struct {
	store *Store
}

// NewStorageAdapter cria um adapter que adapta *Store para core.StorageEngine.
func NewStorageAdapter(s *Store) *StorageAdapter {
	return &StorageAdapter{store: s}
}

// GetMessagesBySession converte storage Message → core.Message.
func (a *StorageAdapter) GetMessagesBySession(sessionID string) ([]core.Message, error) {
	ctx := context.Background()
	raw, err := a.store.Sessions().GetMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]core.Message, 0, len(raw))
	for _, m := range raw {
		out = append(out, core.Message{
			ID:        fmt.Sprintf("%d", m.ID),
			SessionID: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.Time.Format(time.RFC3339),
		})
	}
	return out, nil
}

// SaveMessage converte core.Message → core.RawMessage e salva via store.AppendMessage.
func (a *StorageAdapter) SaveMessage(msg core.Message) error {
	t, _ := time.Parse(time.RFC3339, msg.CreatedAt)
	a.store.AppendMessage(msg.SessionID, core.RawMessage{
		Role:    msg.Role,
		Content: msg.Content,
		Time:    t,
	})
	return nil
}

// GetGreetings retorna as saudações estáticas configuradas.
func (a *StorageAdapter) GetGreetings() ([]core.Greeting, error) {
	return nil, nil
}

// DeleteMessages remove todas as mensagens de uma sessão.
func (a *StorageAdapter) DeleteMessages(sessionID string) error {
	ctx := context.Background()
	return a.store.Sessions().DeleteMessages(ctx, sessionID)
}

// GetSession retorna a sessão pelo ID.
func (a *StorageAdapter) GetSession(id string) (*core.Session, bool) {
	return a.store.GetSession(id)
}

package sessionstore

import (
	"strings"
	"time"

	"ada-love-ide/internal/core"
)

// Delete remove a sessão do banco.
func (s *Saver) Delete(id string) { s.db.DeleteSession(id) }

// Rename aplica novo título e retorna a sessão atualizada.
func (s *Saver) Rename(id, newTitle string) (core.Session, error) {
	sess, ok := s.db.GetSession(id)
	if !ok {
		return core.Session{}, ErrNotFound
	}
	newTitle = truncate(newTitle, 60)
	if newTitle != "" {
		sess.Title = newTitle
		sess.UpdatedAt = time.Now()
		s.db.PutSession(sess)
	}
	return *sess, nil
}

// TogglePin inverte o flag pinned.
func (s *Saver) TogglePin(id string) error {
	sess, ok := s.db.GetSession(id)
	if !ok {
		return ErrNotFound
	}
	sess.Pinned = !sess.Pinned
	s.db.PutSession(sess)
	return nil
}

// SetConfig sobrescreve os campos model/provider/mode/thinking.
// Se provider estiver vazio e model contiver "/", extrai o provider do modelo.
func (s *Saver) SetConfig(id, model, provider, mode, thinking string) error {
	sess, ok := s.db.GetSession(id)
	if !ok {
		return ErrNotFound
	}

	// Se provider não foi passado mas model tem formato "provider/model"
	if provider == "" && model != "" {
		parts := strings.SplitN(model, "/", 2)
		if len(parts) == 2 {
			provider = parts[0]
			sess.Provider = provider
		}
	}

	sess.Model = model
	sess.Mode = mode
	sess.Thinking = thinking
	sess.UpdatedAt = time.Now()
	s.db.PutSession(sess)
	return nil
}

// AppendMessage adiciona uma mensagem na sessão (usado pelo chat).
func (s *Saver) AppendMessage(id string, msg core.RawMessage) error {
	sess, ok := s.db.GetSession(id)
	if !ok {
		return ErrNotFound
	}
	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()
	s.db.PutSession(sess)
	return nil
}

// GetThinking retorna o nível de thinking da sessão.
func (s *Saver) GetThinking(id string) (string, error) {
	sess, ok := s.db.GetSession(id)
	if !ok {
		return "", ErrNotFound
	}
	return sess.Thinking, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Package sessionfetch centraliza todas as operações de LEITURA de
// sessões do banco. Tecnicamente: "recuperar do banco".
package sessionfetch

import (
	"ada-love-ide/internal/core"
	"ada-love-ide/internal/db"
)

// Fetcher envolve um *db.Store com operações de leitura de sessões.
type Fetcher struct{ db *db.Store }

func New(db *db.Store) *Fetcher { return &Fetcher{db: db} }

// List retorna todas as sessões de um workspace, incluindo mensagens.
func (f *Fetcher) List(workspaceID string) []core.Session {
	raw := f.db.ListSessions(workspaceID)
	out := make([]core.Session, 0, len(raw))
	for _, s := range raw {
		s.Messages = f.db.GetMessages(s.ID)
		out = append(out, *s)
	}
	return out
}

// Get retorna uma sessão pelo ID.
func (f *Fetcher) Get(id string) (core.Session, bool) {
	v, ok := f.db.GetSession(id)
	if !ok {
		return core.Session{}, false
	}
	return *v, true
}

// Exists verifica se a sessão existe.
func (f *Fetcher) Exists(id string) bool {
	_, ok := f.db.GetSession(id)
	return ok
}

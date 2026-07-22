// Package sessionstore centraliza todas as operações de ESCRITA de
// sessões no banco. Tecnicamente: "salvar no banco".
package sessionstore

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	core "ada-love-core"
	"ada-love-ide/internal/db"
)

var ErrNotFound = errors.New("sessão não encontrada")

// Saver envolve um *db.Store com operações de mutação de sessões.
type Saver struct{ db *db.Store }

func New(db *db.Store) *Saver { return &Saver{db: db} }

// Create cria uma nova sessão no workspace/worker indicados.
func (s *Saver) Create(workspaceID, workerName string) core.Session {
	id := "sess-" + time.Now().Format("20060102150405") + "-" + randID()
	sess := core.NewSession(id, workspaceID, workerName)
	s.db.InsertSession(&sess)
	return sess
}

// CreateSummarized cria uma nova sessão filha de sourceSessionID.
func (s *Saver) CreateSummarized(workspaceID, workerName, sourceSessionID string) (core.Session, error) {
	src, ok := s.db.GetSession(sourceSessionID)
	if !ok {
		return core.Session{}, ErrNotFound
	}
	child := s.Create(workspaceID, workerName)
	child.ParentSessionID = sourceSessionID
	if strings.TrimSpace(src.Summary) != "" {
		child.Title = "Resumo"
	} else {
		child.Title = "Resumo: " + src.Title
	}
	// copia última mensagem como ponto de partida
	if len(src.Messages) > 0 {
		child.Messages = []core.RawMessage{
			{Role: "system", Content: "Sessão sumarizada a partir de: " + src.Title, Time: time.Now()},
		}
	}
	s.db.PutSession(&child)
	return child, nil
}

func randID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Package modelselect centraliza a descoberta/seleção de modelos
// disponíveis para o chat. Tecnicamente: "escolher um modelo".
package modelselect

import (
	"ada-love-ide/internal/db"
)

// ModelEntry é um modelo exposto para a UI.
type ModelEntry struct {
	Key         string `json:"key"`
	Provider    string `json:"provider"`
	ModelName   string `json:"model_name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

// Selector envolve o *db.Store para listar/escolher modelos.
type Selector struct{ db *db.Store }

func New(db *db.Store) *Selector { return &Selector{db: db} }

// List varre todos os providers e junta todos os modelos.
func (s *Selector) List() []ModelEntry {
	providers := s.db.ListProviders()
	out := []ModelEntry{}
	for providerName, cfg := range providers {
		for modelName := range cfg.Models {
			out = append(out, ModelEntry{
				Key:         providerName + "/" + modelName,
				Provider:    providerName,
				ModelName:   modelName,
				DisplayName: modelName,
				Enabled:     true,
			})
		}
	}
	return out
}

// Pick resolves uma chave "provider/model" para entry real.
func (s *Selector) Pick(key string) (ModelEntry, bool) {
	for _, e := range s.List() {
		if e.Key == key {
			return e, true
		}
	}
	return ModelEntry{}, false
}

// Default retorna o primeiro modelo listado, ou "" se vazio.
func (s *Selector) Default() string {
	list := s.List()
	if len(list) == 0 {
		return ""
	}
	return list[0].Key
}

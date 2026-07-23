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

// List varre todos os providers do worker 'ada' (padrão) e junta todos os modelos.
func (s *Selector) List() []ModelEntry {
	return s.ListByWorker("ada")
}

// ListByWorker varre apenas providers de um worker específico.
func (s *Selector) ListByWorker(worker string) []ModelEntry {
	providers := s.db.ListProvidersByWorker(worker)
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

// Pick resolve uma chave "provider/model" para entry real, filtrando pelo worker informado.
func (s *Selector) Pick(key string, worker ...string) (ModelEntry, bool) {
	w := "ada"
	if len(worker) > 0 {
		w = worker[0]
	}
	for _, e := range s.ListByWorker(w) {
		if e.Key == key {
			return e, true
		}
	}
	return ModelEntry{}, false
}

// Default retorna o primeiro modelo listado do worker informado, ou "" se vazio.
func (s *Selector) Default(worker ...string) string {
	w := "ada"
	if len(worker) > 0 {
		w = worker[0]
	}
	list := s.ListByWorker(w)
	if len(list) == 0 {
		return ""
	}
	return list[0].Key
}

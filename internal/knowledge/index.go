package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

// ═══════════════════════════════════════════════════════════════
// Índice semântico de knowledge
//
// Mantém vetores de embedding em memória (map[int64][]Entry),
// indexados por workspaceID.
//
// REGRA: erros de embedding nunca param o processo.
// Em vez disso, emitem um alerta (fmt.Printf) e retornam TODO
// o conteúdo como fallback, gastando mais tokens mas não
// interrompendo o fluxo.
// ═══════════════════════════════════════════════════════════════

// Index mantém os vetores de embedding em memória e responde
// buscas semânticas por similaridade de cosseno.
type Index struct {
	embedder Embedder
	mu       sync.RWMutex
	entries  map[int64][]Entry // workspaceID → entries
}

// NewIndex cria um novo index semântico.
func NewIndex(embedder Embedder) *Index {
	return &Index{
		embedder: embedder,
		entries:  make(map[int64][]Entry),
	}
}

// IndexWorkspace gera embeddings para todos os knowledge items de um workspace
// e substitui quaisquer entries anteriores.
//
// Se um item individual falhar ao ser embedado, ele é pulado com um alerta
// em vez de abortar. Itens que embedaram com sucesso ficam disponíveis.
func (idx *Index) IndexWorkspace(ctx context.Context, workspaceID int64, items []string) {
	entries := make([]Entry, 0, len(items))

	for i, item := range items {
		vec, err := idx.embedder.Embed(ctx, item)
		if err != nil {
			fmt.Printf("[Knowledge] WARNING: index workspace %d item %d falhou ao embedar (%v) — pulando item\n",
				workspaceID, i, err)
			continue
		}
		entries = append(entries, Entry{
			KnowledgeID: int64(i),
			WorkspaceID: workspaceID,
			Text:        item,
			Vector:      vec,
		})
	}

	idx.mu.Lock()
	idx.entries[workspaceID] = entries
	idx.mu.Unlock()
}

// Search encontra os top-K itens mais relevantes para a query.
//
// Se o embedding da query falhar, um alerta é emitido e TODO o conteúdo
// do workspace é retornado como fallback (gasta mais tokens mas não para).
//
// Retorna nil se o workspace não tiver conhecimento indexado.
func (idx *Index) Search(ctx context.Context, query string, workspaceID int64, topK int) []string {
	idx.mu.RLock()
	entries := idx.entries[workspaceID]
	idx.mu.RUnlock()

	if len(entries) == 0 {
		return nil
	}

	queryVec, err := idx.embedder.Embed(ctx, query)
	if err != nil {
		fmt.Printf("[Knowledge] WARNING: search embed query falhou para workspace %d (%v) — fallback para conteúdo completo\n",
			workspaceID, err)
		return idx.AllTexts(workspaceID)
	}

	// cosine similarity
	type scored struct {
		text string
		sim  float32
	}
	results := make([]scored, 0, len(entries))
	for _, e := range entries {
		sim := cosineSimilarity(queryVec, e.Vector)
		results = append(results, scored{text: e.Text, sim: sim})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].sim > results[j].sim
	})

	if topK > len(results) {
		topK = len(results)
	}
	texts := make([]string, topK)
	for i, r := range results[:topK] {
		texts[i] = r.text
	}
	return texts
}

// AllTexts retorna todos os textos de um workspace (sem vetores).
// Usado como fallback quando o embedding da query falha.
func (idx *Index) AllTexts(workspaceID int64) []string {
	idx.mu.RLock()
	entries := idx.entries[workspaceID]
	idx.mu.RUnlock()

	texts := make([]string, len(entries))
	for i, e := range entries {
		texts[i] = e.Text
	}
	return texts
}

// Count retorna o número de entries indexadas para um workspace.
func (idx *Index) Count(workspaceID int64) int {
	idx.mu.RLock()
	n := len(idx.entries[workspaceID])
	idx.mu.RUnlock()
	return n
}

// ── cosine similarity ────────────────────────────────────────

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

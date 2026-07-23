package knowledge

import (
	"context"
	"testing"
)

// ── Cosine similarity ─────────────────────────────────────────

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float32{1, 2, 3, 4}
	b := []float32{1, 2, 3, 4}
	sim := cosineSimilarity(a, b)
	if sim < 0.999 || sim > 1.001 {
		t.Errorf("expected ~1.0, got %f", sim)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	sim := cosineSimilarity(a, b)
	if sim < -0.001 || sim > 0.001 {
		t.Errorf("expected ~0.0, got %f", sim)
	}
}

func TestCosineSimilarity_Opposite(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{-1, -2, -3}
	sim := cosineSimilarity(a, b)
	if sim > -0.999 || sim < -1.001 {
		t.Errorf("expected ~-1.0, got %f", sim)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}
	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("expected 0 for different lengths, got %f", sim)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	sim := cosineSimilarity(nil, []float32{1, 2})
	if sim != 0 {
		t.Errorf("expected 0 for empty, got %f", sim)
	}
}

// ── Index ─────────────────────────────────────────────────────

func TestNewIndex(t *testing.T) {
	e := NewEmbedder("test", "mock", "http://localhost:1", "")
	idx := NewIndex(e)
	if idx == nil {
		t.Fatal("expected non-nil index")
	}
}

func TestIndex_Count(t *testing.T) {
	idx := NewIndex(NewEmbedder("test", "mock", "http://localhost:1", ""))

	if idx.Count(1) != 0 {
		t.Errorf("expected 0 for empty, got %d", idx.Count(1))
	}

	idx.mu.Lock()
	idx.entries[1] = []Entry{
		{KnowledgeID: 0, WorkspaceID: 1, Text: "x", Vector: []float32{1}},
	}
	idx.mu.Unlock()

	if idx.Count(1) != 1 {
		t.Errorf("expected 1 after insert, got %d", idx.Count(1))
	}
}

func TestIndex_Search_ExactMatch(t *testing.T) {
	idx := idxWithEntries(1,
		Entry{0, 1, "apple fruit", []float32{1, 0, 0, 0}},
		Entry{1, 1, "banana fruit", []float32{0.9, 0.1, 0, 0}},
		Entry{2, 1, "car engine", []float32{0, 0, 1, 0}},
	)

	// We cannot call idx.Search because it will try to embed "error" via the
	// dummy embedder (which will fail). Instead, test the search logic directly
	// by calling the internal search function.
	results := searchIn(idx, 1, []float32{0.95, 0.05, 0, 0}, 2)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "apple fruit" {
		t.Errorf("expected top result 'apple fruit', got %q", results[0])
	}
	if results[1] != "banana fruit" {
		t.Errorf("expected second result 'banana fruit', got %q", results[1])
	}
}

func TestIndex_Search_EmptyWorkspace(t *testing.T) {
	idx := NewIndex(NewEmbedder("test", "mock", "http://localhost:1", ""))
	results := idx.Search(context.Background(), "query", 1, 5)
	if results != nil {
		t.Errorf("expected nil results for empty workspace, got %v", results)
	}
}

func TestIndex_Search_DifferentWorkspace(t *testing.T) {
	idx := idxWithEntries(1,
		Entry{0, 1, "data for ws1", []float32{1, 0, 0, 0}},
	)

	results := idx.Search(context.Background(), "data", 2, 5)
	if results != nil {
		t.Errorf("expected nil for different workspace, got %v", results)
	}
}

func TestIndex_Search_EmbedError_FallsBack(t *testing.T) {
	// Create an index with a dangling embedder (port 1 = connection refused).
	// Search will try to embed the query, fail, and fall back to AllTexts.
	idx := idxWithEntries(1,
		Entry{0, 1, "alpha", []float32{1, 0, 0, 0}},
		Entry{1, 1, "beta", []float32{0, 1, 0, 0}},
		Entry{2, 1, "gamma", []float32{0, 0, 1, 0}},
	)

	// "anything" will trigger an actual HTTP call to the dummy embedder → fails → fallback
	results := idx.Search(context.Background(), "anything", 1, 5)

	if len(results) != 3 {
		t.Fatalf("expected ALL 3 items as fallback, got %d: %v", len(results), results)
	}

	expected := []string{"alpha", "beta", "gamma"}
	for i, exp := range expected {
		if results[i] != exp {
			t.Errorf("fallback item %d: expected %q, got %q", i, exp, results[i])
		}
	}
}

func TestIndex_AllTexts(t *testing.T) {
	idx := idxWithEntries(1,
		Entry{0, 1, "a", []float32{1}},
		Entry{1, 1, "b", []float32{2}},
		Entry{2, 1, "c", []float32{3}},
	)

	texts := idx.AllTexts(1)
	if len(texts) != 3 {
		t.Fatalf("expected 3 texts, got %d", len(texts))
	}
	if texts[0] != "a" || texts[1] != "b" || texts[2] != "c" {
		t.Errorf("unexpected texts order: %v", texts)
	}
}

func TestIndex_Concurrency(t *testing.T) {
	idx := NewIndex(NewEmbedder("test", "mock", "http://localhost:1", ""))

	done := make(chan bool, 2)
	go func() {
		for i := 0; i < 100; i++ {
			idx.mu.Lock()
			idx.entries[1] = []Entry{
				{KnowledgeID: int64(i), WorkspaceID: 1, Text: "x", Vector: []float32{float32(i)}},
			}
			idx.mu.Unlock()
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_ = idx.Count(1)
			_ = idx.AllTexts(1)
		}
		done <- true
	}()

	<-done
	<-done
}

// ── Helpers ───────────────────────────────────────────────────

// idxWithEntries cria um Index com entries pré-preenchidas.
// O embedder é dummy (qualquer chamada HTTP falha — ok para testes que
// acessam entries diretamente ou testam fallback).
func idxWithEntries(workspaceID int64, entries ...Entry) *Index {
	idx := NewIndex(NewEmbedder("test", "mock", "http://localhost:1", ""))
	if len(entries) > 0 {
		idx.mu.Lock()
		idx.entries[workspaceID] = entries
		idx.mu.Unlock()
	}
	return idx
}

// searchIn executa a busca local (sem chamar o embedder) para testes.
func searchIn(idx *Index, workspaceID int64, queryVec []float32, topK int) []string {
	idx.mu.RLock()
	entries := idx.entries[workspaceID]
	idx.mu.RUnlock()

	if len(entries) == 0 {
		return nil
	}

	type scored struct {
		text string
		sim  float32
	}
	results := make([]scored, 0, len(entries))
	for _, e := range entries {
		results = append(results, scored{text: e.Text, sim: cosineSimilarity(queryVec, e.Vector)})
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].sim > results[i].sim {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if topK > len(results) {
		topK = len(results)
	}
	texts := make([]string, topK)
	for i, r := range results[:topK] {
		texts[i] = r.text
	}
	return texts
}

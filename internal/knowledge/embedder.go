package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Embedder gera vetores de embedding chamando uma API compatível com OpenAI.
type Embedder struct {
	provider string // nome do provider (informativo, ex: "openai")
	model    string // nome do modelo, ex: "text-embedding-3-small"
	baseURL  string // URL base da API (já inclui /v1, ex: "https://api.openai.com/v1")
	apiKey   string // chave da API
	client   *http.Client
}

// NewEmbedder cria um Embedder que chama POST {baseURL}/embeddings.
func NewEmbedder(provider, model, baseURL, apiKey string) Embedder {
	return Embedder{
		provider: provider,
		model:    model,
		baseURL:  baseURL,
		apiKey:   apiKey,
		client:   &http.Client{},
	}
}

// Provider retorna o nome do provider.
func (e Embedder) Provider() string { return e.provider }

// Model retorna o nome do modelo de embedding.
func (e Embedder) Model() string { return e.model }

// Embed gera um vetor de embedding para o texto fornecido.
func (e Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]any{
		"input": text,
		"model": e.model,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedder: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.baseURL+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("embedder: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embedder: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embedder: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedder: API %d: %s", resp.StatusCode, string(respData))
	}

	var apiResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("embedder: parse response: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("embedder: no data in response")
	}

	raw := apiResp.Data[0].Embedding
	vec := make([]float32, len(raw))
	for i, v := range raw {
		vec[i] = float32(v)
	}
	return vec, nil
}

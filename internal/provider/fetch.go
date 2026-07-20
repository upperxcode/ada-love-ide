package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ada-love-ide/internal/config/provider"
)

// DiscoveredModel is one row returned by a provider's /models endpoint, with
// its inferred capabilities attached.
type DiscoveredModel struct {
	ID           string
	AlreadyAdded bool
	Capabilities Capabilities
}

// ErrNoAPIKey is returned when a connection type requires a key and none is
// configured. The UI surfaces a hint to open the API-keys modal.
var ErrNoAPIKey = errors.New("provider requires an api key, none configured")

var fetchClient = &http.Client{Timeout: 15 * time.Second}

// FetchModels calls the provider's model-list endpoint, decodes it through the
// adapter matching cfg.TypeConnection, and attaches inferred capabilities.
// AlreadyAdded marks which returned ids already exist in cfg.Models.
func FetchModels(ctx context.Context, cfg provider.ProviderConfig) ([]DiscoveredModel, error) {
	if NeedsAPIKey(cfg.TypeConnection) {
		if len(cfg.APIKeys) == 0 || cfg.APIKeys[0].Key == "" {
			return nil, ErrNoAPIKey
		}
	}

	url := strings.TrimRight(cfg.APIURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if NeedsAPIKey(cfg.TypeConnection) {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKeys[0].Key)
	}

	resp, err := fetchClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("provider %s: HTTP %d: %s", cfg.TypeConnection, resp.StatusCode, body)
	}

	ids, err := decodeModelIDs(cfg.TypeConnection, resp.Body)
	if err != nil {
		return nil, err
	}

	out := make([]DiscoveredModel, 0, len(ids))
	for _, id := range ids {
		_, already := cfg.Models[id]
		out = append(out, DiscoveredModel{
			ID:           id,
			AlreadyAdded: already,
			Capabilities: InferCapabilities(id),
		})
	}
	return out, nil
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// decodeModelIDs reads the /models response. Today every supported type returns
// the OpenAI-compatible shape. When Anthropic or Gemini are wired with their
// own adapters, branch on type here.
func decodeModelIDs(typeConnection string, body io.Reader) ([]string, error) {
	var resp openAIModelsResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode models: %w", err)
	}
	ids := make([]string, 0, len(resp.Data))
	for _, m := range resp.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

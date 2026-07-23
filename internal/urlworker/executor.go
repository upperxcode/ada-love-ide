package urlworker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmitterFn func(event string, data ...any)

type ModelInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
}

// ExecuteChat sends a chat message via HTTP and returns the response.
func ExecuteChat(req Request, emit EmitterFn) (string, error) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: timeout}

	if req.Stream {
		return streamResponse(ctx, client, httpReq, emit)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), nil
}

// streamResponse handles SSE streaming.
func streamResponse(ctx context.Context, client *http.Client, req *http.Request, emit EmitterFn) (string, error) {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// For streaming, remove client-level timeout — rely on context
	client.Timeout = 0

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("SSE request failed: %w", err)
	}
	defer resp.Body.Close()

	// Use a goroutine to read so we can select on ctx.Done()
	type readResult struct {
		text string
		err  error
	}
	lines := make(chan readResult, 64)

	go func() {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			lines <- readResult{text: scanner.Text()}
		}
		lines <- readResult{err: scanner.Err()}
		close(lines)
	}()

	var fullOutput strings.Builder
	for {
		select {
		case <-ctx.Done():
			return strings.TrimSpace(fullOutput.String()), ctx.Err()
		case result, ok := <-lines:
			if !ok {
				return strings.TrimSpace(fullOutput.String()), nil
			}
			if result.err != nil {
				return strings.TrimSpace(fullOutput.String()), result.err
			}
			line := result.text
			if line == "" {
				continue
			}

			if line == "data: [DONE]" {
				goto done
			}

			if strings.HasPrefix(line, "data: ") {
				chunk := strings.TrimPrefix(line, "data: ")
				fullOutput.WriteString(chunk)
				emitSSE(chunk, emit)
			}
		}
	}
done:
	return strings.TrimSpace(fullOutput.String()), nil
}

func emitSSE(data string, emit EmitterFn) {
	if emit == nil {
		return
	}
	var raw map[string]any
	if json.Unmarshal([]byte(data), &raw) == nil {
		if t, ok := raw["type"].(string); ok {
			switch t {
			case "message", "content":
				if c, ok := raw["content"].(string); ok {
					emit("chat:delta", map[string]any{"content": c})
				}
			case "thinking", "thought":
				if c, ok := raw["content"].(string); ok {
					emit("chat:thinking", map[string]any{"content": c})
				}
			case "error":
				if m, ok := raw["message"].(string); ok {
					emit("chat:error", map[string]any{"error": m})
				}
			}
			return
		}
		if c, ok := raw["content"].(string); ok {
			emit("chat:delta", map[string]any{"content": c})
		} else if c, ok := raw["response"].(string); ok {
			emit("chat:delta", map[string]any{"content": c})
		} else {
			emit("chat:delta", map[string]any{"content": data})
		}
	} else {
		emit("chat:delta", map[string]any{"content": data})
	}
}

// FetchModels calls the models endpoint and parses provider/model lines.
func FetchModels(req Request) ([]ModelInfo, error) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create models request: %w", err)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("models request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	raw := strings.TrimSpace(string(body))
	fmt.Printf("[FetchModels] HTTP %d, body length %d\n", resp.StatusCode, len(raw))
	fmt.Printf("[FetchModels] raw body preview: %s\n", raw[:min(len(raw), 300)])
	if raw == "" {
		return nil, nil
	}

	switch req.ModelsFormat {
	case "json_array":
		var jsonModels []ModelInfo
		if json.Unmarshal([]byte(raw), &jsonModels) == nil {
			return jsonModels, nil
		}
	case "providers_obj":
		if models := parseProvidersJSON(raw); len(models) > 0 {
			return models, nil
		}
	case "providers_arr":
		if models := parseProvidersArrayJSON(raw); len(models) > 0 {
			return models, nil
		}
	case "flat":
		if models := parseModelLines(raw); len(models) > 0 {
			return models, nil
		}
	}

	// Fallback: try all formats in order
	if models, _ := tryAllFormats(raw); len(models) > 0 {
		return models, nil
	}
	return nil, nil
}

// parseProvidersJSON parseia formato A: {"provider": {"openai": {"models": {"gpt-4o": {}}}}}
func parseProvidersJSON(raw string) []ModelInfo {
	var data struct {
		Provider map[string]struct {
			Models map[string]any `json:"models"`
		} `json:"provider"`
	}
	if err := json.Unmarshal([]byte(raw), &data); err != nil || len(data.Provider) == 0 {
		return nil
	}
	return flattenProviderMap(data.Provider)
}

type rawModel struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerID"`
	Name       string `json:"name"`
}

type rawProvider struct {
	Name   string              `json:"name"`
	Models map[string]rawModel `json:"models"`
}

type rawProvidersResponse struct {
	Providers []rawProvider `json:"providers"`
}

// parseProvidersArrayJSON parseia formato B:
//
//	{"providers": [{"name":"openai","models":{"gpt-4o":{...}}}], "default":{...}}
func parseProvidersArrayJSON(raw string) []ModelInfo {
	var data rawProvidersResponse
	if err := json.Unmarshal([]byte(raw), &data); err != nil || len(data.Providers) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var models []ModelInfo
	for _, p := range data.Providers {
		for _, m := range p.Models {
			id := m.ID
			if id == "" {
				continue
			}
			if seen[id] {
				continue
			}
			seen[id] = true
			modelName := m.Name
			if modelName == "" {
				modelName = id
			}
			providerName := m.ProviderID
			if providerName == "" {
				providerName = p.Name
			}
			models = append(models, ModelInfo{
				ID:           id,
				Name:         modelName,
				ProviderName: providerName,
			})
		}
	}
	return models
}

func flattenProviderMap(provider map[string]struct {
	Models map[string]any `json:"models"`
}) []ModelInfo {
	seen := make(map[string]bool)
	var models []ModelInfo
	for pName, pData := range provider {
		for mName := range pData.Models {
			prov, model, id := resolveModelID(pName, mName)
			if seen[id] {
				continue
			}
			seen[id] = true
			models = append(models, ModelInfo{
				ID:           id,
				Name:         model,
				ProviderName: prov,
			})
		}
	}
	return models
}

// resolveModelID legacy fallback — usado apenas pelo formato A (providers_obj).
func resolveModelID(parentProvider, modelKey string) (provider, model, id string) {
	if strings.Contains(modelKey, "/") {
		parts := strings.SplitN(modelKey, "/", 2)
		return parts[0], parts[1], modelKey
	}
	return parentProvider, modelKey, parentProvider + "/" + modelKey
}

// tryAllFormats tenta todos os formatos em ordem.
func tryAllFormats(raw string) ([]ModelInfo, bool) {
	if models := parseProvidersJSON(raw); len(models) > 0 {
		return models, true
	}
	if models := parseProvidersArrayJSON(raw); len(models) > 0 {
		return models, true
	}
	if models := parseModelLines(raw); len(models) > 0 {
		return models, true
	}
	return nil, false
}

// parseModelLines parseia formato linha-por-linha: provider/model
func parseModelLines(raw string) []ModelInfo {
	lines := strings.Split(raw, "\n")
	seen := make(map[string]bool)
	var models []ModelInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true

		provider, model := parseModelLine(line)
		if model == "" {
			continue
		}
		models = append(models, ModelInfo{
			ID:           line,
			Name:         model,
			ProviderName: provider,
		})
	}
	return models
}

func parseModelLine(line string) (provider, model string) {
	if idx := strings.LastIndex(line, "/"); idx > 0 && idx < len(line)-1 {
		return line[:idx], line[idx+1:]
	}
	return "", line
}

// ExecuteChatBlocking is the non-streaming variant.
func ExecuteChatBlocking(req Request) (string, error) {
	return ExecuteChat(req, nil)
}

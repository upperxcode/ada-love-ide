package urlworker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"ada-love-ide/internal/reasoning"
	"time"
)

// OpenCodeSessionManager tracks opencode server sessions per Ada session.
type OpenCodeSessionManager struct {
	mu       sync.Mutex
	sessions map[string]string // adaSessionID -> opencodeSessionID
}

var DefaultOpenCodeManager = &OpenCodeSessionManager{
	sessions: make(map[string]string),
}

func (m *OpenCodeSessionManager) Get(adaSessionID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.sessions[adaSessionID]
	return id, ok
}

func (m *OpenCodeSessionManager) Set(adaSessionID, opencodeSessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[adaSessionID] = opencodeSessionID
}

func (m *OpenCodeSessionManager) Delete(adaSessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, adaSessionID)
}

// opencodeServerResponse represents the response from POST /session/:id/message
type opencodeMessageResponse struct {
	Info  opencodeMessageInfo   `json:"info"`
	Parts []opencodeMessagePart `json:"parts"`
}

type opencodeMessageInfo struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Role    string `json:"role"`
}

type opencodeMessagePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// opencodeCreateSessionResponse represents the response from POST /session
type opencodeCreateSessionResponse struct {
	ID string `json:"id"`
}

func (r *Runtime) OpenCodeCreateSession() (string, error) {
	url := fmt.Sprintf("%s/session", r.BaseURL)
	body := `{}`

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	setAuthHeaders(req, r.Config.Environment)

	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: nil}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	fmt.Printf("[OpenCodeCreateSession] HTTP %d, body: %q\n", resp.StatusCode, string(raw))

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("server error %d: %s", resp.StatusCode, string(raw))
	}

	var result opencodeCreateSessionResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parse session response (HTTP %d): %w body=%q", resp.StatusCode, err, string(raw))
	}
	if result.ID == "" {
		return "", fmt.Errorf("empty session ID (HTTP %d) body=%q", resp.StatusCode, string(raw))
	}
	fmt.Printf("[OpenCodeCreateSession] session created: %s\n", result.ID)
	return result.ID, nil
}

func (r *Runtime) OpenCodeSendMessage(opencodeSessionID, message, model string, emit EmitterFn, system string, contextParts ...string) (string, string, error) {
	modelID, providerID := resolveModelProvider(r.BaseURL, model)
	payload := map[string]any{
		"model": map[string]string{
			"id":         modelID,
			"providerID": providerID,
			"modelID":    modelID,
		},
	}
	if system != "" {
		payload["system"] = system
	}
	var parts []map[string]string
	for _, cp := range contextParts {
		if cp != "" {
			parts = append(parts, map[string]string{"type": "text", "text": cp})
		}
	}
	parts = append(parts, map[string]string{"type": "text", "text": message})
	payload["parts"] = parts
	bodyBytes, _ := json.Marshal(payload)

	// 1. Connect to event stream first
	eventCtx, eventCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer eventCancel()

	eventReq, _ := http.NewRequestWithContext(eventCtx, "GET", r.BaseURL+"/event", nil)
	setAuthHeaders(eventReq, r.Config.Environment)
	eventClient := &http.Client{Timeout: 0}
	eventResp, err := eventClient.Do(eventReq)
	if err != nil {
		return "", "", fmt.Errorf("event stream: %w", err)
	}
	defer eventResp.Body.Close()

	fmt.Printf("[OpenCodeSendMessage] streaming via prompt_async, session=%s\n", opencodeSessionID)

	// 2. Send message asynchronously
	asyncURL := fmt.Sprintf("%s/session/%s/prompt_async", r.BaseURL, opencodeSessionID)
	asyncReq, _ := http.NewRequest("POST", asyncURL, bytes.NewReader(bodyBytes))
	asyncReq.Header.Set("Content-Type", "application/json")
	setAuthHeaders(asyncReq, r.Config.Environment)
	asyncClient := &http.Client{Timeout: 10 * time.Second}
	asyncResp, err := asyncClient.Do(asyncReq)
	if err != nil {
		return "", "", fmt.Errorf("prompt_async: %w", err)
	}
	asyncResp.Body.Close()

	partTypes := make(map[string]string) // partID → type
	var textBuf strings.Builder          // accumulator for text part
	var thinkingBuf strings.Builder      // accumulator for thinking/ reasoning text
	reasoningParser := reasoning.NewParser()
	currentReasoningType := reasoning.Plan

	// 3. Read SSE events, streaming as they arrive
	scanner := bufio.NewScanner(eventResp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var evt struct {
			Type       string `json:"type"`
			Properties struct {
				SessionID string          `json:"sessionID"`
				Part      json.RawMessage `json:"part"`
				PartID    string          `json:"partID"`
				Field     string          `json:"field"`
				Delta     string          `json:"delta"`
				Info      json.RawMessage `json:"info"`
			} `json:"properties"`
		}
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		if evt.Properties.SessionID != opencodeSessionID {
			continue
		}

		fmt.Printf("[OpenCodeSSE] type=%s\n", evt.Type)

		switch evt.Type {
		case "message.part.updated":
			var part struct {
				ID   string `json:"id"`
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(evt.Properties.Part, &part); err != nil {
				continue
			}
			partTypes[part.ID] = part.Type
			fmt.Printf("[OC] updated id=%s type=%s text_len=%d\n", part.ID, part.Type, len(part.Text))
			// Only store the type — text comes from deltas to avoid duplication
		case "message.part.delta":
			if evt.Properties.Field != "text" || evt.Properties.Delta == "" {
				continue
			}
			partType, known := partTypes[evt.Properties.PartID]
			deltaPreview := evt.Properties.Delta[:min(len(evt.Properties.Delta), 40)]
			if !known {
				fmt.Printf("[OC] delta UNKNOWN partID=%s delta=%q\n", evt.Properties.PartID, deltaPreview)
				partTypes[evt.Properties.PartID] = "pending"
				continue
			}
			if partType == "pending" || partType == "" {
				fmt.Printf("[OC] delta PENDING partID=%s delta=%q\n", evt.Properties.PartID, deltaPreview)
				continue
			}
			fmt.Printf("[OC] delta type=%s partID=%s delta=%q\n", partType, evt.Properties.PartID, deltaPreview)
			if partType == "reasoning" {
				thinkingBuf.WriteString(evt.Properties.Delta)
				detected, _ := reasoningParser.Feed(evt.Properties.Delta)
				currentReasoningType = detected
				if emit != nil { emit("chat:thinking", map[string]any{"content": evt.Properties.Delta, "type": string(currentReasoningType)}) }
			} else if partType == "text" {
				textBuf.WriteString(evt.Properties.Delta)
				fullText := textBuf.String()
				if emit != nil { emit("chat:delta", map[string]any{"content": fullText}) }
			}

		case "message.updated":
			var info struct {
				Finish string `json:"finish"`
				Text   string `json:"text"`
			}
			if evt.Properties.Info != nil {
				json.Unmarshal(evt.Properties.Info, &info)
			}
			if info.Finish == "stop" || info.Finish == "length" || info.Finish == "error" {
				goto done
			}

		case "session.idle":
			goto done
		case "session.status":
			var status struct {
				Type string `json:"type"`
			}
			if evt.Properties.Info != nil {
				json.Unmarshal(evt.Properties.Info, &status)
			}
			if status.Type == "idle" {
				goto done
			}
		}
	}
done:
	finalText := strings.TrimSpace(textBuf.String())
	finalThinking := strings.TrimSpace(thinkingBuf.String())
	fmt.Printf("[OpenCodeSendMessage] streaming done, output len=%d, thinking len=%d\n", len(finalText), len(finalThinking))
	return finalText, finalThinking, nil
}

func (r *Runtime) OpenCodeListModels() ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/config/providers", r.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("list models request: %w", err)
	}
	setAuthHeaders(req, r.Config.Environment)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseProvidersArrayJSON(string(body)), nil
}

// resolveModelProvider returns the modelID and providerID for a model string.
// If the model is "provider/model", splits and uses directly.
// If the model is just "name", fetches the provider list to find the providerID.
func resolveModelProvider(baseURL, model string) (modelID, providerID string) {
	if strings.Contains(model, "/") {
		// Already in provider/model format
		parts := strings.SplitN(model, "/", 2)
		return parts[1], parts[0]
	}

	// Try fetching providers to find the matching model
	models, err := fetchAllModels(baseURL)
	if err != nil || len(models) == 0 {
		return model, model // fallback: use model as both
	}
	for _, m := range models {
		if m.ID == model || m.Name == model {
			return m.ID, m.ProviderName
		}
	}
	return model, model
}

// fetchAllModels fetches and parses the full model list from /config/providers.
func fetchAllModels(baseURL string) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/config/providers", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseProvidersArrayJSON(string(body)), nil
}

// splitModel splits "providerID/modelName" into (providerID, modelName).
func splitModel(s string) (providerID, modelName string) {
	if idx := strings.LastIndex(s, "/"); idx > 0 && idx < len(s)-1 {
		return s[:idx], s[idx+1:]
	}
	return "", s
}

func setAuthHeaders(req *http.Request, envJSON string) {
	if envJSON == "" || envJSON == "{}" {
		return
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(envJSON), &env); err != nil {
		return
	}
	for k, v := range env {
		req.Header.Set(k, v)
	}
}

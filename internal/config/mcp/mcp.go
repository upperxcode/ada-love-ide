package mcp

type HeaderEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MCPServerUI struct {
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	URL           string            `json:"url"`
	Enabled       bool              `json:"enabled"`
	Icon          string            `json:"icon"`
	Color         string            `json:"color"`
	Timeout       int               `json:"timeout"`
	OAuthClientID string            `json:"oauth_client_id"`
	Headers       []HeaderEntry     `json:"headers"`
}

type ConnectionDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type ConnectionTestResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMS int    `json:"latency_ms"`
}

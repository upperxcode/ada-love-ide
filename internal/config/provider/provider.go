package provider

type ModelSettings struct {
	ContextSize int     `json:"context_size,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	Type        string  `json:"type,omitempty"`
	Vision      bool    `json:"vision,omitempty"`
	Embedding   bool    `json:"embedding,omitempty"`
	Tools       bool    `json:"tools,omitempty"`
	Free        bool    `json:"free,omitempty"`
	Thinking    bool    `json:"thinking,omitempty"`
}

type ProviderAPIKey struct {
	Key     string `json:"key"`
	UserKey string `json:"user_key"`
}

	type ProviderConfig struct {
		Icon           string                   `json:"icon"`
		Color          string                   `json:"color"`
		APIURL         string                   `json:"api_url"`
		APIKeys        []ProviderAPIKey         `json:"api_keys"`
		TypeConnection string                   `json:"type_connection"`
		Strategy       string                   `json:"strategy"`
		Models         map[string]ModelSettings `json:"models"`
	}

func New(name string) ProviderConfig {
	return ProviderConfig{
		Icon:           "🔌",
		Color:          "#3b82f6",
		APIKeys:        []ProviderAPIKey{},
		Models:         map[string]ModelSettings{},
		TypeConnection: "openai",
	}
}

type ProviderModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Vision      bool   `json:"vision"`
	Embedding   bool   `json:"embedding"`
	Tools       bool   `json:"tools"`
	Free        bool   `json:"free"`
	Thinking    bool   `json:"thinking"`
	ContextSize int    `json:"context_size,omitempty"`
}

type ProviderTestResult struct {
	OK      bool   `json:"ok"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

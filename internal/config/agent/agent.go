package agent

type AgentConfig struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	Type          string  `json:"type"`
	Icon          string  `json:"icon"`
	Color         string  `json:"color"`
	MaxIterations int     `json:"max_iterations"`
	Temperature   float64 `json:"temperature"`
	SystemPrompt  string  `json:"system_prompt"`
}

func New(name string) AgentConfig {
	return AgentConfig{
		Name: name,
		Type: "executor",
		Icon: "🤖",
	}
}

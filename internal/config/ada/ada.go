package ada

import (
	"ada-love-ide/internal/config/agent"
	"ada-love-ide/internal/config/mcp"
	"ada-love-ide/internal/config/provider"
	"ada-love-ide/internal/config/worker"
	"ada-love-ide/internal/config/workspace"
)

type AdaConfig struct {
	ActiveWorkspacePath  string                             `json:"active_workspace_path"`
	ActiveWorkspaceIndex int                                `json:"active_workspace_index"`
	Workspaces           []workspace.WorkspaceConfig        `json:"workspaces"`
	TinyBrainProvider    string                             `json:"tiny_brain_provider"`
	TinyBrainModel       string                             `json:"tiny_brain_model"`
	TinyBrainTools       []string                           `json:"tiny_brain_tools"`
	Workers              []worker.WorkerConfig              `json:"workers"`
	WorkerCategories     []string                           `json:"worker_categories"`
	Agents               []agent.AgentConfig                `json:"agents"`
	AgentCategories      []string                           `json:"agent_categories"`
	ProviderKeys         map[string]string                  `json:"provider_keys"`
	ProviderBases        map[string]string                  `json:"provider_bases"`
	ModelSettings        map[string]any                     `json:"model_settings"`
	ModelList            []any                              `json:"model_list"`
	Providers            map[string]provider.ProviderConfig `json:"providers"`
	EmbeddingModel       string                             `json:"embedding_model"`
	EmbeddingProvider    string                             `json:"embedding_provider"`
	ImageModel           string                             `json:"image_model"`
	ImageProvider        string                             `json:"image_provider"`
	SpecModel            string                             `json:"spec_model"`
	SpecProvider         string                             `json:"spec_provider"`
	SpecTools            []string                           `json:"spec_tools"`
	MCPServers           map[string]mcp.MCPServerUI         `json:"mcp_servers"`
	ActiveSessionID      string                             `json:"active_session_id"`
	SidebarVisible       bool                               `json:"sidebar_visible"`
}

func New() AdaConfig {
	return AdaConfig{
		Workspaces:       []workspace.WorkspaceConfig{},
		WorkerCategories: []string{},
		AgentCategories:  []string{},
		ProviderKeys:     map[string]string{},
		ProviderBases:    map[string]string{},
		ModelSettings:    map[string]any{},
		ModelList:        []any{},
		Providers:        map[string]provider.ProviderConfig{},
		SpecTools:        []string{},
		MCPServers:       map[string]mcp.MCPServerUI{},
		SidebarVisible:   true,
	}
}

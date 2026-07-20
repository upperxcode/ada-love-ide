package main

import (
	"context"
	"fmt"

	"ada-love-ide/internal/config/ada"
)

// GetAdaConfig devolve o config completo do app.
func (a *App) GetAdaConfig() ada.AdaConfig {
	cfg := ada.New()
	cfg.ActiveWorkspacePath = a.eng.DB.ActiveWorkspace()
	cfg.Workspaces = a.eng.DB.ListWorkspaces()
	cfg.Workers = a.eng.DB.ListWorkers()
	cfg.Agents = a.eng.DB.ListAgents()
	cfg.WorkerCategories = a.eng.DB.WorkerCategories()
	cfg.AgentCategories = a.eng.DB.AgentCategories()
	cfg.Providers = a.eng.DB.ListProviders()
	cfg.MCPServers = a.eng.DB.ListMCPServers()
	cfg.ActiveSessionID = a.eng.DB.ActiveSessionID()
	cfg.SidebarVisible = a.eng.DB.SidebarVisible()
	cfg.ActiveWorkspaceIndex = 0
	cfg.ModelList = []any{}

	// Carrega fixed models (tinybrain, spec, etc)
	cfg.EmbeddingProvider, cfg.EmbeddingModel, _ = a.eng.DB.GetFixedModel("embedding")
	cfg.ImageProvider, cfg.ImageModel, _ = a.eng.DB.GetFixedModel("image")
	cfg.SpecProvider, cfg.SpecModel, cfg.SpecTools = a.eng.DB.GetFixedModel("spec")
	cfg.TinyBrainProvider, cfg.TinyBrainModel, cfg.TinyBrainTools = a.eng.DB.GetFixedModel("tinybrain")

	fmt.Printf("[Backend] GetAdaConfig: %d MCP servers found\n", len(cfg.MCPServers))
	for name := range cfg.MCPServers {
		fmt.Printf("  - MCP: %s\n", name)
	}

	return cfg
}

// SetAdaConfig salva apenas o estado global do app (active session, sidebar).
// NÃO deve ser usado para salvar listas complexas (workers, agents, workspaces)
// para evitar race conditions e deleções acidentais no startup.
func (a *App) SetAdaConfig(cfg ada.AdaConfig) {
	if cfg.ActiveSessionID != "" {
		a.eng.DB.SetActiveSessionID(cfg.ActiveSessionID)
	}
	a.eng.DB.SetSidebarVisible(cfg.SidebarVisible)

	// Persiste MCP servers apenas se houver mudança manual (substitui o mapa completo)
	// No startup, o frontend envia o que recebeu do backend, então isso é neutro.
	if cfg.MCPServers != nil {
		a.eng.DB.ReplaceMCPServers(cfg.MCPServers)
	}

	// Fixed models são leves e seguros de persistir
	a.eng.DB.SaveFixedModel("embedding", cfg.EmbeddingProvider, cfg.EmbeddingModel, nil)
	a.eng.DB.SaveFixedModel("image", cfg.ImageProvider, cfg.ImageModel, nil)
	a.eng.DB.SaveFixedModel("spec", cfg.SpecProvider, cfg.SpecModel, cfg.SpecTools)
	a.eng.DB.SaveFixedModel("tinybrain", cfg.TinyBrainProvider, cfg.TinyBrainModel, cfg.TinyBrainTools)
}

var _ = context.Background

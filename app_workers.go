package main

import (
	"ada-love-ide/internal/config/agent"
	"ada-love-ide/internal/config/worker"
)

// GetWorkers retorna todos os workers.
func (a *App) GetWorkers() []worker.WorkerConfig {
	return a.eng.DB.ListWorkers()
}

// SetWorkers sobrescreve todos os workers.
func (a *App) SetWorkers(list []worker.WorkerConfig) {
	a.eng.DB.SetWorkers(list)
}

// GetWorkerCategories retorna categorias de worker.
func (a *App) GetWorkerCategories() []string {
	return a.eng.DB.WorkerCategories()
}

// SetWorkerCategories define categorias de worker.
func (a *App) SetWorkerCategories(c []string) {
	a.eng.DB.SetWorkerCategories(c)
}

// GetAgents retorna todos os agentes.
func (a *App) GetAgents() []agent.AgentConfig {
	return a.eng.DB.ListAgents()
}

// SetAgents sobrescreve todos os agentes.
func (a *App) SetAgents(list []agent.AgentConfig) {
	a.eng.DB.SetAgents(list)
}

// GetAgentCategories retorna categorias de agente.
func (a *App) GetAgentCategories() []string {
	return a.eng.DB.AgentCategories()
}

// SetAgentCategories define categorias de agente.
func (a *App) SetAgentCategories(c []string) {
	a.eng.DB.SetAgentCategories(c)
}

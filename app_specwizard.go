package main

import (
	"context"

	"ada-love-ide/internal/config/specwizard"
	"ada-love-ide/internal/specwizardmgr"
)

// GetSpecWizards lista todos os spec-wizards.
func (a *App) GetSpecWizards() []specwizard.SpecWizardConfig {
	return a.eng.SpecWizardMgr.List()
}

// GetSpecWizard recupera um spec-wizard por ID.
func (a *App) GetSpecWizard(id string) (*specwizard.SpecWizardConfig, error) {
	return a.eng.SpecWizardMgr.Get(id)
}

// SaveSpecWizard cria/atualiza um spec-wizard.
func (a *App) SaveSpecWizard(w specwizard.SpecWizardConfig) {
	a.eng.SpecWizardMgr.Save(w)
}

// DeleteSpecWizard remove um spec-wizard.
func (a *App) DeleteSpecWizard(id string) {
	a.eng.SpecWizardMgr.Delete(id)
}

// GetExperts retorna a lista de plugins de expert disponiveis.
func (a *App) GetExperts() []map[string]any {
	return a.eng.SpecWizardMgr.GetExperts()
}

// GetPatterns busca arquiteturas do expert para a linguagem.
func (a *App) GetPatterns(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetPatterns(lang)
}

// GetArchitectures retorna arquiteturas de todos os experts.
func (a *App) GetArchitectures() []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetArchitectures()
}

// GetStacks busca stacks do expert para a linguagem.
func (a *App) GetStacks(lang string) []map[string]any {
	return a.eng.SpecWizardMgr.GetStacks(lang)
}

// GetStateManagement busca opções de gerenciamento de estado do expert.
func (a *App) GetStateManagement(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetStateManagement(lang)
}

// GetPersistenceOptions retorna estratégias de persistência do expert.
func (a *App) GetPersistenceOptions(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetPersistenceOptions(lang)
}

// GetEngineeringPhilosophies retorna filosofias de engenharia do expert.
func (a *App) GetEngineeringPhilosophies(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetEngineeringPhilosophies(lang)
}

// GetDesignPatterns retorna design patterns (GoF) do expert.
func (a *App) GetDesignPatterns(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetDesignPatterns(lang)
}

// GetDataPatterns retorna data patterns do expert.
func (a *App) GetDataPatterns(lang string) []specwizardmgr.Option {
	return a.eng.SpecWizardMgr.GetDataPatterns(lang)
}

// ComputeHealth calcula a saúde da arquitetura com base nas seleções.
func (a *App) ComputeHealth(cfg specwizard.SpecWizardConfig) int {
	return a.eng.SpecWizardMgr.ComputeHealth(cfg)
}

// GetRecommendations gera cards de recomendação da arquitetura.
func (a *App) GetRecommendations(cfg specwizard.SpecWizardConfig) []specwizardmgr.Recommendation {
	return a.eng.SpecWizardMgr.GetRecommendations(cfg)
}

// InferField usa o modelo "spec" para inferir/sugerir o valor de um campo
// do spec wizard com base nas configurações atuais.
func (a *App) InferField(fieldName string, cfg specwizard.SpecWizardConfig) (string, error) {
	return a.eng.SpecWizardMgr.InferField(context.Background(), fieldName, cfg)
}

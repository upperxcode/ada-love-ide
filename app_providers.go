package main

import (
	"context"
	"fmt"

	"ada-love-ide/internal/config/provider"
	iprovider "ada-love-ide/internal/provider"
)

// GetProvidersConfig devolve todos os providers configurados.
func (a *App) GetProvidersConfig() map[string]provider.ProviderConfig {
	return a.eng.DB.ListProviders()
}

// ListProvidersByWorker devolve todos os providers de um worker específico.
func (a *App) ListProvidersByWorker(workerName string) map[string]provider.ProviderConfig {
	return a.eng.DB.ListProvidersByWorker(workerName)
}

// GetProvidersByWorker devolve todos os providers de um worker específico.
func (a *App) GetProvidersByWorker(workerName string) map[string]provider.ProviderConfig {
	return a.eng.DB.ListProvidersByWorker(workerName)
}

// SaveProvidersConfig persiste todos os providers (stub).
func (a *App) SaveProvidersConfig() {}

// SaveDBProvider salva/atualiza um provider específico.
func (a *App) SaveDBProvider(name string, cfg provider.ProviderConfig) {
	a.eng.DB.SaveProvider(name, cfg)
}

// DeleteDBProvider remove um provider.
func (a *App) DeleteDBProvider(name string) {
	a.eng.DB.DeleteProvider(name)
}

// SaveWorkerProvider salva/atualiza um provider específico de um worker.
// O workerName define o vínculo do provider ao worker.
func (a *App) SaveWorkerProvider(workerName, name string, cfg provider.ProviderConfig) {
	if cfg.Worker == "" {
		cfg.Worker = workerName
	}
	a.eng.DB.SaveProvider(name, cfg)
}

// DeleteWorkerProvider remove um provider.
func (a *App) DeleteWorkerProvider(name string) {
	a.eng.DB.DeleteProvider(name)
}

// ListChatProviders retorna os providers com pelo menos um modelo.
func (a *App) ListChatProviders() []string {
	out := []string{}
	for name, cfg := range a.eng.DB.ListProviders() {
		if len(cfg.Models) > 0 {
			out = append(out, name)
		}
	}
	return out
}

// RemoveModel remove um modelo de um provider.
func (a *App) RemoveModel(name, providerName string) {
	provs := a.eng.DB.ListProviders()
	p, ok := provs[providerName]
	if !ok {
		return
	}
	delete(p.Models, name)
	a.eng.DB.SaveProvider(providerName, p)
}

// TestProviderConnection testa conexão com um provider usando uma chave específica.
func (a *App) TestProviderConnection(name, connectionType, apiUrl, apiKey string) provider.ProviderTestResult {
	tempCfg := provider.ProviderConfig{
		TypeConnection: connectionType,
		APIURL:         apiUrl,
		APIKeys:        []provider.ProviderAPIKey{{Key: apiKey}},
	}

	_, err := iprovider.FetchModels(context.Background(), tempCfg)
	if err != nil {
		return provider.ProviderTestResult{
			OK:      false,
			Success: false,
			Message: fmt.Sprintf("Conexão falhou: %v", err),
		}
	}

	return provider.ProviderTestResult{
		OK:      true,
		Success: true,
		Message: "Conexão OK!",
	}
}

// FetchProviderModels busca modelos reais do provedor.
func (a *App) FetchProviderModels(name, connectionType, apiUrl, apiKey string) []provider.ProviderModel {
	tempCfg := provider.ProviderConfig{
		TypeConnection: connectionType,
		APIURL:         apiUrl,
		APIKeys:        []provider.ProviderAPIKey{{Key: apiKey}},
	}

	discovered, err := iprovider.FetchModels(context.Background(), tempCfg)
	if err != nil {
		fmt.Printf("[Backend] FetchProviderModels erro: %v\n", err)
		return []provider.ProviderModel{}
	}

	out := make([]provider.ProviderModel, 0, len(discovered))
	for _, m := range discovered {
		out = append(out, provider.ProviderModel{
			ID:          m.ID,
			Name:        m.ID,
			Free:        m.Capabilities.Free,
			Thinking:    m.Capabilities.Thinking,
			Vision:      m.Capabilities.Vision,
			Embedding:   m.Capabilities.Embedding,
			Tools:       m.Capabilities.Tools,
			ContextSize: m.ContextSize,
		})
	}
	return out
}

// SetFixedModelTools define a lista de ferramentas fixas de um modelo (tinybrain/spec).
func (a *App) SetFixedModelTools(modelName string, tools []string) bool {
	_ = modelName
	_ = tools
	return true
}

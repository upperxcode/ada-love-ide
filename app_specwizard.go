package main

import (
	"fmt"
	"time"

	"ada-love-ide/internal/config/specwizard"
)

// GetSpecWizards lista todos os spec-wizards.
func (a *App) GetSpecWizards() []specwizard.SpecWizardConfig {
	return a.eng.DB.ListWizards()
}

// GetSpecWizard recupera um spec-wizard por ID.
func (a *App) GetSpecWizard(id string) (*specwizard.SpecWizardConfig, error) {
	w, ok := a.eng.DB.GetWizard(id)
	if !ok {
		return nil, nil
	}
	return &w, nil
}

// SaveSpecWizard cria/atualiza um spec-wizard.
func (a *App) SaveSpecWizard(w specwizard.SpecWizardConfig) {
	if w.ID == "" {
		w.ID = time.Now().Format("20060102150405")
	}
	now := time.Now()
	w.UpdatedAt = now
	a.eng.DB.PutWizard(w)
}

// DeleteSpecWizard remove um spec-wizard.
func (a *App) DeleteSpecWizard(id string) { a.eng.DB.DeleteWizard(id) }

// GetExperts retorna a lista de plugins de expert disponiveis.
func (a *App) GetExperts() []map[string]any {
	plugins := a.eng.Plugins.List()
	result := make([]map[string]any, 0, len(plugins))
	for _, p := range plugins {
		result = append(result, map[string]any{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"language":    p.Language,
		})
	}
	return result
}

// GetPatterns busca padroes do expert para a linguagem especificada.
func (a *App) GetPatterns(lang string) []map[string]any {
	plugin, ok := a.eng.Plugins.FindByLanguage(lang)
	if !ok {
		return []map[string]any{}
	}

	if err := a.eng.Plugins.EnsureRunning(plugin); err != nil {
		fmt.Printf("[SpecWizard] Failed to start expert %s: %v\n", plugin.ID, err)
		return []map[string]any{}
	}

	resp, err := a.eng.Plugins.CallExpert(plugin, "options")
	if err != nil {
		fmt.Printf("[SpecWizard] Failed to call expert %s: %v\n", plugin.ID, err)
		return []map[string]any{}
	}

	patterns := make([]map[string]any, 0)
	if archs, ok := resp["architectures"].([]interface{}); ok {
		for _, a := range archs {
			if m, ok := a.(map[string]interface{}); ok {
				patterns = append(patterns, map[string]any{
					"id":          m["id"],
					"name":        m["name"],
					"description": m["description"],
					"category":    "architecture",
					"group":       "architectures",
					"scope":       "project",
				})
			}
		}
	}
	return patterns
}

// GetArchitectures retorna arquiteturas de todos os experts.
func (a *App) GetArchitectures() []map[string]any {
	seen := make(map[string]bool)
	var result []map[string]any

	for _, plugin := range a.eng.Plugins.List() {
		if err := a.eng.Plugins.EnsureRunning(plugin); err != nil {
			continue
		}

		resp, err := a.eng.Plugins.CallExpert(plugin, "options")
		if err != nil {
			continue
		}

		if archs, ok := resp["architectures"].([]interface{}); ok {
			for _, a := range archs {
				if m, ok := a.(map[string]interface{}); ok {
					id, _ := m["id"].(string)
					if id != "" && !seen[id] {
						seen[id] = true
						result = append(result, map[string]any{
							"id":          m["id"],
							"name":        m["name"],
							"description": m["description"],
						})
					}
				}
			}
		}
	}
	return result
}

// GetStacks busca stacks do expert para a linguagem especificada.
func (a *App) GetStacks(lang string) []map[string]any {
	plugin, ok := a.eng.Plugins.FindByLanguage(lang)
	if !ok {
		return []map[string]any{}
	}

	if err := a.eng.Plugins.EnsureRunning(plugin); err != nil {
		fmt.Printf("[SpecWizard] Failed to start expert %s: %v\n", plugin.ID, err)
		return []map[string]any{}
	}

	resp, err := a.eng.Plugins.CallExpert(plugin, "options")
	if err != nil {
		fmt.Printf("[SpecWizard] Failed to call expert %s: %v\n", plugin.ID, err)
		return []map[string]any{}
	}

	stacks := make([]map[string]any, 0)
	if templates, ok := resp["stack_templates"].([]interface{}); ok {
		for _, t := range templates {
			if m, ok := t.(map[string]interface{}); ok {
				stacks = append(stacks, map[string]any{
					"id":        m["id"],
					"name":      m["name"],
					"libraries": m["libraries"],
				})
			}
		}
	}
	return stacks
}

// GetStateManagement busca opções de gerenciamento de estado do expert.
func (a *App) GetStateManagement(lang string) []map[string]any {
	plugin, ok := a.eng.Plugins.FindByLanguage(lang)
	if !ok {
		return []map[string]any{}
	}

	if err := a.eng.Plugins.EnsureRunning(plugin); err != nil {
		return []map[string]any{}
	}

	resp, err := a.eng.Plugins.CallExpert(plugin, "options")
	if err != nil {
		return []map[string]any{}
	}

	options := make([]map[string]any, 0)
	if states, ok := resp["state_management"].([]interface{}); ok {
		for _, s := range states {
			if m, ok := s.(map[string]interface{}); ok {
				options = append(options, map[string]any{
					"id":          m["id"],
					"name":        m["name"],
					"description": m["description"],
				})
			}
		}
	}
	return options
}

// SuggestFieldValue usa um mock para sugerir valor de campo.
func (a *App) SuggestFieldValue(fieldName, context, currentValue string) string {
	if currentValue != "" {
		return currentValue
	}
	return "mock-suggestion-" + fieldName
}

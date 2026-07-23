package main

import "fmt"

// GetFixedModels returns all fixed models using per-name lookup.
func (a *App) GetFixedModels() ([]map[string]any, error) {
	fmt.Println("[FixedModel] GetFixedModels called (per-name)")
	result := []map[string]any{}
	knownNames := []string{"Classifier", "embedding", "image", "spec", "tinybrain"}
	for _, name := range knownNames {
		provider, model, tools := a.eng.DB.GetFixedModel(name)
		fmt.Printf("[FixedModel] " + "%s -> provider=%s model=%s tools=%v\n", name, provider, model, tools)
		result = append(result, map[string]any{
			"name":     name,
			"provider": provider,
			"model":    model,
			"tools":    tools,
		})
	}
	fmt.Printf("[FixedModel] Returning %d items\n", len(result))
	return result, nil
}

// SaveFixedModel saves or updates a fixed model.
func (a *App) SaveFixedModel(name, provider, model string, tools []string) {
	a.eng.DB.SaveFixedModel(name, provider, model, tools)
	fmt.Printf("[FixedModel] Saved: " + "%s (provider=%s, model=%s, tools=%v)\n", name, provider, model, tools)
}

// DeleteFixedModel deletes a fixed model by name.
func (a *App) DeleteFixedModel(name string) {
	a.eng.DB.DeleteFixedModel(name)
	fmt.Printf("[FixedModel] Deleted: " + "%s\n", name)
}

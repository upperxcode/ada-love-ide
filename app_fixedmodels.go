package main

import "fmt"

// GetFixedModels retorna todos os modelos fixos.
func (a *App) GetFixedModels() []map[string]any {
	fmt.Println("[FixedModel] GetFixedModels called")
	result := a.eng.DB.ListFixedModels()
	if result == nil {
		// Try to read directly from DB as fallback
		fmt.Println("[FixedModel] ListFixedModels returned nil, trying direct query...")
		db := a.eng.DB.DB()
		rows, err := db.Query("SELECT id, name, provider, model FROM fixed_models")
		if err != nil {
			fmt.Printf("[FixedModel] Direct query error: %v\n", err)
			return nil
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var name, provider, model string
			if err := rows.Scan(&id, &name, &provider, &model); err != nil {
				fmt.Printf("[FixedModel] Direct scan error: %v\n", err)
				continue
			}
			fmt.Printf("[FixedModel] Direct hit: id=%d name=%s provider=%s model=%s\n", id, name, provider, model)
		}
		return nil
	}
	fmt.Printf("[FixedModel] GetFixedModels returning %d items\n", len(result))
	return result
}

// SaveFixedModel salva ou atualiza um modelo fixo.
func (a *App) SaveFixedModel(name, provider, model string, tools []string) {
	a.eng.DB.SaveFixedModel(name, provider, model, tools)
	_ = model
	_ = tools
	fmt.Printf("[FixedModel] Saved: %s (provider=%s, model=%s, tools=%v)\n", name, provider, model, tools)
}

// DeleteFixedModel remove um modelo fixo pelo nome.
func (a *App) DeleteFixedModel(name string) {
	a.eng.DB.DeleteFixedModel(name)
	fmt.Printf("[FixedModel] Deleted: %s\n", name)
}


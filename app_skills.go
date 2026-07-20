package main

import (
	"context"
	"fmt"
	"path/filepath"

	"ada-love-ide/internal/config/skill"
)

// SearchSkills busca skills em registries remotos e locais.
func (a *App) SearchSkills(query string) []skill.SearchResult {
	// Busca em registries remotos
	results, err := a.eng.SkillReg.SearchAll(context.Background(), query, 20)
	if err != nil {
		fmt.Printf("[Backend] SearchSkills: erro nos registries: %v\n", err)
	}
	if results == nil {
		results = []skill.SearchResult{}
	}
	fmt.Printf("[Backend] SearchSkills: query=%q, %d resultados dos registries\n", query, len(results))
	return results
}

// InstallSkill baixa e instala um skill de um registry remoto.
func (a *App) InstallSkill(registryName, slug, version string) error {
	reg := a.eng.SkillReg.GetRegistry(registryName)
	if reg == nil {
		return fmt.Errorf("registry %q não encontrado", registryName)
	}

	targetDir := filepath.Join(a.eng.Skills.Dir(), slug)
	fmt.Printf("[Backend] InstallSkill: registry=%s, slug=%s, version=%s, target=%s\n",
		registryName, slug, version, targetDir)

	if err := reg.DownloadAndInstall(context.Background(), slug, version, targetDir); err != nil {
		return fmt.Errorf("falha ao instalar skill de %s: %w", registryName, err)
	}
	return nil
}

// GetInstalledSkills lista nomes dos skills instalados.
func (a *App) GetInstalledSkills() []string {
	result := a.eng.Skills.ListInstalled()
	fmt.Printf("[Backend] Skills instaladas: %v\n", result)
	return result
}

// GetSkillFullInfo retorna metadados completos de um skill instalado.
func (a *App) GetSkillFullInfo(name string) (*skill.SkillFullInfo, error) {
	info, err := a.eng.Skills.GetInfo(name)
	if err != nil {
		fmt.Printf("[Backend] Erro ao buscar info de skill '%s': %v\n", name, err)
		return nil, err
	}
	fmt.Printf("[Backend] Info de skill '%s': name=%s, description=%s, tags=%v\n", name, info.Name, info.Description, info.Tags)
	return info, nil
}

// UninstallSkill remove um skill do filesystem.
func (a *App) UninstallSkill(name string) error {
	return a.eng.Skills.Uninstall(name)
}

// GetSkillDetails retorna o conteúdo markdown de um skill.
func (a *App) GetSkillDetails(name string) string {
	info, err := a.eng.Skills.GetInfo(name)
	if err != nil {
		return "# " + name + "\n\nSkill não encontrado."
	}
	return info.Markdown
}

// GetSkills retorna todos os skills do banco.
func (a *App) GetSkills() []skill.SkillConfig {
	return a.eng.DB.ListSkills()
}

// UpdateSkillConfig atualiza os metadados de uma skill no banco.
func (a *App) UpdateSkillConfig(cfg skill.SkillConfig) error {
	// 1. Salva o conteúdo markdown no filesystem (mantém compatibilidade)
	if err := a.eng.Skills.SaveCustom(cfg.Name, cfg.Description, cfg.Tags, cfg.Content); err != nil {
		return err
	}
	// 2. Salva metadados (active, color, icon) no banco de dados
	return a.eng.DB.PutSkill(cfg)
}

// SaveCustomSkill grava uma skill customizada.
func (a *App) SaveCustomSkill(name, description, tagsCSV, content string) error {
	return a.eng.Skills.SaveCustom(name, description, tagsCSV, content)
}

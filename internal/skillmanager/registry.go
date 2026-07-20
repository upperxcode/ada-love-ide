// Package skillmanager gerencia skills no filesystem e em registries remotos.
package skillmanager

import (
	"context"
	"sort"
	"sync"

	"ada-love-ide/internal/config/skill"
)

// SkillRegistry é a interface que um registry remoto de skills deve implementar.
type SkillRegistry interface {
	// Name retorna o identificador único do registry (ex: "clawhub", "github").
	Name() string

	// Search busca skills no registry que correspondam à query.
	Search(ctx context.Context, query string, limit int) ([]skill.SearchResult, error)

	// DownloadAndInstall baixa e instala uma skill do registry no targetDir.
	DownloadAndInstall(ctx context.Context, slug, version, targetDir string) error
}

// RegistryManager coordena múltiplos registries de skills.
type RegistryManager struct {
	registries    []SkillRegistry
	maxConcurrent int
	mu            sync.RWMutex
}

// NewRegistryManager cria um RegistryManager vazio com limite de 2 buscas concorrentes.
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		maxConcurrent: 2,
	}
}

// AddRegistry adiciona um registry ao manager.
func (rm *RegistryManager) AddRegistry(r SkillRegistry) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.registries = append(rm.registries, r)
}

// GetRegistry busca um registry pelo nome. Retorna nil se não encontrado.
func (rm *RegistryManager) GetRegistry(name string) SkillRegistry {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for _, r := range rm.registries {
		if r.Name() == name {
			return r
		}
	}
	return nil
}

// SearchAll busca em todos os registries concorrentemente, mescla os resultados
// ordenados por score decrescente e limita ao limite solicitado.
func (rm *RegistryManager) SearchAll(ctx context.Context, query string, limit int) ([]skill.SearchResult, error) {
	rm.mu.RLock()
	regs := make([]SkillRegistry, len(rm.registries))
	copy(regs, rm.registries)
	rm.mu.RUnlock()

	if len(regs) == 0 {
		return nil, nil
	}

	type regResult struct {
		results []skill.SearchResult
		err     error
	}

	// Semáforo para limitar concorrência
	sem := make(chan struct{}, rm.maxConcurrent)
	resultsCh := make(chan regResult, len(regs))

	var wg sync.WaitGroup
	for _, reg := range regs {
		wg.Add(1)
		go func(r SkillRegistry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results, err := r.Search(ctx, query, limit)
			resultsCh <- regResult{results: results, err: err}
		}(reg)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var all []skill.SearchResult
	var lastErr error
	for res := range resultsCh {
		if res.err != nil {
			lastErr = res.err
			continue
		}
		all = append(all, res.results...)
	}

	if len(all) == 0 && lastErr != nil {
		return nil, lastErr
	}

	// Ordena por score decrescente
	sort.Slice(all, func(i, j int) bool {
		return all[i].Score > all[j].Score
	})

	if len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}

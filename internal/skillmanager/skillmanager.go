// Package skillmanager gerencia skills no filesystem.
// Skills são diretórios em SkillsDir contendo um SKILL.md com
// YAML frontmatter (name, description, version, tags).
package skillmanager

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"ada-love-ide/internal/config/skill"
)

// Manager opera sobre um diretório raiz de skills.
type Manager struct {
	dir string
}

// New cria um Manager apontando para dir (ex: ~/.opencode/skills).
func New(dir string) *Manager { return &Manager{dir: dir} }

// Dir retorna o diretório raiz das skills.
func (m *Manager) Dir() string { return m.dir }

// ListInstalled lista nomes de skills instalados (diretórios com SKILL.md).
func (m *Manager) ListInstalled() []string {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(m.dir, e.Name(), "SKILL.md")); err == nil {
			out = append(out, e.Name())
		}
	}
	return out
}

// GetInfo lê metadados de um skill instalado.
func (m *Manager) GetInfo(name string) (*skill.SkillFullInfo, error) {
	p := filepath.Join(m.dir, name, "SKILL.md")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	meta := parseFrontmatter(string(data))
	info := &skill.SkillFullInfo{
		Name:        or(meta["name"], name),
		Description: or(meta["description"], ""),
		Version:     or(meta["version"], "0.0.0"),
		Registry:    or(meta["registry"], "local"),
		Markdown:    string(data),
		Tags:        splitTags(meta["tags"]),
		LineCount:   lines(string(data)),
		CharCount:   len(data),
	}
	return info, nil
}

// Install cria um skill a partir de conteúdo markdown.
func (m *Manager) Install(name, content string) error {
	dir := filepath.Join(m.dir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
}

// Uninstall remove um skill (diretório inteiro).
func (m *Manager) Uninstall(name string) error {
	return os.RemoveAll(filepath.Join(m.dir, name))
}

// Search filtra skills instalados por query (case-insensitive em nome + conteúdo).
func (m *Manager) Search(query string) []skill.SearchResult {
	q := strings.ToLower(query)
	var out []skill.SearchResult
	for _, name := range m.ListInstalled() {
		info, err := m.GetInfo(name)
		if err != nil {
			continue
		}
		// Se query vazia, aceita todas as skills
		// Se query não vazia, procura no nome, descrição ou conteúdo
		matches := q == "" ||
			strings.Contains(strings.ToLower(info.Name), q) ||
			strings.Contains(strings.ToLower(info.Description), q) ||
			strings.Contains(strings.ToLower(info.Markdown), q)
		if !matches {
			continue
		}
		out = append(out, skill.SearchResult{
			Name:        info.Name,
			DisplayName: info.Name,
			Summary:     info.Description,
			Description: info.Description,
			Version:     info.Version,
			Score:       1.0,
		})
	}
	return out
}

// SaveCustom grava uma skill customizada com frontmatter completo.
func (m *Manager) SaveCustom(name, description, tagsCSV, content string) error {
	// Constrói o SKILL.md com frontmatter + corpo markdown
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	if description != "" {
		b.WriteString("description: " + description + "\n")
	}
	if tagsCSV != "" {
		b.WriteString("tags: " + tagsCSV + "\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(content)
	return m.Install(name, b.String())
}

// ── helpers ────────────────────────────────────────────────────

func parseFrontmatter(md string) map[string]string {
	meta := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(md))
	inFront := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if !inFront {
				inFront = true
				continue
			}
			break
		}
		if !inFront {
			break
		}
		if i := strings.IndexByte(line, ':'); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			meta[key] = strings.Trim(val, "\"'")
		}
	}
	return meta
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func lines(s string) int {
	return strings.Count(s, "\n") + 1
}

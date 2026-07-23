package summary

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ── Manifest detection ──────────────────────────────────────────

var manifestFiles = []struct {
	File     string // filename to look for in workspace root
	Language string // detected language/ecosystem
	Parser   func(path string) (string, error)
}{
	{File: "go.mod", Language: "Go", Parser: parseGoMod},
	{File: "package.json", Language: "Node.js / TypeScript", Parser: parsePackageJSON},
	{File: "Cargo.toml", Language: "Rust", Parser: parseCargoToml},
	{File: "pyproject.toml", Language: "Python", Parser: parsePyprojectToml},
	{File: "requirements.txt", Language: "Python", Parser: parseRequirementsTxt},
	{File: "Gemfile", Language: "Ruby", Parser: parseGemfile},
	{File: "composer.json", Language: "PHP", Parser: parseComposerJSON},
	{File: "build.gradle", Language: "Java / Kotlin (Gradle)", Parser: parseBuildGradle},
	{File: "pom.xml", Language: "Java (Maven)", Parser: parsePomXML},
}

// GenerateFromWorkspace lê as fontes do workspace e monta um resumo textual.
// Retorna uma string de 1-3 parágrafos (~200-500 tokens) e um hash SHA256
// das fontes para detectar mudanças futuras.
func GenerateFromWorkspace(workspaceDir string) (summary string, sourceHash string, err error) {
	if workspaceDir == "" {
		return "", "", fmt.Errorf("workspace directory is empty")
	}

	// 1. Detecta linguagem e lê manifesto
	lang, manifestInfo := detectManifest(workspaceDir)

	// 2. Lê AGENT.md / .AGENT.md
	agentContent := readAgentFile(workspaceDir)

	// 3. Lê README.md
	readmeContent := readReadme(workspaceDir)

	// 4. Estrutura de diretórios (2 níveis)
	dirTree := buildDirectoryTree(workspaceDir, 2)

	// 5. Computa o hash das fontes (antes de qualquer processamento)
	sourceHash = hashSources(workspaceDir, agentContent, readmeContent, dirTree)

	// Monta o resumo final
	summary = buildSummary(lang, manifestInfo, agentContent, readmeContent, dirTree)

	return summary, sourceHash, nil
}

// hashSources computa um hash SHA256 de todas as fontes do workspace
// para detectar alterações que exijam regeneração do summary.
func hashSources(workspaceDir, agentContent, readmeContent, dirTree string) string {
	h := sha256.New()

	// Inclui o nome do diretório (mudar de projeto = novo hash)
	h.Write([]byte(workspaceDir + "\n"))

	// Inclui conteúdo do manifesto (se existir)
	for _, mf := range manifestFiles {
		path := filepath.Join(workspaceDir, mf.File)
		if data, err := os.ReadFile(path); err == nil {
			h.Write([]byte(mf.File + ":"))
			h.Write(data)
			h.Write([]byte("\n"))
		}
	}

	// Inclui conteúdo do AGENT.md
	if agentContent != "" {
		h.Write([]byte("AGENT.md:"))
		h.Write([]byte(agentContent))
		h.Write([]byte("\n"))
	}

	// Inclui conteúdo do README.md
	if readmeContent != "" {
		h.Write([]byte("README.md:"))
		h.Write([]byte(readmeContent))
		h.Write([]byte("\n"))
	}

	// Inclui estrutura de diretórios
	if dirTree != "" {
		h.Write([]byte("dirs:"))
		h.Write([]byte(dirTree))
		h.Write([]byte("\n"))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// detectManifest procura o primeiro arquivo de manifesto na raiz e extrai informações.
func detectManifest(workspaceDir string) (language string, info string) {
	for _, mf := range manifestFiles {
		path := filepath.Join(workspaceDir, mf.File)
		if _, err := os.Stat(path); err == nil {
			info, err := mf.Parser(path)
			if err != nil {
				return mf.Language, ""
			}
			return mf.Language, info
		}
	}
	return "", ""
}

// readAgentFile procura AGENT.md, .AGENT.md, .AGENTS.md, etc.
func readAgentFile(dir string) string {
	candidates := []string{
		"AGENT.md",
		".AGENT.md",
		".AGENTS.md",
		".agent.md",
		".agents.md",
		"AGENTS.md",
	}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return ""
}

// readReadme lê o README.md do workspace.
func readReadme(dir string) string {
	candidates := []string{"README.md", "Readme.md", "readme.md"}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return ""
}

// extractAgentRules extrai as regras mais relevantes do AGENT.md,
// ignorando frontmatter YAML, blocos de código, e metadados de configuração.
func extractAgentRules(content string) string {
	lines := strings.Split(content, "\n")
	var relevant []string
	inCodeBlock := false
	inFrontMatter := false
	frontMatterCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detecta frontmatter YAML (delimitado por ---)
		if trimmed == "---" {
			frontMatterCount++
			if frontMatterCount <= 2 {
				inFrontMatter = !inFrontMatter
				continue
			}
		}
		if inFrontMatter {
			continue
		}

		// Pula blocos de código
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Ignora linhas vazias
		if trimmed == "" {
			continue
		}

		// Captura apenas headings (##, ###, ####) e list items (- , *)
		// que contenham conteúdo relevante (não-metadados)
		if strings.HasPrefix(trimmed, "##") || strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "####") {
			// Pula headings de metadados
			heading := strings.TrimLeft(trimmed, "# ")
			skipHeadings := []string{"name", "description", "tools", "model", "maxturns", "skills", "mcpservers"}
			isMeta := false
			for _, skip := range skipHeadings {
				if strings.EqualFold(heading, skip) {
					isMeta = true
					break
				}
			}
			if !isMeta {
				relevant = append(relevant, trimmed)
			}
			continue
		}

		// List items que parecem regras (não confundir com campos YAML)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			// Pula se parece campo YAML (ex: "- name: ..." ou "- description: ...")
			if isYAMLFieldLine(trimmed) {
				continue
			}
			relevant = append(relevant, trimmed)
			continue
		}

		// Linhas com numeração (1., 2., etc.) - regras numeradas
		if len(trimmed) > 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
			relevant = append(relevant, trimmed)
			continue
		}

		// Linhas de destaque (bold, itálico) ou definições curtas
		if strings.Contains(trimmed, "**") || strings.Contains(trimmed, "__") {
			relevant = append(relevant, trimmed)
			continue
		}
	}

	// Limita a quantidade e remove duplicatas aproximadas
	seen := make(map[string]bool)
	var unique []string
	for _, r := range relevant {
		normalized := strings.ToLower(strings.TrimSpace(r))
		if !seen[normalized] {
			seen[normalized] = true
			unique = append(unique, r)
		}
	}
	if len(unique) > 20 {
		unique = unique[:20]
	}

	return strings.Join(unique, "\n")
}

// isYAMLFieldLine verifica se uma linha parece um campo YAML (ex: "- name: ...")
func isYAMLFieldLine(line string) bool {
	// Remove marcador de lista
	content := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
	// Se tem ":" e a parte antes do ":" parece um nome de campo YAML
	if idx := strings.Index(content, ":"); idx > 0 {
		fieldName := strings.TrimSpace(content[:idx])
		// Campos YAML típicos
		yamlFields := map[string]bool{
			"name": true, "description": true, "tools": true, "model": true,
			"maxturns": true, "skills": true, "mcpservers": true, "icon": true,
			"color": true, "id": true, "type": true, "version": true,
			"language": true, "framework": true, "engine": true,
		}
		if yamlFields[strings.ToLower(fieldName)] {
			return true
		}
	}
	return false
}

// extractReadmeSummary extrai a primeira sentença ou parágrafo do README.
func extractReadmeSummary(content string) string {
	// Pega o primeiro parágrafo não vazio (ignora front matter)
	lines := strings.Split(content, "\n")
	var paragraph strings.Builder
	inFrontMatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			inFrontMatter = !inFrontMatter
			continue
		}
		if inFrontMatter {
			continue
		}
		if trimmed == "" && paragraph.Len() > 0 {
			break
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if paragraph.Len() > 0 {
				paragraph.WriteString(" ")
			}
			paragraph.WriteString(trimmed)
		}
		if paragraph.Len() > 500 {
			break
		}
	}
	result := paragraph.String()
	if len(result) > 500 {
		result = result[:500]
	}
	return result
}

// extractFrontmatterDesc extrai o campo "description" do frontmatter YAML do AGENT.md.
// Suporta tanto formato padrão (---\\nfield:\\n---) quanto compacto (---field: value\\n---).
func extractFrontmatterDesc(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return ""
	}

	inFrontMatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Se a linha começa com "---" e tem mais conteúdo, é frontmatter compacto
		if strings.HasPrefix(trimmed, "---") {
			if !inFrontMatter {
				inFrontMatter = true
				// Pode ser "---name: ..." - processa o resto da linha
				rest := strings.TrimPrefix(trimmed, "---")
				if rest != "" {
					if extracted := extractYAMLField(rest, "description"); extracted != "" {
						return extracted
					}
				}
				continue
			}
			break // Fim do frontmatter
		}

		if inFrontMatter {
			if extracted := extractYAMLField(trimmed, "description"); extracted != "" {
				return extracted
			}
		}
	}
	return ""
}

// extractYAMLField extrai o valor de um campo YAML simples (formato "field: value")
func extractYAMLField(line, field string) string {
	prefix := field + ":"
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	val = strings.Trim(val, `"'`)
	return val
}

// buildDirectoryTree constrói a árvore de diretórios até depth níveis.
func buildDirectoryTree(dir string, maxDepth int) string {
	var sb strings.Builder
	buildTreeRec(dir, "", 0, maxDepth, &sb)
	return sb.String()
}

func buildTreeRec(dir string, prefix string, depth int, maxDepth int, sb *strings.Builder) {
	if depth > maxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Filter out hidden dirs and common noise
	var dirs []os.DirEntry
	var files []os.DirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if name == "node_modules" || name == "target" || name == "vendor" || name == ".git" {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	// Show dirs first
	for i, d := range dirs {
		isLast := i == len(dirs)-1 && len(files) == 0
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		sb.WriteString(fmt.Sprintf("%s%s%s/\n", prefix, connector, d.Name()))
		subPrefix := prefix + "│   "
		if isLast {
			subPrefix = prefix + "    "
		}
		buildTreeRec(filepath.Join(dir, d.Name()), subPrefix, depth+1, maxDepth, sb)
	}

	// Then files
	for i, f := range files {
		isLast := i == len(files)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, f.Name()))
	}
}

// buildSummary monta o resumo final em texto plano estruturado, 1-3 parágrafos.
// Formato: nome + descrição, stack, arquitetura/regras, estrutura.
func buildSummary(language, manifestInfo, agentContent, readmeContent, dirTree string) string {
	var sb strings.Builder

	// 1. Nome do projeto — extraído do manifesto ou do título do README
	projectName := extractProjectName(language, manifestInfo, readmeContent)
	if projectName != "" {
		sb.WriteString(projectName)
	}

	// 2. Descrição — primeira frase significativa do README, ou do frontmatter do AGENT.md
	projectDesc := extractReadmeSummary(readmeContent)
	if projectDesc == "" {
		projectDesc = extractFrontmatterDesc(agentContent)
	}
	if projectDesc != "" {
		if sb.Len() > 0 {
			sb.WriteString(" — ")
		}
		// Limpa: remove quebras, trunca
		oneLine := strings.ReplaceAll(projectDesc, "\n", " ")
		if len(oneLine) > 200 {
			oneLine = oneLine[:200] + "..."
		}
		sb.WriteString(oneLine)
	}
	if sb.Len() > 0 {
		sb.WriteString("\n\n")
	}

	// 3. Stack
	if language != "" {
		sb.WriteString("Stack: ")
		sb.WriteString(language)
		if manifestInfo != "" {
			sb.WriteString(" — ")
			sb.WriteString(manifestInfo)
		}
		sb.WriteString("\n")
	}

	// 4. Arquitetura e regras — extraídas do AGENT.md (limpas)
	rules := extractAgentRules(agentContent)
	if rules != "" {
		sb.WriteString("Architecture / Conventions:\n")
		sb.WriteString(rules)
		sb.WriteString("\n")
	}

	// 5. Estrutura de diretórios (2 níveis, compacta)
	if dirTree != "" {
		sb.WriteString("Directory structure:\n")
		lines := strings.Split(dirTree, "\n")
		// Mostra apenas top-level + alguns deep
		limit := 20
		if len(lines) < limit {
			limit = len(lines)
		}
		for _, line := range lines[:limit] {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		if len(lines) > limit {
			sb.WriteString(fmt.Sprintf("  … and %d more entries\n", len(lines)-limit))
		}
		sb.WriteString("\n")
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		result = "No project metadata found."
	}
	return result
}

// extractProjectName tenta extrair o nome do projeto das fontes disponíveis.
func extractProjectName(language, manifestInfo, readmeContent string) string {
	// Tenta extrair do manifesto (ex: "Module: github.com/user/project, Dependencies: 5")
	if manifestInfo != "" {
		// Formato: "Module: <module>, Dependencies: <N>"
		if strings.HasPrefix(manifestInfo, "Module: ") {
			rest := strings.TrimPrefix(manifestInfo, "Module: ")
			if idx := strings.Index(rest, ","); idx >= 0 {
				rest = rest[:idx]
			}
			// Pega o último segmento do path do módulo
			rest = strings.TrimSpace(rest)
			if idx := strings.LastIndex(rest, "/"); idx >= 0 {
				return rest[idx+1:]
			}
			return rest
		}
		// Formato: "Package: <name>, ..."
		if strings.HasPrefix(manifestInfo, "Package: ") {
			rest := strings.TrimPrefix(manifestInfo, "Package: ")
			if idx := strings.Index(rest, ","); idx >= 0 {
				rest = rest[:idx]
			}
			return strings.Trim(rest, `"'`)
		}
		// Formato: "Crate: <name>"
		if strings.HasPrefix(manifestInfo, "Crate: ") {
			return strings.TrimPrefix(manifestInfo, "Crate: ")
		}
	}

	// Tenta extrair do README (primeiro heading h1)
	if readmeContent != "" {
		for _, line := range strings.Split(readmeContent, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# ") {
				name := strings.TrimPrefix(trimmed, "# ")
				name = strings.TrimSpace(name)
				if name != "" && !strings.HasPrefix(name, "[") {
					return name
				}
			}
		}
	}

	return ""
}

// ── Manifest parsers ────────────────────────────────────────────

func parseGoMod(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	module := extractModuleName(content)
	deps := countGoDeps(content)
	if module != "" {
		return fmt.Sprintf("Module: %s, Dependencies: %d", module, deps), nil
	}
	return fmt.Sprintf("Dependencies: %d", deps), nil
}

func parsePackageJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	name := extractJSONField(content, "name")
	deps := countJSONDeps(content)
	if name != "" {
		return fmt.Sprintf("Package: %s, Dependencies: %d", name, deps), nil
	}
	return fmt.Sprintf("Dependencies: %d", deps), nil
}

func parseCargoToml(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	name := extractTomlField(content, "name")
	if name != "" {
		return fmt.Sprintf("Crate: %s", name), nil
	}
	return "Rust project", nil
}

func parsePyprojectToml(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	name := extractTomlField(content, "name")
	if name != "" {
		return fmt.Sprintf("Package: %s", name), nil
	}
	return "Python project", nil
}

func parseRequirementsTxt(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}
	return fmt.Sprintf("Dependencies: %d", count), nil
}

func parseGemfile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gem ") {
			count++
		}
	}
	return fmt.Sprintf("Gems: %d", count), nil
}

func parseComposerJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	name := extractJSONField(content, "name")
	deps := countJSONDeps(content)
	if name != "" {
		return fmt.Sprintf("Package: %s, Dependencies: %d", name, deps), nil
	}
	return fmt.Sprintf("Dependencies: %d", deps), nil
}

func parseBuildGradle(path string) (string, error) {
	return "Gradle project", nil
}

func parsePomXML(path string) (string, error) {
	return "Maven project", nil
}

// ── Helpers ─────────────────────────────────────────────────────

func extractModuleName(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func countGoDeps(content string) int {
	count := 0
	inRequireBlock := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Detecta bloco require ()
		if trimmed == "require (" {
			inRequireBlock = true
			continue
		}
		if trimmed == ")" {
			inRequireBlock = false
			continue
		}

		// Dentro de um bloco require, linhas indentadas são dependências
		if inRequireBlock && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			count++
			continue
		}

		// Fora do bloco, linhas que começam com \t (tab) e não são comentários
		if !inRequireBlock && strings.HasPrefix(line, "\t") && !strings.HasPrefix(trimmed, "//") {
			// Pula "replace" dentro de bloco replace () ou linhas avulsas
			if !strings.HasPrefix(trimmed, "replace ") {
				count++
			}
		}
	}
	return count
}

func extractJSONField(content, field string) string {
	// Simple extraction without full JSON parse
	marker := fmt.Sprintf(`"%s"`, field)
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(marker):]
	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return ""
	}
	rest = strings.TrimSpace(rest[colonIdx+1:])
	if strings.HasPrefix(rest, `"`) {
		rest = rest[1:]
		endIdx := strings.Index(rest, `"`)
		if endIdx >= 0 {
			return rest[:endIdx]
		}
	}
	return ""
}

func countJSONDeps(content string) int {
	count := 0
	// Count "dependencies" and "devDependencies" keys
	for _, key := range []string{`"dependencies"`, `"devDependencies"`} {
		idx := strings.Index(content, key)
		if idx >= 0 {
			// Count entries in the object
			rest := content[idx:]
			braceIdx := strings.Index(rest, "{")
			if braceIdx >= 0 {
				objContent := rest[braceIdx:]
				count += strings.Count(objContent, `"`) / 2 // rough estimate
			}
		}
	}
	return count
}

func extractTomlField(content, field string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, field+" ") || strings.HasPrefix(line, field+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"'`)
				return val
			}
		}
	}
	return ""
}

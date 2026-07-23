# Workspace Summary (workspaces.summary)

> Resumo textual do projeto/workspace que substitui o envio de fontes brutas
> (AGENT.md, estrutura de diretórios, dependências) como contexto para o LLM.

## Contrato

### O que é

Um campo `summary` no `WorkspaceConfig` que contém um resumo textual do projeto
em 1–3 parágrafos (~200–500 tokens). Quando presente, o `contextprovider` envia
apenas o summary como contexto — **não** envia mais:

- AGENT.md (conteúdo bruto)
- Estrutura de diretórios (árvore)
- Manifesto do projeto (go.mod, package.json, etc.)

### O que NÃO substitui

- ❌ Histórico da conversa (fica com `max_prompt_send`)
- ❌ Símbolos de código (fica com o `ada-code-indexer`)
- ❌ Conhecimento dinâmico (fica com `inherit_knowledge`)

## Estrutura do summary

```
<nome do projeto> — <descrição de uma linha>

Stack: <linguagem> — <módulo/package>, Dependencies: <N>
Architecture / Conventions:
<princípios extraídos do AGENT.md>
Directory structure:
├── <árvore de diretórios (2 níveis)>
```

### Campos

| Seção | Fonte | Extrai |
|-------|-------|--------|
| Nome | Manifesto (`go.mod`, `package.json`, etc.) | Module/Package name |
| Descrição | `README.md` (1º parágrafo) ou AGENT.md (frontmatter `description:`) | Primeira frase significativa |
| Stack | Manifesto detectado | Linguagem + info do manifesto |
| Dependências | Manifesto | Contagem de dependências |
| Regras | AGENT.md | Headings + list items (ignora frontmatter YAML) |
| Estrutura | Diretório do workspace | Tree de 2 níveis (ignora `.git`, `node_modules`, `vendor`, `target`) |

## Fontes que alimentam

| Fonte | Extrai | Quando muda |
|-------|--------|-------------|
| `{workspace}/AGENT.md` | Regras de desenvolvimento, stack, arquitetura | Edição manual |
| `go.mod` / `package.json` / `Cargo.toml` / etc. | Nome do módulo, linguagem, dependências | `git pull`, nova dep |
| Estrutura de diretórios (2 níveis) | Organização do projeto | Criação/deleção de pastas |
| `README.md` (se existir) | Descrição do projeto, propósito | Edição |

### Detecção de linguagem (ordem de procura)

| Arquivo | Linguagem |
|---------|-----------|
| `go.mod` | Go |
| `package.json` | Node.js / TypeScript |
| `Cargo.toml` | Rust |
| `pyproject.toml` / `requirements.txt` | Python |
| `Gemfile` | Ruby |
| `pom.xml` / `build.gradle` | Java |
| `composer.json` | PHP |
| `*.csproj` | C# |

O scanner procura o **primeiro** arquivo de manifesto na raiz do workspace
e determina a stack.

## Ciclo de vida

```
1. Cria/atualiza workspace
       │ summary = "" (vazio)
       ▼
2. Novo chat criado (sidebar)
       │
       ├─► CreateSession / CreateSessionWithConfig
       │     └─► go ensureWorkspaceSummary(path)
       │           ├─► Computa hash atual das fontes
       │           ├─► Compara com SummaryHash salvo
       │           └─► Se diferente OU sem summary → regenera em background
       │
3. Usuário clica "Generate" (Settings > Workspaces)
       │
       └─► GenerateWorkspaceSummary(path)
             └─► summary.GenerateFromWorkspace(wsDir)
                   ├─► Lê fontes (manifest, AGENT.md, README.md, dirs)
                   ├─► Computa hash SHA256 das fontes
                   ├─► Monta resumo textual
                   └─► Salva summary + hash no DB
```

## Content Hash

Um hash SHA256 é computado sobre todas as fontes para detectar alterações:

```
SHA256(
  workspaceDir + "\n"
  + manifestFile.name + ":" + manifestFile.content + "\n"
  + "AGENT.md:" + agentContent + "\n"
  + "README.md:" + readmeContent + "\n"
  + "dirs:" + dirTree + "\n"
)
```

Quando um novo chat é criado:
- Computa hash atual
- Compara com `SummaryHash` salvo no DB
- Se diferentes → regenera o summary em background (goroutine não-bloqueante)

## Onde o summary entra no contexto

No `Engine.ExtraContext()`, ao montar as camadas de contexto:

```
Se ws.Summary != "":
  → "=== WORKSPACE SUMMARY ===" + ws.Summary
  (NÃO envia AGENT.md, estrutura, manifesto)
Senão:
  → "=== ARCHITECTURAL GOLDEN RULES ===" + AGENT.md (comportamento legado)
```

No `Engine.buildBreakdown()` (debug de uso de contexto):

```
Se ws.Summary != "":
  → "Workspace Summary" (cor: #10b981)
Senão:
  → "Golden rules" (cor: #6366f1)
```

## Arquivos implementados

### Backend (Go)

| Arquivo | Função |
|---------|--------|
| `internal/config/workspace/workspace.go` | Campo `Summary` + `SummaryHash` no `WorkspaceConfig` |
| `internal/db/sessions.go` | Mapeamento no `upsertWorkspace()` e `adaptWorkspaceToInternal()` |
| `internal/summary/generator.go` | Gerador + hash (746 linhas) |
| `internal/engine/engine.go` | `ExtraContext()` e `buildBreakdown()` usam summary |
| `app_workspaces.go` | `ensureWorkspaceSummary()`, `EnsureWorkspaceSummary()`, `GenerateWorkspaceSummary()` |
| `app_sessions.go` | Auto-trigger em `CreateSession()` e `CreateSessionWithConfig()` |

### Storage module (`ada-storage-module`)

| Arquivo | Função |
|---------|--------|
| `storage/workspace_store.go` | Campo `SummaryHash` no struct + SQLs |
| `storage/migrations_workspaces.go` | Migration `workspaceAddSummaryHash` |
| `storage/migrations.go` | Migration v45 registrada |

### Frontend (Svelte)

| Arquivo | Função |
|---------|--------|
| `frontend/src/lib/stores/entities.svelte.ts` | `summary` + `summary_hash` na interface `WorkspaceConfig` |
| `frontend/src/lib/components/settings/WorkspaceDialog.svelte` | Seção Summary: textarea + botão Generate |
| `frontend/src/lib/components/icon/icon-map.ts` | Ícone `file-text` adicionado |

## API (Wails bindings)

```go
// Gera/resumo e salva no DB. Retorna o texto gerado.
func (a *App) GenerateWorkspaceSummary(path string) (string, error)

// Verifica e atualiza se necessário (não-bloqueante).
func (a *App) EnsureWorkspaceSummary(path string) error

// Auto-trigger interno (não exposto via Wails):
func (a *App) ensureWorkspaceSummary(workspacePath string)
```

## Detalhes da implementação

### Generator (`internal/summary/generator.go`)

O pacote `summary` contém:

- **`GenerateFromWorkspace(workspaceDir)`** → `(summary, hash, error)`
  - Detecta linguagem pelo manifesto
  - Lê AGENT.md (com fallback para `.AGENT.md`, `.AGENTS.md`, etc.)
  - Lê README.md
  - Escaneia diretórios (2 níveis, ignorando `.git`, `node_modules`, `target`, `vendor`)
  - Extrai nome do projeto (do manifesto ou README h1)
  - Extrai descrição (README 1º parágrafo ou frontmatter `description:` do AGENT.md)
  - Extrai regras (headings + list items, pulando frontmatter YAML)
  - Computa hash SHA256
  - Monta texto estruturado

- **`hashSources(workspaceDir, agentContent, readmeContent, dirTree)`** → `string`
  - SHA256 de: dir + manifestos + AGENT.md + README.md + estrutura

- **Manifest parsers**: `parseGoMod`, `parsePackageJSON`, `parseCargoToml`, `parsePyprojectToml`, `parseRequirementsTxt`, `parseGemfile`, `parseComposerJSON`, `parseBuildGradle`, `parsePomXML`

### Regras de extração do AGENT.md

A função `extractAgentRules()`:
1. Ignora frontmatter YAML (`---` ... `---`)
2. Ignora blocos de código (```...```)
3. Ignora campos de metadados (`name:`, `description:`, `tools:`, `model:`, etc.)
4. Captura headings (`##`, `###`, `####`)
5. Captura list items (`-`, `*`) que não sejam campos YAML
6. Captura linhas numeradas (`1.`, `2.`, etc.)
7. Captura linhas com destaque (`**bold**`)
8. Deduplica por similaridade
9. Limita a 20 regras

## Testes

Teste interno via Go test no pacote `internal/summary/`:

```go
func TestGenerateOurselves(t *testing.T) {
    s, h, err := summary.GenerateFromWorkspace("/path/to/workspace")
    // Verifica conteúdo e hash
}
```

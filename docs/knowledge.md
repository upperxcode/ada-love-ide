# Semantic Knowledge Search

## Visão Geral

O **Semantic Knowledge Search** permite que o LLM recupere os knowledge items
mais relevantes de um workspace usando **similaridade de cosseno** entre vetores
de embedding, em vez de simplesmente despejar todo o conteúdo no contexto.

Isso reduz o desperdício de tokens e melhora a precisão das respostas, pois
apenas os itens semanticamente próximos à consulta do usuário são injetados.

```
Pacote: internal/knowledge/
Arquivos: embedder.go, store.go, index.go, index_test.go
Testes:   go test ./internal/knowledge/ -v  (13 testes)
```

---

## Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│                    ExtraContext                          │
│  (engine.go:348)                                        │
│  ┌──────────────────────────────────────────────────┐   │
│  │  knowledgeIndex.Search(ctx, query, wsID, 5)      │   │
│  │  ├── OK → top-5 por cosine similarity            │   │
│  │  └── FALHA → AllTexts() → fallback completo      │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              knowledge.Index                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │  entries map[int64][]Entry                       │   │
│  │  embedder  Embedder                              │   │
│  │  mu         sync.RWMutex                         │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              knowledge.Embedder                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  POST {baseURL}/embeddings                       │   │
│  │  Body: {"model": ..., "input": text}             │   │
│  │  Res:   {data: [{embedding: [float]}]}          │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### Fluxo de dados

```
workspace_knowledge (SQLite)
        │
        ▼  store.GetWorkspace(path).Knowledge
knowledge.IndexWorkspace(ctx, wsID, items)
        │
        ├── embedder.Embed(item) para cada item → []float32
        ├── armazena em entries[workspaceID]
        └── substitui entries anteriores
                │
                ▼  ExtraContext
knowledge.Search(ctx, query, wsID, topK=5)
        │
        ├── embedder.Embed(query) → []float32
        ├── cosineSimilarity(queryVec, entry.Vector)
        ├── sort por similaridade (decrescente)
        └── retorna top-K textos
                │
                ▼
        LLM recebe: "=== KNOWLEDGE (relevant) ===\n
                     item A\n---\nitem C\n---\nitem B"
```

---

## Pacote `internal/knowledge/`

### `embedder.go` — Embedder

```go
type Embedder struct {
    provider string  // ex: "openai"
    model    string  // ex: "text-embedding-3-small"
    baseURL  string  // ex: "https://api.openai.com/v1"
    apiKey   string
    client   *http.Client
}

func NewEmbedder(provider, model, baseURL, apiKey string) Embedder
func (e Embedder) Embed(ctx context.Context, text string) ([]float32, error)
```

- Chama `POST {baseURL}/embeddings` com `{"model": e.model, "input": text}`
- Converte `[]float64` da resposta para `[]float32`
- Usa `net/http` padrão — **sem dependências externas**
- O `baseURL` já deve incluir o prefixo `/v1` (ex: `https://api.openai.com/v1`)

### `store.go` — Entry

```go
type Entry struct {
    KnowledgeID int64      // índice 0-based dentro do workspace
    WorkspaceID int64      // workspace a que pertence
    Text        string     // texto original do knowledge item
    Vector      []float32  // vetor de embedding
}
```

### `index.go` — Index (core)

```go
type Index struct {
    embedder Embedder
    mu       sync.RWMutex
    entries  map[int64][]Entry    // workspaceID → entries
}

func NewIndex(embedder Embedder) *Index
func (idx *Index) IndexWorkspace(ctx context.Context, workspaceID int64, items []string)
func (idx *Index) Search(ctx context.Context, query string, workspaceID int64, topK int) []string
func (idx *Index) AllTexts(workspaceID int64) []string
func (idx *Index) Count(workspaceID int64) int
```

#### `IndexWorkspace`

1. Itera sobre `items`
2. Para cada item: `embedder.Embed(ctx, item)`
3. Se falhar → `fmt.Printf("[Knowledge] WARNING: ...")` + **pula o item**
4. Se ok → cria `Entry` com texto + vetor
5. Substitui `entries[workspaceID]` pelos novos entries

#### `Search`

1. Se `Count(workspaceID) == 0` → retorna `nil`
2. `embedder.Embed(ctx, query)` → vetor da consulta
3. Se falhar → `fmt.Printf("[Knowledge] WARNING: ... fallback")` + `AllTexts()`
4. Cosine similarity com todos os entries do workspace
5. Ordena por similaridade (decrescente)
6. Retorna top-K textos

#### `AllTexts`

Retorna todos os textos de um workspace **sem os vetores** — usado como fallback
quando o embedding da query falha.

---

## Integração no Engine

### `engine.go` — Inicialização

Localização: `internal/engine/engine.go:315-345`

```go
// Lê fixed model "embedding"
embedProvider, embedModel, _ := store.GetFixedModel("embedding")

// Cria embedder com credenciais do provider
emb := knowledge.NewEmbedder(embedProvider, embedModel, pCfg.APIURL, apiKey)
knowledgeIndex = knowledge.NewIndex(emb)

// Indexa workspace ativo no startup
wsID := store.WorkspaceIDByPath(activeWorkspacePath)
knowledgeIndex.IndexWorkspace(ctx, wsID, ws.Knowledge)
```

**Engine struct** (engine.go:57):
```go
type Engine struct {
    // ...
    KnowledgeIndex *knowledge.Index
}
```

**Engine return** (engine.go:681):
```go
return &Engine{
    // ...
    KnowledgeIndex: knowledgeIndex,
}
```

### `ExtraContext` — Consumo na query do LLM

Localização: `internal/engine/engine.go:403-428`

```go
// Conhecimento — busca semântica quando possível, fallback para completo
if len(ws.Knowledge) > 0 {
    if knowledgeIndex != nil && userInput != "" {
        wsID := store.WorkspaceIDByPath(ws.Path)
        if wsID > 0 && knowledgeIndex.Count(wsID) > 0 {
            results := knowledgeIndex.Search(ctx, userInput, wsID, 5)
            if len(results) > 0 {
                layers.WriteString("=== KNOWLEDGE (relevant) ===\n")
                layers.WriteString(strings.Join(results, "\n---\n"))
                layers.WriteString("\n")
            }
        } else {
            // Fallback: dump completo (compatibilidade)
        }
    } else {
        // Fallback: dump completo (compatibilidade)
    }
}
```

### `app_workspaces.go` — Reindexação ao salvar

Localização: `app_workspaces.go:37-44`

```go
func (a *App) SaveWorkspace(ws workspace.WorkspaceConfig) {
    a.eng.DB.AddWorkspace(ws)

    // Reindexa o knowledge index
    if a.eng.KnowledgeIndex != nil {
        wsID := a.eng.DB.WorkspaceIDByPath(ws.Path)
        if wsID > 0 {
            a.eng.KnowledgeIndex.IndexWorkspace(context.Background(), wsID, ws.Knowledge)
        }
    }
    // ...
}
```

### `db.Store` — Helper novo

Localização: `internal/db/db.go:131-139`

```go
func (s *Store) WorkspaceIDByPath(path string) int64
```

Converte o `path` do workspace (string usada nas sessions) para o `ID` numérico
(int64) usado como chave no índice.

---

## Regras de Erro (Fallback Chain)

```
Search(query, wsID, topK)
│
├── Count(wsID) == 0 → nil
│
├── embedder.Embed(query) OK → cosine similarity → top-K
│
└── embedder.Embed(query) FALHA
    ├── fmt.Printf("[Knowledge] WARNING: ... fallback para conteúdo completo")
    └── AllTexts(wsID) → retorna TODOS os itens
```

```go
fmt.Printf("[Knowledge] WARNING: search embed query falhou para workspace %d (%v) — fallback para conteúdo completo\n",
    workspaceID, err)
```

Da mesma forma, `IndexWorkspace` pula itens que falham ao embedar:

```
IndexWorkspace:
├── Item A → embed OK → adiciona
├── Item B → embed FALHA → pula com warning
├── Item C → embed OK → adiciona
└── Substitui entries[wsID] = {A, C}
```

---

## Configuração

Para o sistema funcionar, o fixed model `"embedding"` precisa estar configurado
em **Settings > Models** no frontend:

| Campo | Exemplo |
|---|---|
| Provider | `openai` (ou qualquer provider configurado) |
| Model | `text-embedding-3-small` (ou `bge-large-en`, etc.) |

O provider precisa ter um modelo com `Embedding: true` (inferido
automaticamente se o nome contiver "embed", "e5" ou "bge" — veja
`internal/provider/infer.go:20`).

Se o modelo `embedding` não estiver configurado, o Engine loga:
```
[Engine] WARNING: embedding model not configured — semantic knowledge search disabled
```
E o comportamento cai para o dump completo (compatibilidade total).

---

## Dependências

- `net/http` — padrão da stdlib
- Nenhuma nova dependency externa
- O `ada-llm-client` **não é usado** (não tem suporte a embeddings)

---

## Limitações e Próximos Passos

| Situação | Estado |
|---|---|
| Indexar workspace ativo no startup | ✅ Implementado |
| Reindexar ao salvar workspace | ✅ Implementado |
| Fallback para dump completo em caso de erro | ✅ Implementado |
| Reindexar ao trocar de workspace ativo | ❌ Pendente |
| Persistência de vetores entre restarts | ❌ Não fará (reindexar é intencional) |
| Cache de embeddings para evitar chamadas duplicadas | ❌ Futuro |
| Rate-limit em workspaces com muitos itens | ❌ Futuro |
| Reindexar webhook quando knowledge muda fora do save | ❌ Futuro |

---

## Testes

```bash
go test ./internal/knowledge/ -v
```

13 testes cobrindo:

- **Cosine similarity**: idêntico, ortogonal, oposto, lengths diferentes, vazio
- **Index**: criação, count vazio/populado
- **Search**: match exato, workspace vazio, workspace diferente
- **Fallback**: erro de embedding retorna todos os textos
- **AllTexts**: textos sem vetores
- **Concorrência**: leitura/escrita simultânea

```
=== RUN   TestCosineSimilarity_Identical          --- PASS
=== RUN   TestCosineSimilarity_Orthogonal         --- PASS
=== RUN   TestCosineSimilarity_Opposite           --- PASS
=== RUN   TestCosineSimilarity_DifferentLengths   --- PASS
=== RUN   TestCosineSimilarity_Empty              --- PASS
=== RUN   TestNewIndex                            --- PASS
=== RUN   TestIndex_Count                         --- PASS
=== RUN   TestIndex_Search_ExactMatch             --- PASS
=== RUN   TestIndex_Search_EmptyWorkspace         --- PASS
=== RUN   TestIndex_Search_DifferentWorkspace     --- PASS
=== RUN   TestIndex_Search_EmbedError_FallsBack   --- PASS
=== RUN   TestIndex_AllTexts                      --- PASS
=== RUN   TestIndex_Concurrency                   --- PASS
```

# Chat Summary — Sumarização Incremental de Histórico

> Pacote `internal/chatsummary/` que gerencia sumarização assíncrona e
> incremental do histórico de conversa por sessão, evitando perder contexto
> quando o pruning (`ada-context`) descarta mensagens antigas.

## Contrato

### O que é

Um mecanismo que, à medida que novas mensagens são adicionadas a uma sessão,
dispara **em background** uma chamada LLM para condensar todo o histórico
em um resumo conciso (~300 tokens). O resumo é persistido em disco e
prependido ao contexto nas chamadas seguintes, garantindo que informação
importante das mensagens antigas nunca seja perdida — mesmo depois que o
`ada-context` as podar.

### O que NÃO substitui

- ❌ O pruning do `ada-context` (continua sendo responsabilidade do `contextprovider`)
- ❌ O workspace summary (`internal/summary/`) — são complementares
- ❌ A store de sessões no banco SQLite — usa disco próprio

## Estrutura do pacote

```
internal/chatsummary/
├── manager.go       # API pública: Manager, Push, Get, Close
├── summarizer.go    # LLM call para condensar mensagens
├── store.go         # I/O em disco (JSONL + summary)
└── manager_test.go  # Testes do pacote
```

### Dependências

- `github.com/upperxcode/ada-llm-client` — tipos `llm.Message` para chamar o LLM
- stdlib (`os`, `sync`, `context`, `bufio`, `encoding/json`, `fmt`, `log`)

Nenhuma dependência externa nova.

## Arquitetura

### Disco

Cada sessão tem seu próprio diretório em `{baseDir}/{sessionID}/`:

```
{baseDir}/
└── {sessionID}/
    ├── last_summary.txt   ← último resumo gerado (plain text, sobrescrito)
    └── messages.jsonl     ← append-only, um JSON por linha
```

#### messages.jsonl

Formato append-only. Cada linha é um JSON com `role` e `content`:

```jsonl
{"role":"user","content":"qual a capital do Brasil?"}
{"role":"assistant","content":"Brasília"}
{"role":"user","content":"e a da Argentina?"}
{"role":"assistant","content":"Buenos Aires"}
```

- Criado na primeira mensagem se não existir
- Nunca reescrito — só append
- Leitura: `bufio.Scanner` + `json.Unmarshal`

#### last_summary.txt

- Sobrescrito a cada novo resumo gerado pela goroutine
- Se não existir, considera-se que não há resumo anterior
- Texto plano, sem formatação especial

### Tipos

```go
// RawMessage é a versão simplificada usada pelo pacote.
// Carrega apenas o necessário para sumarização e formatação.
type RawMessage struct {
    Role    string `json:"role"`    // "user" | "assistant" | "thinking" | "system"
    Content string `json:"content"` // texto da mensagem
}

// LLMClient é a interface que o summarizer usa.
// O usuário do pacote fornece uma implementação concreta.
type LLMClient interface {
    Chat(ctx context.Context, messages []llm.Message) (string, error)
}
```

### Manager (API pública)

```go
func NewManager(baseDir string, llmClient LLMClient) *Manager

func (m *Manager) Push(ctx context.Context, sessionID string, msg RawMessage, maxSend int) (string, error)
func (m *Manager) Get(ctx context.Context, sessionID string, maxSend int) (string, error)
func (m *Manager) Close() error
```

---

## Fluxo do Push

```
Push(sessionID, msg, maxSend)
  │
  ├─ Síncrono (usa ctx do caller):
  │   ├─ 1. appendMessage() — escreve msg no messages.jsonl
  │   ├─ 2. readSummary() — carrega last_summary.txt (se existir)
  │   ├─ 3. readRecentMessages() — lê últimas maxSend mensagens
  │   ├─ 4. buildContext() — monta string final
  │   └─ 5. Retorna contexto formatado (NUNCA espera goroutine)
  │
  └─ Goroutine (context.Background(), 30s timeout):
      ├─ 1. readAllMessages() — lê TODAS as mensagens do JSONL
      ├─ 2. Se total <= maxSend → SKIP (histórico ainda cabe)
      ├─ 3. generateSummary() — chama LLM com TODAS as mensagens
      ├─ 4. writeSummary() — sobrescreve last_summary.txt
      └─ 5. Se falhar → log.Warning (nunca bloqueia, nunca retorna erro)
```

### Regras da goroutine

| Regra | Motivo |
|---|---|
| Usa `context.Background()` | Nunca cancela por causa do caller |
| Timeout interno de 30s | LLM não pode travar indefinidamente |
| Só chama LLM se `total > maxSend` | Evita custo desnecessário |
| Loga warning em caso de erro | Falha de sumarização não quebra o fluxo |
| `sync.WaitGroup` no Manager | `Close()` aguarda todas as goroutines (timeout 5s) |

---

## Contexto formatado

O que Push/Get retornam:

```
[Resumo das mensagens anteriores — se existir]

role: conteúdo da mensagem
role: conteúdo da mensagem
...
```

### Exemplo real

```
O usuário perguntou sobre capitais de países.
O assistente respondeu sobre Brasil e Argentina.
O usuário quer saber agora sobre Chile.

user: qual a capital do Chile?
assistant: Santiago
```

### Regras de formatação

- Se `last_summary.txt` existe: seu conteúdo vem primeiro, seguido de `\n\n`
- As últimas `maxSend` mensagens completas vêm depois
- Cada mensagem no formato `role: content`, separadas por `\n`
- Nunca há quebra de linha entre mensagens consecutivas (só `\n`)

---

## Summarizer (LLM)

### System prompt usado

```
You are a conversation summarizer. Your task is to condense the conversation
below into a concise summary of at most 300 tokens.

Preserve in the summary:
- Important decisions that were made
- Problem context and requirements discussed
- Agreed next steps or action items

Focus only on information that would be relevant for continuing the conversation.
Output the summary in plain text. Do not add a preamble, title, or commentary.
```

### Como as mensagens são enviadas

1. System message com o prompt acima
2. Todas as mensagens da conversa, em ordem, como `user`/`assistant`/etc.

### Tratamento de erros

- Timeout de 30s → log + goroutine termina
- LLM retorna erro → log + goroutine termina
- Summary vazio → não persiste (writeSummary não é chamado)
- Em nenhum caso o erro propaga para o caller

---

## Testes

```go
// manager_test.go — 9 testes, todos passando

TestPush_NewSession           // Cria arquivos, verifica estrutura no disco
TestPush_MultipleMessages     // Append no JSONL, verifica 3 linhas
TestGet_WithoutSummary        // Retorna mensagens sem resumo (summary não existe)
TestGet_WithExistingSummary   // Retorna resumo + mensagens
TestPush_ContextFormat        // Verifica formato exato "summary\n\nrole: content"
TestPush_TrimsToMaxSend       // Respeita maxSend (corta as mais antigas)
TestConcurrentAccess          // 10 goroutines Push concorrentes
TestPush_ErrorOnEmptyContent  // Conteúdo vazio não quebra
TestClose_Timeout             // Mock lento causa timeout de 5s
```

### Mock do LLMClient

```go
type mockLLMClient struct {
    chatFunc func(ctx context.Context, messages []llm.Message) (string, error)
}

func (m *mockLLMClient) Chat(ctx context.Context, messages []llm.Message) (string, error) {
    return m.chatFunc(ctx, messages)
}
```

---

## Integração (responsabilidade de outro agente)

O `chatsummary` é apenas o pacote — **não** está integrado no fluxo de mensagens.
A integração no `contextprovider` e nos workers será feita por outro agente.

### O que será necessário:

1. **No `chat.go`** (`internal/chat/`):
   - Instanciar `chatsummary.NewManager(baseDir, llmAdapter)`
   - No `Send()`, após receber a resposta, chamar `manager.Push()` com a mensagem do usuário + resposta

2. **No `contextprovider` / `engine.go`**:
   - Substituir ou complementar o envio do histórico bruto pelo contexto retornado por `manager.Get()`
   - Decidir se envia summary + últimas N mensagens vs. mensagens completas

3. **No lifecycle da sessão**:
   - Chamar `manager.Close()` no shutdown
   - Limpar diretório da sessão ao deletar a sessão

---

## Decisões de design

### Por que disco ao invés de banco?

- `messages.jsonl` é append-only — I/O sequencial, sem locks
- `last_summary.txt` é sobrescrito atomicamente (`os.WriteFile`)
- Fácil de debugar: `cat messages.jsonl` mostra o histórico completo
- Zero dependência de schema SQL, migrations, ou transações

### Por que RawMessage próprio ao invés de reusar `core.RawMessage`?

- `core.RawMessage` tem campos demais (`ToolCalls`, `Time`, `ThinkingContent`) que o sumarizador não precisa
- Nosso `RawMessage` é um subset estável, definido no próprio pacote
- Impede acoplamento com o pacote `core`

### Por que goroutine com `context.Background()`?

- A chamada LLM para sumarizar não deve ser cancelada porque o usuário fechou o chat ou navegou
- Se o usuário mandar 3 mensagens rápido, cada Push dispara uma goroutine — a última a terminar vence
- O `Close()` com timeout 5s evita vazamento de goroutines no shutdown

### Por que só chama LLM quando `total > maxSend`?

- Enquanto o histórico couber no limite, não precisa sumarizar
- Evita custo de LLM e latência desnecessários
- O resumo anterior (se existir) continua valendo

### Por que `buildContext()` separa summary de mensagens com `\n\n`?

- O LLM consegue distinguir claramente "isto é um resumo do passado distante" vs. "estas são as últimas mensagens"
- O summary pode ser longo (até 300 tokens) — precisa de separação clara
- Mensagens recentes ficam legíveis para debug

---

## Referências

- Código: `internal/chatsummary/`
- LLM client: `github.com/upperxcode/ada-llm-client` (`llm.Message`)
- Pruning de histórico: `ada-context` (pacote separado, usado pelo contextprovider)
- Workspace config: `internal/config/workspace/workspace.go` (`MaxPromptSend`)

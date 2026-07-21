
# Chat UI Redesign + Chain-of-Thought

## Resumo
Redesign completo da interface de chat: chain-of-thought colapsável, redesign de balões com avatares, animações de entrada, footer melhorado. Inspirado nos componentes `Reasoning`, `Message`, `Shimmer` do `svelt-elements` (reescritos localmente, zero dependências novas).

---

## Arquitetura de Mudanças

### Backend (Go) — 2 arquivos

**1. `ada-love-core/types.go`** — Adicionar 2 campos no `RawMessage`:
```go
ThinkingContent string `json:"thinking_content,omitempty"` // conteúdo de raciocínio do LLM
ThinkingDuration int    `json:"thinking_duration,omitempty"` // duração em segundos
```

**2. `internal/chat/chat.go`** — Adicionar evento `chat:thinking`:
- No `handleStreamEvent`, detectar `reasoning_content` nos tokens do streaming
- Emitir `chat:thinking` com `{session_id, content, duration}` quando reasoning tokens chegam
- No `stream-finished`, salvar `thinking_content` + `thinking_duration` na mensagem via `orch.SaveMessages()`
- Fallback: LLMs sem reasoning → evento não emitido → frontend não mostra CoT (degradação graciosa)

### Frontend (Svelte) — 6 novos + 1 modificado

**Novos componentes** (`components/chat/`):

| Arquivo | Responsabilidade |
|---|---|
| `Reasoning.svelte` | Container colapsável com tracking de duração. Props: `isStreaming`, `open`, `duration`. Auto-close após 1s. Usa `Collapsible` existente. |
| `ReasoningTrigger.svelte` | Header: ícone `brain` + texto dinâmico ("Thinking..." → "Thought for 12s") + chevron animado. Usa `CollapsibleTrigger`. |
| `ReasoningContent.svelte` | Body colapsável com slide-in/out animations. Usa `MarkdownRenderer` para conteúdo. |
| `MessageBubble.svelte` | Wrapper de mensagem: avatar (32px) + content area (card sutil `bg-secondary/30 border-subtle rounded-xl`) + footer. |
| `ThinkingShimmer.svelte` | 2-3 linhas shimmer animado para loading state do CoT. CSS keyframes puro. |
| `reasoning-context.svelte.ts` | Context Svelte 5 com `#isStreaming`, `#isOpen`, `#duration` (private fields, getters/setters). |

**Modificado: `ChatPanel.svelte`**:
- Interface `Message` ganha campos: `thinkingContent?: string`, `thinkingDuration?: number`
- Escuta novo evento `chat:thinking` → popula thinkingContent/thinkingDuration por mensagem
- Remove: `showThinking` box atual, `"Ada is typing..."` indicator
- Substitui rendering inline de mensagens por `<MessageBubble>` component
- Footer: hover-reveal (`group-hover:opacity-100`) + copy button com feedback animado (ícone `check` por 2s)
- `loadSession`: mapeia `thinking_content` e `thinking_duration` dos RawMessages

### Balões de Mensagem (design)

**Usuário:**
- `self-end max-w-[75%] rounded-2xl rounded-br-md bg-[var(--accent-primary)] text-sm shadow-md shadow-black/20`
- Avatar à esquerda: circle 32px, `bg-accent-primary/20`, ring, iniciais

**Assistente:**
- Layout flex row: avatar (ícone `bot`, 32px) + content area (full-width)
- Content area: `rounded-xl bg-[var(--bg-secondary)]/30 border border-[var(--border-subtle)] px-4 py-3`
- CoT: `Reasoning` aparece dentro do content area, acima do MarkdownRenderer

**Animações:** Todas as mensagens: `animate-in fade-in slide-in-from-bottom-2 duration-300`

### Wails Events (interface de comunicação)

| Evento | Direção | Status |
|---|---|---|
| `chat:delta` | Backend → Frontend | existente (sem mudanças) |
| `chat:thinking` | Backend → Frontend | **novo** — `{session_id, content, duration}` |
| `chat:turnEnd` | Backend → Frontend | existente (sem mudanças) |
| `chat:error` | Backend → Frontend | existente (sem mudanças) |
| `SendMessage()` | Frontend → Backend | existente (sem mudanças na assinatura) |

### O que NÃO muda
- `SendMessage` API assinatura
- `Orchestrator`, `StorageAdapter`, `Session` struct
- `app_sessions.go`, `app_chat.go` (binding)
- `Sidebar.svelte`, `ChatLayout.svelte`
- MarkdownRenderer, icon-map, theme system

---

## Ordem de Implementação

1. **Backend**: Adicionar campos `ThinkingContent`/`ThinkingDuration` no `RawMessage` (ada-love-core)
2. **Backend**: Adicionar evento `chat:thinking` no `chat.go` + parse de reasoning tokens
3. **Frontend**: Criar `reasoning-context.svelte.ts`
4. **Frontend**: Criar `Reasoning.svelte`, `ReasoningTrigger.svelte`, `ReasoningContent.svelte`
5. **Frontend**: Criar `ThinkingShimmer.svelte`
6. **Frontend**: Criar `MessageBubble.svelte` (avatar + content area + footer)
7. **Frontend**: Refatorar `ChatPanel.svelte` — substituir rendering inline, adicionar listener `chat:thinking`, remover indicadores antigos
8. **Frontend**: Atualizar `Message` interface + `loadSession` para mapear campos de CoT
9. **Teste**: Verificar streaming, persistência, reload de sessão, fallback sem reasoning

---

## Design Doc
Salvar em `docs/superpowers/specs/2026-07-21-chat-redesign-design.md` e fazer commit.

# Chat UI Redesign + Chain-of-Thought

**Data:** 2026-07-21
**Status:** Implementado
**Inspirado em:** `svelt-elements` (reescrito localmente, zero dependências novas)

## Resumo

Redesign completo da interface de chat com 3 objetivos:
1. **Simplificar** — interações claras e guiadas
2. **Humanizar** — avatares, tom conversacional, animações
3. **Equilibrar** — layout limpo que elimina atritos

## O que foi implementado

### Backend (ada-love-core)

**`types.go`** — 2 campos novos no `RawMessage`:
- `ThinkingContent string` — conteúdo de raciocínio do LLM (para uso futuro)
- `ThinkingDuration int` — duração em segundos

### Frontend — 6 novos arquivos

| Arquivo | Responsabilidade |
|---|---|
| `reasoning-context.svelte.ts` | Context Svelte 5 com `#isStreaming`, `#isOpen`, `#duration` (private fields) |
| `Reasoning.svelte` | Container colapsável com tracking de duração, auto-close após 1s |
| `ReasoningTrigger.svelte` | Header: ícone brain/loader + "Thinking..." → "Thought for 12s" + chevron animado |
| `ReasoningContent.svelte` | Body colapsável com slide-in/out animations + shimmer loading |
| `ThinkingShimmer.svelte` | 3 linhas shimmer animado (CSS keyframes puro) |
| `MessageBubble.svelte` | Wrapper: avatar + content area (card sutil) + footer hover-reveal |

### Frontend — 1 arquivo modificado

**`ChatPanel.svelte`:**
- `Message` interface ganha `thinkingDuration?: number`
- `chat:turnEnd` calcula e salva `thinkingDuration` na mensagem
- Rendering inline substituído por `<MessageBubble>` component
- Removidos: `thinkingContent`, `showThinking`, thinking box, "Ada is typing..." indicator
- Footer com hover-reveal (`group-hover:opacity-100`) + copy com feedback animado

## Decisões arquiteturais

### Por que não persistir CoT no backend agora?

O `ada-storage-module` usa colunas individuais no SQLite (não JSON blobs). Persistir `thinking_content`/`thinking_duration` exigiria:
- Nova migração no `ada-storage-module`
- Mudanças no `storage.Message` struct
- Mudanças no adapter `AppendMessage`/`GetMessages`

O `ada-llm-client` atualmente descarta `reasoning_content` no parse SSE (`streamChunk.Delta` só lê `content`). Sem dados reais de raciocínio, a persistência seria inútil.

**Decisão:** Duração trackeada client-side. Backend event `chat:thinking` e persistência real serão implementados quando o upstream suportar `reasoning_content`.

### Por que reescrever em vez de importar `svelt-elements`?

A lib traria `svelte-streamdown`, `shiki`, `runed`, `mode-watcher`, `@ai-sdk/svelte` — 5+ dependências pesadas para usar 3 componentes. A reescrita local usou:
- `$effect()` nativo no lugar de `runed/watch`
- `Collapsible` já existente no projeto (shadcn/bits-ui)
- `Icon` system já existente no projeto
- CSS custom properties do tema já existente

## Eventos Wails (interface de comunicação)

| Evento | Direção | Status |
|---|---|---|
| `chat:delta` | Backend → Frontend | existente (sem mudanças) |
| `chat:turnEnd` | Backend → Frontend | existente (sem mudanças) |
| `chat:thinking` | Backend → Frontend | **futuro** — `{session_id, content, duration}` |

## Design visual

### Balão do usuário
- `rounded-2xl rounded-br-md` (estilo speech bubble)
- `shadow-md shadow-black/20`
- Avatar "ME" à direita (circle 32px, accent/20)

### Balão do assistente
- Full-width com avatar `bot` à esquerda (32px, bg-tertiary, ring)
- Content area: `rounded-xl bg-secondary/30 border-subtle`
- Animação: `animate-in fade-in slide-in-from-bottom-2 duration-300`

### Reasoning (CoT)
- Badge "Thinking..." com spinner → "Thought for 12s" colapsado
- Expande com slide-in animation
- Shimmer placeholder durante streaming sem conteúdo

### Footer
- Hover-reveal: `opacity-0 group-hover:opacity-100`
- Badge de tempo: `bg-tertiary rounded-md`
- Copy com feedback: ícone `check` + "Copiado!" por 2s
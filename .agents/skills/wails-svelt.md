---
name: wails-svelte-architecture
description: Garante o alinhamento arquitetural do projeto Wails + Svelte 5 + TypeScript + Shadcn, impedindo violações de backend e promovendo padrões modernos de frontend.
---

# Wails & Svelte Architectural Assistant

Esta skill orienta o agente a desenvolver funcionalidades estritamente alinhadas com o ecossistema e com a hierarquia de dependências do projeto.

## Quando usar esta Skill

Use automaticamente sempre que o usuário solicitar:
- Criação de novas telas, componentes ou layouts no frontend.
- Implementação de novas funcionalidades, serviços, pacotes ou tabelas de banco de dados no backend (Go).
- Integração e comunicação de dados entre o Svelte e o Go usando os bindings do Wails.
- Resolução de bugs visuais, problemas de concorrência ou sincronização de estado.

---

## Diretrizes de Execução

### Passo 1: Validar a Hierarquia do Backend (Regra em Pedra)
Antes de escrever qualquer linha de código em Go, certifique-se de que o pacote modificado respeita o fluxo acíclico:
- `app_*.go / main.go` -> `engine` -> Camadas de Serviço (`chat`, `commands`, `db`, `modelselect`, etc.).
- O pacote `core` contém apenas interfaces e nunca importa pacotes internos.
- Adapters pertencem à camada de baixo. Não introduza ciclos.

### Passo 2: Aplicar Padrões de Frontend (Svelte 5 + TypeScript)
- **Estado Reativo:** Use a sintaxe de Runes do Svelte 5 (`$state`, `$derived`, `$effect`) em vez da reatividade antiga do Svelte 4 (`let` reativo ou `$: `).
- **Sem HTTP/HTMX:** Toda comunicação deve ser feita invocando diretamente as funções assíncronas geradas pelo Wails em `../wailsjs/go/`.
- **Tipagem Segura:** Crie ou exija interfaces TypeScript espelhando fielmente as structs do Go retornadas pelo backend.

### Passo 3: Componentização e UI (Shadcn-Svelte + Tailwind)
- Componentes complexos (modais, menus, dropdowns) devem ser modulares e injetados via Shadcn em `$lib/components/ui/`.
- Ícones devem utilizar o pacote padrão `lucide-svelte`.
- O estilo visual deve seguir estritamente o tema utilitário do Tailwind mapeado em `src/app.css`.

### Passo 4: Tratar Falhas na Interface (Sem Mascarar Erros)
- Nunca esconda falhas de carga de dados com fallbacks invisíveis ou dropdowns misteriosamente vazios.
- Diante de dados faltantes ou erros de persistência, exiba um estado visual de erro claro (ícone de alerta + mensagem amigável de erro), emita logs detalhados e forneça um botão de "Tentar Novamente".

---

## Exemplos de Resposta ao Usuário

### Exemplo 1: Adicionando uma nova tabela ao Banco e exibindo na UI

**Prompt do Usuário:** *"Preciso listar as sessões salvas na barra lateral."*

**Ação do Agente:**
1. Criar a struct no arquivo `internal/db/session_store.go` e expor na interface do `core`.
2. Integrar a chamada no Svelte 5 usando `$state` e carregando no `onMount`:

```html
<script lang="ts">
  import { ListarSessões } from "../wailsjs/go/main/App";
  import { onMount } from "svelte";

  // Svelte 5 Runes para estado
  let sessoes = \$state<string[]>([]);
  let erro = \$state<string | null>(null);

  onMount(async () => {
    try {
      sessoes = await ListarSessões();
    } catch (e) {
      erro = "Não foi possível carregar o histórico.";
      console.error("[UI Error] Falha ao buscar sessões:", e);
    }
  });
</script>

<div class="space-y-2">
  {#if erro}
    <p class="text-xs text-red-400">{erro}</p>
  {:else}
    {#each sessoes as sessao}
      <button class="text-sm text-zinc-400 hover:text-white">{sessao}</button>
    {/each}
  {/if}
</div>
```

---

## Verificação de Qualidade (Checklist do Agente)

Antes de entregar qualquer código, certifique-se mentalmente:
- [ ] O código em Go introduz algum ciclo de importação?
- [ ] O componente Svelte está usando Runes do Svelte 5 corretamente?
- [ ] O TypeScript está validando estritamente os retornos do Wails?
- [ ] Caso o backend falhe, a interface vai expor o erro em vez de mascará-lo?

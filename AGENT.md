

---name: "Developer Guidelines"
description: "Regras de desenvolvimento e filosofia adotada pelo time"
tools: []
model: ""
maxTurns: 0
skills: []
mcpServers: []
---
# Hierarquia de Dependências — REGRA GRAVADA EM PEDRA
As dependências entre pacotes DEVEM seguir ESTRITAMENTE esta hierarquia. Nenhum PR será aceito se violar estas regras.


[Camada 1: Entrada]
└── app_*.go / main.go
       │
       ▼
[Camada 2: Orquestração]
└── engine (composition root)
       │
       ├─► chat ──────────────┐
       ├─► commands ──────────┤
       ├─► sessionfetch ──────┤
       ├─► sessionstore ──────┤
       └─► modelselect ───────┼─► [Camada 3: Persistência]
                              │   └── db
                              │          │
                              ▼          ▼
                       [Camada 4: Contratos Globais]
                       └── core (interfaces)
                              │
                              ├─► session/types (RawMessage, Session)
                              └─► config/* (provider, workspace, agent, …)

## Regras absolutas

1. **core NUNCA pode importar nenhum pacote interno.** `core` contém apenas interfaces e o orquestrador. Suas únicas dependências são a stdlib e pacotes externos (ex.: `ada-llm-client`, `ada-commands`).
2. **A direção da seta é SAGRADA:** pacotes de cima importam pacotes de baixo. NUNCA o contrário.
   - `app_*.go` → `engine` → `chat`, `commands`, `db`, `modelselect`, ...
   - `db` → `core` (implementa interfaces)
   - `sessionfetch`, `sessionstore` → `db`, `session/types`
3. **NUNCA use `var _ = ...` para silenciar imports.** Se um import não é usado explicitamente, remova-o. Tipos inferidos por chamadas de função não contam como uso.
4. **Adapters que conectam camadas pertencem à camada DE BAIXO.** Exemplo: `StorageAdapter` (que adapta `*Store` para `core.StorageEngine`) está em `internal/db/adapter.go`, NÃO em `core/`.
5. **Zero ciclos.** O grafo de dependências deve ser sempre acíclico. Ciclos são bugs arquiteturais e devem ser eliminados na raiz, nunca contornados.
6. **`config/*` e `session/types` são folhas:** não importam nenhum outro pacote interno. Apenas stdlib e types externos.

## Consequências de violação

- Violação da regra #1 ou #2 = rejeição automática do PR.
- Violação da regra #3 ou #4 = correção obrigatória antes do merge.
- Qualquer ciclo introduzido = rollback imediato.

---

# KISS — Keep It Simple, Stupid

Mantenha tudo simples. Prefira soluções diretas e fáceis de entender em vez de abstrações complexas. Simplicidade facilita revisão de código, testes e manutenção a longo prazo.

## Princípios práticos (Backend & Arquitetura)

- Mantenha arquivos pequenos e com responsabilidade única, não muito maior que 300 linhas.
  - Cada arquivo deve ter foco claro; evite arquivos gigantes com muitas responsabilidades.
- Em Go, prefira funções/métodos pequenos e coesos.
  - Extraia funções quando um trecho de código atinge mais do que algumas linhas ou faz mais de uma coisa.
  - Coloque funções auxiliares relacionadas em arquivos separados com nomes descritivos (ex.: `db_migrations.go`, `workspace_store.go`, `agent_parser.go`).
- Prefira composição a herança; mantenha as interfaces pequenas e específicas.
- Escreva testes automáticos para funcionalidades críticas.
- Documente decisões arquiteturais importantes no repositório (`README`, `docs/`) — não no código apenas.
- Ao introduzir uma nova dependência, avalie custo/benefício e prefira dependências pequenas e ativas.

### Exemplo de Organização de Arquivos (Go)
Se um serviço possui partes bem definidas como inicialização, migrations e handlers, fragmente-o logicamente:
- `init.go`
- `migrations.go`
- `handlers.go`
- `store.go`

---

# Stack Frontend Oficial — Svelte 5, Shadcn-Svelte & TypeScript

A interface do aplicativo não utiliza templates no backend (SSR artificial) ou requisições HTTP simuladas (HTMX). Adotamos uma separação clara entre **Frontend (Visual e Estado Local)** e **Backend (Regras de Negócio e Acesso ao S.O.)** utilizando os bindings automáticos do Wails.


┌────────────────────────────────────────────────────────┐
│ FRONTEND (Vite)                                        │
│ Svelte 5 (UI/Estado) ──> Shadcn-Svelte + Tailwind      │
└───────────────────────────┬────────────────────────────┘
│ (Chamadas Diretas)
▼
┌────────────────────────────────────────────────────────┐
│ BACKEND (Wails)                                        │
│ Wails Bindings ──> Go Engine ──> DB/System             │
└────────────────────────────────────────────────────────┘


## Diretrizes de Desenvolvimento do Frontend

1. **Separação Rígida de Responsabilidades:**
   - **Go (Backend):** Processa dados pesados, gerencia concorrência, acessa banco de dados e APIs externas. Retorna structs tipadas.
   - **Svelte (Frontend):** Controla o fluxo de telas, reatividade local e estados de componentes (aberto/fechado, carregando, transições).
2. **Tipagem Estrita com TypeScript:**
   - Todas as chamadas para as funções Go devem consumir os arquivos gerados automaticamente pelo Wails em `frontend/wailsjs/go/`.
   - Modificações em structs do Go exigem a reconfiguração/verificação automática dos tipos gerados no frontend para evitar falhas silenciosas de interface.
3. **Componentização Modular (Shadcn-Svelte):**
   - Não reinvente componentes complexos (Modais, Comboboxes, Alertas). Use a CLI do `shadcn-svelte` para injetá-los na pasta `$lib/components/ui/`.
   - Modificações visuais nos componentes devem respeitar o padrão de classes utilitárias do Tailwind CSS.
4. **Gerenciamento de Ícones e Temas:**
   - Use o ecossistema `lucide-svelte` para garantir ícones vetoriais padronizados e performáticos.
	   - Mudanças de cor globais e suporte a temas (Ex: Dark/Light Mode) devem ser controlados puramente por variáveis CSS nativas integradas ao arquivo `src/app.css` mapeado no Tailwind.

5. **Padrão de Formulários de Entidades (Settings):**
   - **Layout Row-based:** Campos curtos devem seguir o padrão `SettingRow` (Label/Descrição à esquerda, Input à direita).
   - **Textareas Longos:** Campos de texto extenso (ex: `System Prompt`, `Content`, `Personality`) devem usar `fullWidth: true` no config, ocupando a largura total com Label acima e Input abaixo.
   - **Gestão de Ícones/Cores:** Removidos dos campos de formulário. Ícones e cores agora são gerenciados exclusivamente via `EntityHeader`, acessíveis por cliques no ícone (emoji picker) ou botão dedicado (color picker) na barra superior do diálogo.

### Exemplo Prático de Consumo Tipado (Svelte 5)

```html
<script lang="ts">
  import { Button } from "\$lib/components/ui/button";
  import * as Card from "\$lib/components/ui/card";
  import { Cpu, CheckCircle } from "lucide-svelte";
  import { onMount } from "svelte";
  
  // Binding automático gerado pelo Wails (TypeScript de ponta a ponta)
  import { ObterStatusSistema } from "../wailsjs/go/main/App"; 

  // Garantia de tipagem alinhada com as structs do Go
  interface StatusSistema {
    online: boolean;
    usoCpu: number;
  }

  let status = \$state<StatusSistema | null>(null);

  onMount(async () => {
    // Chamada assíncrona nativa do Wails, sem overhead HTTP
    status = await ObterStatusSistema();
  });
</script>

<main class="p-6 max-w-sm mx-auto space-y-4">
  {#if status}
    <Card.Root>
      <Card.Header>
        <Card.Title class="flex items-center gap-2">
          <CheckCircle class="text-green-500 w-5 h-5" />
          Status do Engine
        </Card.Title>
      </Card.Header>
      <Card.Content>
        <div class="flex justify-between items-center bg-muted p-3 rounded-md">
          <span class="flex items-center gap-2 text-sm text-muted-foreground">
            <Cpu class="w-4 h-4" /> Uso de CPU
          </span>
          <span class="font-bold">{status.usoCpu}%</span>
        </div>
      </Card.Content>
      <Card.Footer>
        <Button class="w-full">Reexecutar Diagnóstico</Button>
      </Card.Footer>
    </Card.Root>
  {:else}
    <p class="text-center text-muted-foreground animate-pulse text-sm">Sincronizando com Go...</p>
  {/if}
</main>
```

---

# Política de UI e Fallbacks — Não Mascarar Erros

## FALHA NÃO DEVE SER MASCARADA POR FALLBACKS AUTOMÁTICOS

- **Regra:** não implemente fallbacks na UI que ocultem ou mascarem problemas de persistência, migração ou carga de dados. Um dropdown vazio ou um campo sem valor deve expor claramente que há um problema no backend (log + mensagem de erro amigável), e a causa raiz deve ser corrigida.

- **Motivação prática:**
  - Fallbacks escondem bugs e condições de corrida. Quando a UI apresenta um valor por "fallback" não é evidente que os dados primários (por ex. tabelas normalizadas no DB) não foram carregados corretamente.
  - Isso dificulta debug e promove acúmulo de dívida técnica: correções de curto prazo viram soluções permanentes inadvertidas.

- **Comportamento esperado diante de dados faltantes:**
  1. A UI deve indicar claramente um estado "incompleto" (ex.: mensagem ou ícone informando "Dados ausentes — ver logs"), não preencher o campo com dados de outra fonte invisível.
  2. Registrar um log no frontend e no backend com contexto suficiente: endpoint/função chamada, timestamp, user action, workspace/ID e qualquer payload relevante.
  3. Fornecer um caminho de correção claro (ex.: botão "Recarregar dados", instruções para reexecutar migração, ou link para a documentação de troubleshooting).

- **Procedimentos de debugging que o time deve seguir (prioritários):**
  1. Verificar os logs do backend na inicialização — procurar por mensagens de migração e pelos logs: `[DB] fixed_model loaded` e `[DB] SaveFixedModelRow`.
  2. Consultar diretamente a tabela no DB (sqlite3) para confirmar o conteúdo de `fixed_models` e `fixed_model_tools`.
  3. Verificar se engine já concluiu a inicialização antes de servir `GetAdaConfig` (evitar condição de corrida).
  4. Confirmar que `provider_models` foram migrados para `provider_models` (`GetProvidersFull`) e que `deadaptProviderConfig` mapeou os models para `adaCfg.Providers`.

- **Quando um fallback for considerado necessário (exceção):**
  - Deve ser altamente visível e temporal: exibir claramente que é um fallback (UI badge "fallback"), criar um ticket automático/alerta e expirar o fallback após X minutos.
  - Preferir sempre mostrar a falha e exigir correção do backend em vez de esconder o problema.

- **Checklist de implementação segura (ao adicionar qualquer comportamento que envolva exibir modelos ou dados derivados):**
  - [ ] Existe logging suficiente no backend para rastrear carga/migração dos dados?
  - [ ] O frontend valida a presença explícita do dado primário antes de renderizar (ex.: `adaCfg.tiny_brain.provider !== undefined`)?
  - [ ] Em caso de ausência, a UI apresenta mensagem de erro e botão de recarregar, não uma lista preenchida por outro recurso invisível.
  - [ ] Há um caminho de correção (documentado em README/docs) para a operação de migração/seed que populará as tabelas normalizadas.

Seguindo essa política evitamos mascarar problemas e mantemos a observabilidade e correção das causas raiz. Se um dropdown estiver vazio, tratamos isso como um sinal de erro a ser investigado — não como um motivo para silenciar o bug com um fallback invisível.


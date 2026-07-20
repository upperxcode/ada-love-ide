# Plano: Corrigir os 39 erros de type do `svelte-check` (bits-ui 2.18.1)

## Diagnóstico raiz
O `bits-ui` está instalado na **2.18.1** (versão declarada em `package.json`, não há descompasso de versão). Os componentes de UI (shadcn-style) e seus consumidores foram escritos para uma API **anterior** em que `asChild` existia. Na 2.18.1 a API mudou para o padrão `WithChild`: em vez de `asChild`, usa-se um snippet `child` que recebe `{ props }` e se espalha no elemento filho. Vários componentes locais também ficaram defasados (Select, ToggleGroup, Textarea).

## Correções por arquivo

### 1. `ui/select/select-label.svelte` — Select removou `Label`/`LabelProps`
Remover o arquivo (e seu `export` em `ui/select/index.ts`) — o `Label` não existe mais na API 2.18.1. Os consumidores não o usam (verificado: nenhum `<Select.Label>` no código).

### 2. `ui/select/select-content.svelte` — `position="popper"` removido
Remover `position="popper"`. Manter `sideOffset`/`align` (ainda válidos). O `ContentProps` agora aceita `side`/`sideOffset`/`align`/`strategy` direto.

### 3. `ui/select/select-item.svelte` — `style` inválido + erro "Expected 1 arguments"
O erro "Expected 1 arguments" vem do `@render children?.()` dentro de `SelectPrimitive.Item` (o item espera snippet com arg). Corrigir: `<SelectPrimitive.Item ...>{#snippet children()}{/snippet}</SelectPrimitive.Item>` ou remover o `style` e ajustar o render. Repassar `style` para o elemento nativo via `restProps` (já espalhado) — remover o `style="color:..."` fixo que não pertence ao tipo.

### 4. `ui/toggle-group/toggle-group.svelte` — `RootProps` virou union (exige `type`)
A interface `ToggleGroupProps extends ToggleGroupPrimitive.RootProps` quebra porque `RootProps` é `Single | Multiple` (discriminated union). Corrigir tipando como `ToggleGroupPrimitive.RootProps` + props locais via interseção com `type?: "single" | "multiple"` e tornando `value`/`class`/`children`/`variant`/`size`/`spacing` próprios. Adicionar `type` ao `ToggleGroupPrimitive.Root` (os consumidores já passam `type="multiple"`).

### 5. `ui/toggle-group/toggle-group-item.svelte` — "Expected 1 arguments"
`ToggleGroupItemProps` usa snippet `child` com `{ pressed }`. O `@render children?.()` precisa do arg: `{#snippet children()}...{/snippet}` ou `{@render children?.()}`. Ajustar para o formato de snippet do item.

### 6. `ui/textarea/textarea.svelte` — `{#if}` dentro de `<textarea>` + sem default export
Um `<textarea>` não pode conter blocos Svelte. O `children` nunca é usado (o componente é controlado por `value`/`bind:value`); remover o bloco `{#if children}`. Isso devolve o default export e resolve o `index.ts`.

### 7. `Icon` (`icon/Icon.svelte`) — `style` não existe em `IconProps`
`IconProps` aceita `color`, não `style`. Converter todos os `<Icon ... style="color: ..." />` para `color="..."` em:
- `settings/EnvEditor.svelte` (rotaciona ícone)
- `settings/APIKeyManager.svelte` (2 usos)
- `settings/ModelListCollapsible.svelte`
- `settings/ModelManagerDialog.svelte` (4 usos)
- `settings/SpecWizardDialog.svelte` (1 uso)

### 8. `asChild` → snippet `child` (padrão 2.18.1)
- **Collapsible.Trigger** (já usam `{#snippet child({ props: tp })}` corretamente): só **remover o `asChild`** em `EnvEditor.svelte`, `APIKeyManager.svelte`, `ModelListCollapsible.svelte`, `SpecWizardDialog.svelte`.
- **Tooltip.Trigger** (usam `<button>` direto com `asChild`): converter para snippet `child`:
  ```svelte
  <TooltipTrigger>
    {#snippet child({ props })}
      <button {...props} type="button" onclick={...} class={...}>...</button>
    {/snippet}
  </TooltipTrigger>
  ```
  em `EntityCard.svelte` (2x) e `CardList.svelte` (1x).

### 9. `bind:open` não-bindable — `SettingsPanel.svelte` usa `bind:open` em SpecWizardDialog e EntityEditDialog
Marcar `open` como `$bindable()` nas props de `SpecWizardDialog.svelte` e `EntityEditDialog.svelte`:
`let { open = $bindable(), onOpenChange, ... }: ... = $props();`

### 10. `implicit any` em callbacks — `SpecWizardDialog.svelte`
Tipar os parâmetros dos `.filter(p => ...)`, `(_, idx) => ...` etc. com o tipo do array (ex.: `p: string`, `(_: unknown, idx: number)`), ou adicionar `// svelte-ignore` — preferencialmente tipar.

## Warnings (21, maioria a11y)
São `a11y_label_has_associated_control` e `a11y_interactive_supports_focus` (labels sem `for`, divs com role). **Não quebram o build**. Proponho deixá-los como estão nesta primeira leva (são visuais/acessibilidade, fora do escopo "erros de type"), a menos que você queira resolver também. Posso adicionar `<!-- svelte-ignore -->` onde fizer sentido.

## Verificação
Rodar `svelte-check` ao final e confirmar `0 errors`. (O build Wails/Go não é afetado — só tipos de frontend.)

## Arquivos tocados (16)
- `ui/select/select-label.svelte` (remover), `ui/select/index.ts` (remover export)
- `ui/select/select-content.svelte`, `ui/select/select-item.svelte`
- `ui/toggle-group/toggle-group.svelte`, `ui/toggle-group/toggle-group-item.svelte`
- `ui/textarea/textarea.svelte`
- `icon/Icon.svelte` (só leitura — correção nos consumidores)
- `settings/EntityCard.svelte`, `settings/CardList.svelte`, `settings/EnvEditor.svelte`, `settings/APIKeyManager.svelte`, `settings/ModelListCollapsible.svelte`, `settings/ModelManagerDialog.svelte`, `settings/SpecWizardDialog.svelte`, `settings/EntityEditDialog.svelte`, `settings/SettingsPanel.svelte` (indireto via props)

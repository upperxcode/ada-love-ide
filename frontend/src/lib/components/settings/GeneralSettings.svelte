<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import { theme } from '$lib/stores/theme.svelte';
	import SettingRow from './SettingRow.svelte';

import { providersStore } from '$lib/stores/providers.svelte';

// ── Fixed Models State ──
const FIXED_MODEL_NAMES = ['Classifier', 'embedding', 'image', 'spec', 'tinybrain'];
const FIXED_MODEL_LABELS: Record<string, string> = {
	Classifier: 'Classifier',
	embedding: 'Embedding',
	image: 'Image',
	spec: 'Spec',
	tinybrain: 'TinyBrain',
};
function displayName(name: string): string {
	return FIXED_MODEL_LABELS[name] || name;
}
let fixedModels = $state<Record<string, { provider: string; model: string; tools: string[] }>>({});
let availableTools = $state<any[]>([]);
let openStates = $state<Record<string, boolean>>({});
let toolsOpen = $state<Record<string, boolean>>({});

const providerOptions = $derived(
	providersStore.providers.map(p => ({ value: p.name, label: p.name }))
);

function getModelOptions(providerName: string) {
	const models = providersStore.getModels(providerName);
	return models.map(m => ({ value: m.name, label: m.name }));
}

function toggleOpen(name: string) {
	openStates[name] = !openStates[name];
	openStates = { ...openStates };
}

function toggleToolsOpen(name: string) {
	toolsOpen[name] = !toolsOpen[name];
	toolsOpen = { ...toolsOpen };
}

function toggleTool(name: string, toolName: string) {
	if (!fixedModels[name].tools) fixedModels[name].tools = [];
	const idx = fixedModels[name].tools.indexOf(toolName);
	if (idx >= 0) {
		fixedModels[name].tools.splice(idx, 1);
	} else {
		fixedModels[name].tools.push(toolName);
	}
	fixedModels = { ...fixedModels };
}

async function saveFixedModel(name: string) {
	const fm = fixedModels[name];
	console.log(`[FixedModel] Saving "${name}":`, JSON.stringify(fm));
	if (!fm) return;
	try {
		await (window as any).go.main.App.SaveFixedModel(name, fm.provider, fm.model, fm.tools || []);
		await loadFixedModels();
	} catch (e) {
		console.error('[FixedModel] Save failed:', e);
	}
}

async function loadFixedModels() {
	try {
		console.log('[FixedModel] Loading from backend...');
		const list = await (window as any).go.main.App.GetFixedModels();
		console.log('[FixedModel] Backend returned:', JSON.stringify(list));
		const map: Record<string, { provider: string; model: string; tools: string[] }> = {};
		// Initialize all known names even if not in DB
		for (const n of FIXED_MODEL_NAMES) map[n] = { provider: '', model: '', tools: [] };
		// Overlay DB values
		for (const m of list || []) {
			if (m.name) {
				console.log(`[FixedModel] DB item "${m.name}" →`, JSON.stringify(m));
				map[m.name] = { provider: m.provider || '', model: m.model || '', tools: m.tools || [] };
			} else {
				console.warn('[FixedModel] DB item without name:', JSON.stringify(m));
			}
		}
		console.log('[FixedModel] Final map keys:', Object.keys(map));
		fixedModels = map;
	} catch (e) {
		console.error('[FixedModel] Load failed:', e);
	}
}

async function loadTools() {
	try {
		const cmds = await (window as any).go.main.App.ListAllCommands();
		availableTools = cmds || [];
	} catch (e) {
		console.error('[FixedModel] Load tools failed:', e);
	}
}

// Load on mount
$effect(() => {
	console.log('[FixedModel] $effect running, providersStore.loaded:', providersStore.loaded);
	if (!providersStore.loaded) providersStore.load();
	loadTools();
	loadFixedModels();
});


	// ── Derived options arrays for Select ──
	const colorThemeOptions = $derived(
		theme.availableColorThemes.map((ct) => ({ value: ct.id, label: ct.name }))
	);

	const iconThemeOptions = $derived(
		theme.availableIconThemes.map((it) => ({ value: it.id, label: it.name }))
	);

	const fontThemeOptions = $derived(
		theme.availableFontThemes.map((ft) => ({ value: ft.id, label: ft.name }))
	);

	const fontSizeOptions = $derived(
		theme.availableFontSizePresets.map((fs) => ({
			value: fs.id,
			label: fs.name,
		}))
	);

	const styleThemeOptions = $derived(
		theme.availableStyleThemes.map((st) => ({
			value: st.id,
			label: st.name,
			description: st.description,
		}))
	);
</script>

<div class="flex flex-col gap-3 p-4 bg-[var(--surface-form)]">

	<!-- ═══════════════════════════════════════════════════════════════
	     Card: System (app-wide theming)
	     ═══════════════════════════════════════════════════════════════ -->
	<div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] overflow-hidden">
		<div class="px-4 py-3 border-b border-[var(--border-primary)] flex items-center gap-2">
			<div class="flex items-center justify-center w-5 h-5 rounded" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
				<Icon name="monitor" size={12} />
			</div>
			<h4 class="text-xs font-semibold tracking-wider" style="color: var(--text-secondary)">SYSTEM</h4>
		</div>
		<div class="divide-y divide-[var(--border-primary)]">

			<!-- Row: Color Theme -->
			<div class="px-4">
				<SettingRow
					label="Color Theme"
					description="Base color palette for the entire interface"
				>
					<ThemedSelect
						value={theme.colorThemeId}
						onValueChange={(v) => theme.setColorTheme(v)}
						options={colorThemeOptions}
						placeholder="Select theme"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Vibrance -->
			<div class="px-4">
				<SettingRow
					label="Vibrance"
					description="Adjust the color saturation intensity"
				>
					<div class="flex items-center gap-2.5">
						<input
							type="range"
							value={theme.vibrance}
							min={0}
							max={200}
							step={5}
							class={cn(
								'w-28 h-1.5 rounded-full appearance-none cursor-pointer',
								'bg-[var(--bg-elevated)]',
								'[&::-webkit-slider-thumb]:appearance-none',
								'[&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3',
								'[&::-webkit-slider-thumb]:rounded-full',
								'[&::-webkit-slider-thumb]:bg-[var(--text-primary)]',
								'[&::-webkit-slider-thumb]:cursor-pointer',
								'[&::-webkit-slider-thumb]:shadow-sm',
								'[&::-webkit-slider-thumb]:transition-transform',
								'[&::-webkit-slider-thumb]:hover:scale-125',
								'[&::-webkit-slider-thumb]:active:scale-110',
								'[&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3',
								'[&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-0',
								'[&::-moz-range-thumb]:bg-[var(--text-primary)]',
								'[&::-moz-range-thumb]:cursor-pointer'
							)}
							oninput={(e) => theme.setVibrance(Number((e.target as HTMLInputElement).value))}
						/>
						<span class="text-[10px] font-mono tabular-nums w-8 text-right" style="color: var(--text-muted)">
							{theme.vibrance}%
						</span>
					</div>
				</SettingRow>
			</div>

			<!-- Row: Icon Theme -->
			<div class="px-4">
				<SettingRow
					label="Icon Theme"
					description="Switch between icon families for the UI"
				>
					<ThemedSelect
						value={theme.iconThemeId}
						onValueChange={(v) => theme.setIconTheme(v)}
						options={iconThemeOptions}
						placeholder="Select icons"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Font Theme -->
			<div class="px-4">
				<SettingRow
					label="Font Theme"
					description="Choose the typeface for the interface"
				>
					<ThemedSelect
						value={theme.fontThemeId}
						onValueChange={(v) => theme.setFontTheme(v)}
						options={fontThemeOptions}
						placeholder="Select font"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Font Size -->
			<div class="px-4">
				<SettingRow
					label="Font Size"
					description="Set the base text size across the app"
				>
					<ThemedSelect
						value={theme.fontSizeId}
						onValueChange={(v) => theme.setFontSize(v)}
						options={fontSizeOptions}
						placeholder="Select size"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Style Theme -->
			<div class="px-4">
				<SettingRow
					label="Style Theme"
					description="Controls visual style: border radius, shadows, and overall shape"
				>
					<ThemedSelect
						value={theme.styleThemeId}
						onValueChange={(v) => theme.setStyleTheme(v)}
						options={styleThemeOptions}
						placeholder="Select style"
						class="w-36"
					/>
				</SettingRow>
			</div>

		</div>
	</div>

	<!-- ═══════════════════════════════════════════════════════════════
  <!-- ═══════════════════════════════════════════════════════════════
       Card: Chat (chat-specific theming)
       ═══════════════════════════════════════════════════════════════ -->
  <div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] overflow-hidden">
    <div class="px-4 py-3 border-b border-[var(--border-primary)] flex items-center gap-2">
      <div class="flex items-center justify-center w-5 h-5 rounded" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
        <Icon name="messageSquare" size={12} />
      </div>
      <h4 class="text-xs font-semibold tracking-wider" style="color: var(--text-secondary)">CHAT</h4>
    </div>
    <div class="divide-y divide-[var(--border-primary)]">
      <!-- Row: Chat Color Theme -->
      <div class="px-4">
        <SettingRow label="Color Theme" description="Color palette exclusive to the chat area">
          <ThemedSelect value={theme.chatColorThemeId} onValueChange={(v) => theme.setChatColorTheme(v)} options={colorThemeOptions} placeholder="Select theme" class="w-36" />
        </SettingRow>
      </div>
      <!-- Row: Chat Font Theme -->
      <div class="px-4">
        <SettingRow label="Font Theme" description="Typeface used only inside the chat area">
          <ThemedSelect value={theme.chatFontThemeId} onValueChange={(v) => theme.setChatFontTheme(v)} options={fontThemeOptions} placeholder="Select font" class="w-36" />
        </SettingRow>
      </div>
      <!-- Row: Chat Font Size -->
      <div class="px-4">
        <SettingRow label="Font Size" description="Text size used only inside the chat area">
          <ThemedSelect value={theme.chatFontSizeId} onValueChange={(v) => theme.setChatFontSize(v)} options={fontSizeOptions} placeholder="Select size" class="w-36" />
        </SettingRow>
      </div>
    </div>
  </div>

  <!-- ═══════════════════════════════════════════════════════════════
       Card: Fixed Models
       ═══════════════════════════════════════════════════════════════ -->
  <div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] overflow-hidden">
    <div class="px-4 py-3 border-b border-[var(--border-primary)] flex items-center gap-2">
      <div class="flex items-center justify-center w-5 h-5 rounded" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
        <Icon name="cpu" size={12} />
      </div>
      <h4 class="text-xs font-semibold tracking-wider" style="color: var(--text-secondary)">FIXED MODELS</h4>
    </div>
    <div class="divide-y divide-[var(--border-primary)]">
      {#each FIXED_MODEL_NAMES as name}
        {@const fm = fixedModels[name]}
        <div class="px-4 py-3 bg-[var(--surface-hover)]/40">
          <div class="flex items-center justify-between cursor-pointer" onclick={() => toggleOpen(name)}>
            <span class="text-[12px] font-medium" style="color: var(--text-primary)">{displayName(name)}</span>
            <div class="flex items-center gap-2 min-w-0">
              {#if fm?.provider}
                <span class="text-[10px] px-1.5 py-0.5 rounded bg-[var(--surface-hover)] truncate max-w-[160px]" style="color: var(--text-muted)">{fm.provider}/{fm.model || '?'}</span>
              {/if}
              <span class="shrink-0"><Icon name={openStates[name] ? 'chevron-up' : 'chevron-down'} size={14} style="color: var(--text-muted)" /></span>
            </div>
          </div>
          {#if openStates[name] && fm}
            <div class="mt-3 flex flex-col gap-3">
              <div class="flex flex-col gap-1 py-2">
                <div class="flex flex-col gap-0.5">
                  <span class="text-[12px] font-medium leading-tight" style="color: var(--text-primary)">Provider</span>
                  <span class="text-[11px] leading-tight" style="color: var(--text-muted)">AI provider for this model</span>
                </div>
                <ThemedSelect value={fm.provider} onValueChange={(v) => { fm.provider = v; fm.model = ''; fixedModels = { ...fixedModels }; }} options={providerOptions} placeholder="Not set" class="w-full" />
              </div>
              <div class="flex flex-col gap-1 py-2">
                <div class="flex flex-col gap-0.5">
                  <span class="text-[12px] font-medium leading-tight" style="color: var(--text-primary)">Model</span>
                  <span class="text-[11px] leading-tight" style="color: var(--text-muted)">Model identifier</span>
                </div>
                <ThemedSelect value={fm.model} onValueChange={(v) => { fm.model = v; fixedModels = { ...fixedModels }; }} options={getModelOptions(fm.provider)} placeholder="Not set" class="w-full" />
              </div>
              <div class="border-t border-[var(--border-primary)] pt-2">
                <div class="flex items-center justify-between cursor-pointer" onclick={() => toggleToolsOpen(name)}>
                  <span class="text-[11px] font-medium" style="color: var(--text-muted)">Tools ({(fm.tools || []).length})</span>
                  <Icon name={toolsOpen[name] ? 'chevron-up' : 'chevron-down'} size={12} style="color: var(--text-muted)" />
                </div>
                {#if toolsOpen[name]}
                  <div class="mt-2 flex flex-wrap gap-1.5">
                    {#each availableTools as tool}
                      <button type="button" onclick={() => toggleTool(name, tool.name)} class={cn(
                        'px-2 py-1 rounded text-[10px] font-medium transition-colors cursor-pointer',
                        (fm.tools || []).includes(tool.name) ? 'bg-[var(--accent-primary)]/15 text-[var(--accent-primary)]' : 'bg-[var(--surface-hover)] text-[var(--text-muted)] hover:bg-[var(--surface-active)]'
                      )}>
                        {tool.name}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>
              <button type="button" onclick={() => saveFixedModel(name)} class="self-end px-3 py-1.5 rounded-lg text-[11px] font-semibold transition-all cursor-pointer hover:brightness-110 active:scale-[0.97]" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
                Save {displayName(name)}
              </button>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>

	<!-- ═══════════════════════════════════════════════════════════════
	     Card: Shortcuts (placeholder)
	     ═══════════════════════════════════════════════════════════════ -->
	<div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] px-4 py-5">
		<div class="flex flex-col items-center gap-3">
			<div
				class="flex items-center justify-center w-10 h-10 rounded-xl opacity-15"
				style="border: 1px solid var(--border-primary)"
			>
				<Icon name="keyboard" size={18} />
			</div>
			<p class="text-xs text-center" style="color: var(--text-faint)">
				Keyboard shortcuts — coming soon
			</p>
		</div>
	</div>
</div>

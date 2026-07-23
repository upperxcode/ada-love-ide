<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { Button } from '$lib/components/ui/button';
	import * as ToggleGroup from '$lib/components/ui/toggle-group';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
		DialogOverlay,
	} from '$lib/components/ui/dialog';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface ModelInfo {
		id: string;
		name: string;
		free: boolean;
		thinking: boolean;
		vision: boolean;
		embedding: boolean;
		tools: boolean;
		installed: boolean;
		context_size: number;
	}

	interface ModelManagerDialogProps {
		open: boolean;
		onClose: () => void;
		providerName: string;
		connectionType: string;
		apiUrl: string;
		apiKeys: string; // JSON string
		currentModels: Record<string, any>;
		onSelect: (selected: Record<string, any>) => void;
	}

	let { open, onClose, providerName, connectionType, apiUrl, apiKeys, currentModels, onSelect }: ModelManagerDialogProps = $props();

	let searchQuery = $state('');
	let activeFilters = $state<string[]>([]);
	let selectedIds = $state<Set<string>>(new Set());
	let availableModels = $state<ModelInfo[]>([]);
	let loading = $state(false);
	// Track editable context_size per model (overrides backend default)
	let contextSizes = $state<Record<string, number>>({});
	let editingContext = $state<string | null>(null);
	let editValue = $state<string>('');

	// Parse currentModels robustly (may be a JSON string or object)
	function resolveCurrentModels(cm: Record<string, any> | string): Record<string, any> {
		if (typeof cm === 'string') {
			try { return JSON.parse(cm || '{}'); } catch { return {}; }
		}
		return cm || {};
	}

	// Load real models from backend when dialog opens (guard prevents re-run loop)
	let wasOpen = $state(false);
	$effect(() => {
		if (open && !wasOpen) {
			wasOpen = true;
			const cm = resolveCurrentModels(currentModels);
			selectedIds = new Set(Object.keys(cm));
			loadModels();
		} else if (!open) {
			wasOpen = false;
		}
	});

	async function loadModels() {
		loading = true;
		try {
			// Extract first key for discovery
			let key = "";
			try {
				const parsedKeys = JSON.parse(apiKeys || "[]");
				key = parsedKeys[0] || "";
			} catch {
				key = "";
			}

			const list = await (window as any).go.main.App.FetchProviderModels(
				providerName,
				connectionType,
				apiUrl,
				key
			);
			
			const cm = resolveCurrentModels(currentModels);
			const sizes: Record<string, number> = {};
			availableModels = list.map((m: any) => {
				// Use saved context_size from currentModels if available, else from backend, else default 128000
				const savedSize = cm[m.id]?.context_size;
				const backendSize = m.context_size;
				const ctxSize = savedSize || backendSize || 128000;
				sizes[m.id] = ctxSize;
				return {
					id: m.id,
					name: m.name,
					free: m.free,
					thinking: m.thinking,
					vision: m.vision,
					embedding: m.embedding,
					tools: m.tools,
					installed: Object.keys(cm).includes(m.id),
					context_size: ctxSize
				};
			});
			contextSizes = sizes;
		} catch (e) {
			toastStore.error('Fetch Models Error', String(e));
		} finally {
			loading = false;
		}
	}

	let filteredModels = $derived.by(() => {
		return availableModels.filter(m => {
			const matchesSearch = m.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
								 m.id.toLowerCase().includes(searchQuery.toLowerCase());
			
			const matchesFilters = activeFilters.every(f => (m as any)[f.toLowerCase()]);
			
			return matchesSearch && matchesFilters;
		});
	});

	function toggleSelectAll() {
		const current = filteredModels;
		if (selectedIds.size === current.length) {
			selectedIds.clear();
		} else {
			current.forEach(m => selectedIds.add(m.id));
		}
		selectedIds = new Set(selectedIds);
	}

	function handleConfirm() {
		const result: Record<string, any> = {};
		availableModels.forEach(m => {
			if (selectedIds.has(m.id)) {
				result[m.id] = {
					free: m.free,
					thinking: m.thinking,
					vision: m.vision,
					embedding: m.embedding,
					tools: m.tools,
					context_size: contextSizes[m.id] ?? 128000
				};
			}
		});
		onSelect(result);
		onClose();
	}

	function startEditContext(modelId: string) {
		editingContext = modelId;
		editValue = String(contextSizes[modelId] ?? 128000);
	}

	function saveEditContext() {
		const id = editingContext;
		if (!id) return;
		const val = parseInt(editValue, 10);
		if (!isNaN(val) && val > 0) {
			contextSizes = { ...contextSizes, [id]: val };
		}
		editingContext = null;
	}

	function handleContextKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			saveEditContext();
			return;
		}
		if (e.key === 'Escape') {
			editingContext = null;
			return;
		}
		// Permitir teclas de navegação/controle
		if (['Backspace', 'Delete', 'Tab', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(e.key)) {
			return;
		}
		// Bloquear qualquer tecla que não seja dígito
		if (!/^\d$/.test(e.key)) {
			e.preventDefault();
		}
	}

	function handleContextInput(e: Event) {
		const input = e.target as HTMLInputElement;
		const digits = input.value.replace(/\D/g, '');
		if (digits !== input.value) {
			input.value = digits;
		}
		editValue = digits;
	}

	const inputBase = 'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<Dialog 
	bind:open
>
	<DialogPortal>
		<DialogOverlay class="z-[60]" />
		<DialogContent 
			class="z-[70] sm:max-w-xl p-0 overflow-hidden flex flex-col bg-[var(--surface-form)] rounded-2xl border border-[var(--border-primary)] shadow-2xl max-h-[80dvh]"
			showCloseButton={false}
			interactOutsideBehavior="ignore"
			escapeKeydownBehavior="ignore"
		>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-[var(--border-primary)] bg-[var(--surface-form)]">
				<div class="flex items-center justify-between">
					<div class="flex flex-col">
						<h3 class="text-base font-bold" style="color: var(--text-primary)">Manage Models</h3>
						<p class="text-xs" style="color: var(--text-muted)">{providerName}</p>
					</div>
					<button type="button" onclick={onClose} class="h-8 w-8 rounded-lg hover:bg-black/10 flex items-center justify-center transition-colors cursor-pointer">
						<Icon name="x" size={16} />
					</button>
				</div>
			</div>

			<!-- Search and Filters -->
			<div class="px-6 py-4 flex flex-col gap-4 border-b border-[var(--border-primary)]">
				<div class="relative">
					<Icon name="search" size={16} class="absolute left-3.5 top-1/2 -translate-y-1/2 text-[var(--text-faint)]" />
					<input 
						type="text" 
						bind:value={searchQuery}
						placeholder="Search models..." 
						class={cn(inputBase, 'w-full pl-10')} 
					/>
				</div>

				<div class="flex items-center gap-3">
					<span class="text-[10px] font-bold uppercase tracking-widest" style="color: var(--text-faint)">Filters:</span>
					<ToggleGroup.Root type="multiple" variant="outline" size="sm" bind:value={activeFilters} spacing={1}>
						<ToggleGroup.Item value="free" class="data-[state=on]:*:[svg]:text-green-500">Free</ToggleGroup.Item>
						<ToggleGroup.Item value="thinking" class="data-[state=on]:*:[svg]:text-[var(--accent-primary)]">Thinking</ToggleGroup.Item>
						<ToggleGroup.Item value="vision" class="data-[state=on]:*:[svg]:text-blue-500">Vision</ToggleGroup.Item>
						<ToggleGroup.Item value="embedding" class="data-[state=on]:*:[svg]:text-yellow-500">Embedding</ToggleGroup.Item>
						<ToggleGroup.Item value="tools" class="data-[state=on]:*:[svg]:text-purple-500">Tool</ToggleGroup.Item>
						<ToggleGroup.Item value="installed" class="data-[state=on]:*:[svg]:text-orange-500">Instalado</ToggleGroup.Item>
					</ToggleGroup.Root>
				</div>
			</div>

			<!-- List -->
			<div class="flex-1 overflow-y-auto p-2 min-h-[300px]">
				{#if loading}
					<div class="flex flex-col items-center justify-center py-20 opacity-50">
						<Icon name="loader" size={32} class="animate-spin mb-4" />
						<p class="text-xs uppercase font-bold tracking-widest">Fetching models from provider...</p>
					</div>
				{:else if filteredModels.length > 0}
					<div class="flex flex-col gap-1">
						{#each filteredModels as model}
							<button 
								type="button"
								onclick={() => {
									if (selectedIds.has(model.id)) selectedIds.delete(model.id);
									else selectedIds.add(model.id);
									selectedIds = new Set(selectedIds);
								}}
								class={cn(
									"flex items-center gap-4 px-4 py-3 rounded-xl border transition-all text-left cursor-pointer",
									selectedIds.has(model.id) 
										? "bg-[var(--accent-primary)]/5 border-[var(--accent-primary)]/30" 
										: "bg-transparent border-transparent hover:bg-[var(--surface-hover)]"
								)}
							>
								<div class={cn(
									"flex h-5 w-5 shrink-0 items-center justify-center rounded border transition-all",
									selectedIds.has(model.id) 
										? "bg-[var(--accent-primary)] border-[var(--accent-primary)] text-white" 
										: "border-[var(--border-primary)]"
								)}>
									{#if selectedIds.has(model.id)}
										<Icon name="check" size={12} />
									{/if}
								</div>

								<div class="flex flex-1 flex-col min-w-0">
									<span class="text-sm font-bold truncate" style="color: var(--text-primary)">{model.name}</span>
									<span class="text-[10px] font-mono truncate" style="color: var(--text-faint)">{model.id}</span>
								</div>

								<div class="flex items-center gap-2">
									{#if model.thinking} <Icon name="brain" size={14} color="var(--accent-primary)" /> {/if}
									{#if model.vision} <Icon name="eye" size={14} color="var(--status-info)" /> {/if}
									{#if model.tools} <Icon name="wrench" size={14} color="var(--text-muted)" /> {/if}
									{#if model.embedding} <Icon name="layers" size={14} color="var(--status-warning)" /> {/if}
									
									<!-- Context size: shows "128K" text, double-click to edit -->
									<div class="flex items-center">
										{#if editingContext === model.id}
											<input
												type="text"
												inputmode="numeric"
												value={editValue}
												oninput={handleContextInput}
												onclick={(e) => e.stopPropagation()}
												onkeydown={handleContextKeydown}
												onblur={saveEditContext}
												autofocus
												class="w-16 rounded border border-[var(--accent-primary)] bg-[var(--surface-input)] px-1.5 py-0.5 text-[10px] font-mono text-center outline-none"
												style="color: var(--text-primary)"
											/>
										{:else}
											<button
												type="button"
												ondblclick={(e) => { e.stopPropagation(); startEditContext(model.id); }}
												class="group/ctx relative px-1.5 py-0.5 rounded text-[10px] font-mono tabular-nums cursor-pointer hover:bg-[var(--surface-hover)] transition-colors"
												style="color: var(--text-faint)"
												title="Double-click to edit context size"
											>
												<span class="tabular-nums">{contextSizes[model.id] ?? 128000}</span>
												<span class="text-[8px] ml-0.5 opacity-50">ctx</span>
											</button>
										{/if}
									</div>

									{#if model.free} <span class="text-[8px] font-bold uppercase px-1.5 py-0.5 rounded bg-[var(--status-success)]/10 text-[var(--status-success)] border border-[var(--status-success)]/20">Free</span> {/if}
								</div>
							</button>
						{/each}
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center py-20 opacity-30 text-center px-10">
						<Icon name="search" size={32} class="mb-4" />
						<p class="text-xs uppercase font-bold tracking-widest">No models found matching your search or filters.</p>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 border-t border-[var(--border-primary)] bg-[var(--surface-form)] flex items-center justify-between">
				<button 
					type="button"
					onclick={toggleSelectAll}
					class="text-[11px] font-bold uppercase tracking-wider hover:underline cursor-pointer" 
					style="color: var(--text-muted)"
				>
					{selectedIds.size === filteredModels.length ? 'Deselect All' : 'Select All'}
				</button>

				<div class="flex items-center gap-3">
					<button
						type="button"
						onclick={onClose}
						class="px-4 py-2 text-[11px] font-bold uppercase tracking-wider hover:bg-black/5 rounded-lg transition-colors cursor-pointer"
						style="color: var(--text-muted)"
					>
						Cancel
					</button>
					<button
						type="button"
						onclick={handleConfirm}
						class="px-6 py-2 bg-[var(--accent-primary)] text-[var(--accent-primary-fg)] rounded-lg text-[11px] font-bold uppercase tracking-wider hover:brightness-110 active:scale-[0.97] transition-all shadow-lg cursor-pointer"
					>
						Load Selected ({selectedIds.size})
					</button>
				</div>
			</div>
		</DialogContent>
	</DialogPortal>
</Dialog>

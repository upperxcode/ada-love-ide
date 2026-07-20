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

	// Load real models from backend when dialog opens
	$effect(() => {
		if (open) {
			const keys = Object.keys(currentModels || {});
			selectedIds = new Set(keys);
			loadModels();
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
			
			availableModels = list.map((m: any) => ({
				id: m.id,
				name: m.name,
				free: m.free,
				thinking: m.thinking,
				vision: m.vision,
				embedding: m.embedding,
				tools: m.tools,
				installed: Object.keys(currentModels || {}).includes(m.id)
			}));
		} catch (e) {
			toastStore.error('Fetch Models Error', String(e));
		} finally {
			loading = false;
		}
	}

	let filteredModels = $derived(() => {
		return availableModels.filter(m => {
			const matchesSearch = m.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
								 m.id.toLowerCase().includes(searchQuery.toLowerCase());
			
			const matchesFilters = activeFilters.every(f => (m as any)[f.toLowerCase()]);
			
			return matchesSearch && matchesFilters;
		});
	});

	function toggleSelectAll() {
		const current = filteredModels();
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
					tools: m.tools
				};
			}
		});
		onSelect(result);
		onClose();
	}

	const inputBase = 'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<Dialog 
	bind:open 
	onOpenChange={(val) => {
		// Prevent closing by clicking outside if val is false
		// But allow explicit onClose calls
		if (!val) {
			// If bits-ui tries to close it (esc or click outside), we ignore it
			// The only way to close is through Cancel or Confirm buttons
			return;
		}
	}}
>
	<DialogPortal>
		<DialogOverlay class="z-[60]" />
		<DialogContent 
			class="z-[70] sm:max-w-xl p-0 overflow-hidden flex flex-col bg-[var(--bg-tertiary)] rounded-2xl border border-[var(--border-primary)] shadow-2xl max-h-[80dvh]"
			showCloseButton={false}
			interactOutsideBehavior="ignore"
			escapeKeydownBehavior="ignore"
		>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-[var(--border-primary)] bg-[var(--surface-elevated)]/50">
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
				{:else if filteredModels().length > 0}
					<div class="flex flex-col gap-1">
						{#each filteredModels() as model}
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
									{#if model.free} <span class="text-[8px] font-bold uppercase px-1.5 py-0.5 rounded bg-green-500/10 text-green-500 border border-green-500/20">Free</span> {/if}
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
			<div class="px-6 py-4 border-t border-[var(--border-primary)] bg-[var(--surface-elevated)]/50 flex items-center justify-between">
				<button 
					type="button"
					onclick={toggleSelectAll}
					class="text-[11px] font-bold uppercase tracking-wider hover:underline cursor-pointer" 
					style="color: var(--text-muted)"
				>
					{selectedIds.size === filteredModels().length ? 'Deselect All' : 'Select All'}
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

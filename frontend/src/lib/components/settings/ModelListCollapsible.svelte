<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { Label } from '$lib/components/ui/label';

	interface ModelListCollapsibleProps {
		value?: string; // JSON Record<string, ModelSettings>
		onOpenManager: () => void;
		onchange?: (value: string) => void;
		label?: string;
	}

	let { value = '{}', onOpenManager, onchange, label }: ModelListCollapsibleProps = $props();

	// Internal state parsed from value
	let models = $state<Record<string, any>>({});
	$effect(() => {
		try {
			models = typeof value === 'string' ? JSON.parse(value || '{}') : (value || {});
		} catch (e) {
			models = {};
		}
	});

	let isOpen = $state(false);
	const modelEntries = $derived(Object.entries(models));

	function toggleProp(modelId: string, prop: string) {
		const updatedModels = { ...models };
		const settings = { ...updatedModels[modelId] };
		settings[prop] = !settings[prop];
		updatedModels[modelId] = settings;
		
		models = updatedModels;
		onchange?.(JSON.stringify(updatedModels));
	}

	function removeModel(modelId: string) {
		const updatedModels = { ...models };
		delete updatedModels[modelId];
		models = updatedModels;
		onchange?.(JSON.stringify(updatedModels));
	}

	function getHealthColor(health: number = 100) {
		if (health >= 80) return 'var(--status-success)'; // Verde
		if (health >= 40) return 'var(--status-warning)'; // Amarelo/Laranja
		return 'var(--status-error)'; // Vermelho
	}

	// Property definitions
	const propDefs = [
		{ id: 'free', icon: 'sparkles', activeColor: 'text-green-500', label: 'Free' },
		{ id: 'vision', icon: 'eye', activeColor: 'text-orange-500', label: 'Vision' },
		{ id: 'tools', icon: 'wrench', activeColor: 'text-blue-500', label: 'Tools' },
		{ id: 'thinking', icon: 'brain', activeColor: 'text-cyan-500', label: 'Thinking' },
		{ id: 'embedding', icon: 'layers', activeColor: 'text-purple-500', label: 'Embedding' },
	];
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center justify-between px-1">
		{#if label}
			<Label class="text-[11px] font-medium" style="color: var(--text-secondary)">{label}</Label>
		{/if}
		<button
			type="button"
			onclick={onOpenManager}
			class="flex items-center gap-1.5 px-2 py-1 rounded-md bg-[var(--accent-primary)]/10 hover:bg-[var(--accent-primary)]/20 transition-colors text-[10px] font-bold uppercase tracking-wider cursor-pointer"
			style="color: var(--accent-primary)"
		>
			<Icon name="settings" size={12} />
			Manage Models
		</button>
	</div>

	{#if modelEntries.length > 0}
		<Collapsible.Root bind:open={isOpen} class="w-full">
					<Collapsible.Trigger>
						{#snippet child({ props: tp })}
						<button
							{...tp}
							class="flex w-full items-center justify-between px-4 py-2.5 rounded-lg border border-[var(--border-primary)] bg-[var(--surface-elevated)] hover:bg-[var(--surface-hover)] transition-colors cursor-pointer"
						>
							<div class="flex items-center gap-2">
								<div class="flex h-5 w-5 items-center justify-center rounded bg-[var(--status-info)]/10 text-[var(--status-info)]">
									<Icon name="cube" size={12} />
								</div>
								<span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--text-secondary)">
									{modelEntries.length} {modelEntries.length === 1 ? 'Model' : 'Models'} supported
								</span>
							</div>
							<Icon 
								name="chevron-down" 
								size={14} 
								class={cn('transition-transform duration-200', isOpen && 'rotate-180')} 
								color="var(--text-faint)"
							/>
						</button>
					{/snippet}
				</Collapsible.Trigger>
			<Collapsible.Content class="pt-2">
				<div class="flex flex-col gap-1.5 pb-1">
					{#each modelEntries as [id, settings] (id)}
						<div class="group flex items-center gap-3 rounded-lg border border-[var(--border-primary)]/50 bg-[var(--bg-secondary)] px-4 py-2.5 hover:border-[var(--border-hover)] transition-all">
							<!-- Health indicator -->
							<div 
								class="w-2 h-2 rounded-full shrink-0 shadow-[0_0_8px_rgba(0,0,0,0.2)]"
								style="background-color: {getHealthColor(settings.health)}"
								title={`Health: ${settings.health ?? 100}%`}
							></div>

							<div class="flex flex-1 flex-col min-w-0">
								<span class="text-[12px] font-bold truncate" style="color: var(--text-primary)">{id}</span>
							</div>
							
							<!-- Interactive Property Icons -->
							<div class="flex items-center gap-3">
								{#each propDefs as p}
									<button
										type="button"
										onclick={() => toggleProp(id, p.id)}
										class={cn(
											"transition-all cursor-pointer hover:scale-110 active:scale-95",
											settings[p.id] ? p.activeColor : "text-[var(--text-faint)] opacity-30 grayscale"
										)}
										title={`Toggle ${p.label}`}
									>
										<Icon name={p.icon} size={14} />
									</button>
								{/each}
							</div>

							<!-- Delete button -->
							<button
								type="button"
								onclick={() => removeModel(id)}
								class="opacity-0 group-hover:opacity-100 flex h-7 w-7 items-center justify-center rounded-md hover:bg-[var(--status-error)]/10 transition-all cursor-pointer ml-1"
								style="color: var(--status-error)"
								title="Remove model"
							>
								<Icon name="trash-2" size={14} />
							</button>
						</div>
					{/each}
				</div>
			</Collapsible.Content>
		</Collapsible.Root>
	{:else}
		<div class="flex flex-col items-center justify-center py-6 border border-dashed border-[var(--border-primary)] rounded-lg opacity-40">
			<Icon name="cube" size={24} class="mb-2" />
			<p class="text-[10px] uppercase font-bold tracking-widest">No models configured</p>
		</div>
	{/if}
</div>

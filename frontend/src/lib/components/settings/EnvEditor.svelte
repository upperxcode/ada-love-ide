<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { Label } from '$lib/components/ui/label';

	interface EnvEditorProps {
		value?: string; // JSON string
		onchange?: (value: string) => void;
		label?: string;
	}

	let { value = '{}', onchange, label }: EnvEditorProps = $props();

	// Parse JSON value to internal state
	let env = $state<Record<string, string>>({});
	$effect(() => {
		try {
			env = JSON.parse(value || '{}');
		} catch (e) {
			env = {};
		}
	});

	// New item state
	let newKey = $state('');
	let newVal = $state('');

	function addItem() {
		if (!newKey || !newVal) return;
		const updated = { ...env, [newKey]: newVal };
		env = updated;
		onchange?.(JSON.stringify(updated));
		newKey = '';
		newVal = '';
	}

	function removeItem(key: string) {
		const updated = { ...env };
		delete updated[key];
		env = updated;
		onchange?.(JSON.stringify(updated));
	}

	let isOpen = $state(true); // Default open for better UX
	const keys = $derived(Object.keys(env));

	const inputBase = 'rounded-lg px-4 py-2.5 text-[13px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<div class="flex flex-col gap-3">
	{#if label}
		<Label class="text-[11px] font-medium block mb-0.5" style="color: var(--text-secondary)">{label}</Label>
	{/if}

	<!-- Add form -->
	<div class="flex items-end gap-2">
		<div class="flex flex-1 flex-col gap-1.5">
			<label class="text-[9px] uppercase font-bold tracking-widest px-1" style="color: var(--text-faint)">Key</label>
			<input
				type="text"
				bind:value={newKey}
				placeholder="e.g. API_KEY"
				class={cn(inputBase, 'w-full')}
				style="color: var(--text-primary)"
			/>
		</div>
		<div class="flex flex-[1.5] flex-col gap-1.5">
			<label class="text-[9px] uppercase font-bold tracking-widest px-1" style="color: var(--text-faint)">Value</label>
			<input
				type="text"
				bind:value={newVal}
				placeholder="secret_value_123"
				class={cn(inputBase, 'w-full')}
				style="color: var(--text-primary)"
			/>
		</div>
		<button
			type="button"
			disabled={!newKey || !newVal}
			onclick={addItem}
			class={cn(
				'h-11 px-4 rounded-lg flex items-center justify-center transition-all',
				!newKey || !newVal 
					? 'opacity-30 cursor-not-allowed grayscale bg-[var(--surface-input)] border border-[var(--border-primary)]' 
					: 'cursor-pointer bg-[var(--accent-primary)] hover:brightness-110 active:scale-95'
			)}
			style={!newKey || !newVal ? '' : 'color: var(--accent-primary-fg)'}
		>
			<Icon name="plus" size={18} />
		</button>
	</div>

	<!-- Collapsible list -->
	{#if keys.length > 0}
		<Collapsible.Root bind:open={isOpen} class="w-full">
			<Collapsible.Trigger>
				{#snippet child({ props: tp })}
					<button
						{...tp}
						class="flex w-full items-center justify-between px-4 py-2.5 rounded-lg border border-[var(--border-primary)] bg-[var(--surface-elevated)] hover:bg-[var(--surface-hover)] transition-colors cursor-pointer"
					>
						<div class="flex items-center gap-2">
							<div class="flex h-5 w-5 items-center justify-center rounded bg-[var(--accent-primary)]/10 text-[var(--accent-primary)]">
								<Icon name="log" size={12} />
							</div>
							<span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--text-secondary)">
								{keys.length} Environment {keys.length === 1 ? 'variable' : 'variables'}
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
					{#each keys as key}
						<div class="group flex items-center gap-3 rounded-lg border border-[var(--border-primary)]/50 bg-[var(--bg-secondary)] px-4 py-2.5 hover:border-[var(--border-hover)] transition-all">
							<div class="flex flex-1 flex-col min-w-0">
								<span class="text-[10px] font-mono font-bold uppercase truncate" style="color: var(--accent-primary)">{key}</span>
								<span class="text-[12px] font-mono truncate" style="color: var(--text-muted)">{env[key]}</span>
							</div>
							<button
								type="button"
								onclick={() => removeItem(key)}
								class="opacity-0 group-hover:opacity-100 flex h-8 w-8 items-center justify-center rounded-md hover:bg-[var(--status-error)]/10 transition-all cursor-pointer"
								style="color: var(--status-error)"
								title="Remove {key}"
							>
								<Icon name="trash-2" size={14} />
							</button>
						</div>
					{/each}
				</div>
			</Collapsible.Content>
		</Collapsible.Root>
	{/if}
</div>

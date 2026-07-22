<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { Label } from '$lib/components/ui/label';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface APIKeyManagerProps {
		value?: string; // JSON string of string[] or raw string[]
		onchange?: (value: string) => void;
		label?: string;
		providerName?: string;
		apiUrl?: string;
		connectionType?: string;
	}

	let { value = '[]', onchange, label, providerName, apiUrl, connectionType = 'openai' }: APIKeyManagerProps = $props();

	// Parse keys to internal state
	let keys = $derived.by(() => {
		try {
			if (typeof value === 'string') {
				const parsed = JSON.parse(value || '[]');
				return Array.isArray(parsed) ? parsed : [];
			} else if (Array.isArray(value)) {
				return value;
			}
			return [];
		} catch {
			return [];
		}
	});

	// Connection test state per key value hash (stable across add/remove)
	let testingStates = $state<Record<string, 'idle' | 'loading' | 'success' | 'error'>>({});

	function keyHash(k: string): string {
		let h = 0;
		for (let i = 0; i < k.length; i++) {
			h = ((h << 5) - h) + k.charCodeAt(i);
			h |= 0;
		}
		return String(h);
	}

	// New key state
	let newKey = $state('');

	function addKey() {
		if (!newKey.trim()) return;
		const updated = [...keys, newKey.trim()];
		onchange?.(JSON.stringify(updated));
		newKey = '';
	}

	function removeKey(index: number) {
		const key = keys[index];
		const updated = keys.filter((_, i) => i !== index);
		onchange?.(JSON.stringify(updated));
		if (key) {
			const newStates = { ...testingStates };
			delete newStates[keyHash(key)];
			testingStates = newStates;
		}
	}

	async function testKey(index: number) {
		const key = keys[index];
		if (!key) return;
		const hash = keyHash(key);

		testingStates = { ...testingStates, [hash]: 'loading' };
		try {
			const result = await (window as any).go.main.App.TestProviderConnection(
				providerName || 'custom',
				connectionType,
				apiUrl || '',
				key
			);

			if (result.success) {
				testingStates = { ...testingStates, [hash]: 'success' };
				toastStore.success('Key Valid', 'Connection established successfully');
			} else {
				testingStates = { ...testingStates, [hash]: 'error' };
				toastStore.error('Key Invalid', result.message);
			}
		} catch (e) {
			testingStates = { ...testingStates, [hash]: 'error' };
			toastStore.error('Test Error', String(e));
		}
	}

	let isOpen = $state(true);

	const inputBase = 'rounded-lg px-4 py-2.5 text-[13px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<div class="flex flex-col gap-3">
	{#if label}
		<Label class="text-[11px] font-medium block mb-0.5" style="color: var(--text-secondary)">{label}</Label>
	{/if}

	<!-- Add form -->
	<div class="flex items-end gap-2">
		<div class="flex flex-1 flex-col gap-1.5">
			<!-- svelte-ignore a11y_label_has_associated_control -->
			<label class="text-[9px] uppercase font-bold tracking-widest px-1" style="color: var(--text-faint)">New API Key</label>
			<div class="relative flex items-center">
				<input
					type="password"
					bind:value={newKey}
					placeholder="sk-..."
					class={cn(inputBase, 'w-full pr-10')}
					style="color: var(--text-primary)"
				/>
				<div class="absolute right-3 text-[var(--text-faint)]">
					<Icon name="key" size={14} />
				</div>
			</div>
		</div>
		<button
			type="button"
			disabled={!newKey.trim()}
			onclick={addKey}
			class={cn(
				'h-11 px-4 rounded-lg flex items-center justify-center transition-all',
				!newKey.trim()
					? 'opacity-30 cursor-not-allowed grayscale bg-[var(--surface-input)] border border-[var(--border-primary)]' 
					: 'cursor-pointer bg-[var(--accent-primary)] hover:brightness-110 active:scale-95'
			)}
			style={!newKey.trim() ? '' : 'color: var(--accent-primary-fg)'}
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
								<Icon name="key" size={12} />
							</div>
							<span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--text-secondary)">
								{keys.length} API {keys.length === 1 ? 'key' : 'keys'} configured
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
					{#each keys as key, index}
						<div class="group flex items-center gap-3 rounded-lg border border-[var(--border-primary)]/50 bg-[var(--bg-secondary)] px-4 py-2.5 hover:border-[var(--border-hover)] transition-all">
							<div class="flex flex-1 items-center gap-2 min-w-0">
								<Icon name="key" size={12} color="var(--accent-primary)" />
								<span class="text-[12px] font-mono truncate" style="color: var(--text-muted)">
									{key.slice(0, 8)}••••••••{key.slice(-4)}
								</span>
							</div>

							<!-- Test Icon button -->
							<button
								type="button"
								onclick={() => testKey(index)}
								disabled={testingStates[keyHash(key)] === 'loading'}
								class={cn(
									"flex h-8 w-8 items-center justify-center rounded-md transition-all cursor-pointer",
									testingStates[keyHash(key)] === 'success' ? "bg-[var(--status-success)]/10 text-[var(--status-success)]" :
									testingStates[keyHash(key)] === 'error' ? "bg-red-500/10 text-red-500" :
									"hover:bg-[var(--surface-hover)] text-[var(--text-faint)] hover:text-[var(--text-primary)]"
								)}
								title="Test this key"
							>
								{#if testingStates[keyHash(key)] === 'loading'}
									<Icon name="loader" size={14} class="animate-spin" />
								{:else if testingStates[keyHash(key)] === 'success'}
									<Icon name="check" size={14} />
								{:else if testingStates[keyHash(key)] === 'error'}
									<Icon name="x" size={14} />
								{:else}
									<Icon name="send" size={14} />
								{/if}
							</button>

							<button
								type="button"
								onclick={() => removeKey(index)}
								class="opacity-0 group-hover:opacity-100 flex h-8 w-8 items-center justify-center rounded-md hover:bg-[var(--status-error)]/10 transition-all cursor-pointer"
								style="color: var(--status-error)"
								title="Remove key"
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

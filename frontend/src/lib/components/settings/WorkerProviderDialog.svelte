<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { Button } from '$lib/components/ui/button';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
		DialogOverlay,
	} from '$lib/components/ui/dialog';
	import SettingRow from './SettingRow.svelte';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface ModelInfo {
		id: string;
		name: string;
		free: boolean;
		thinking: boolean;
		vision: boolean;
		embedding: boolean;
		tools: boolean;
		context_size: number;
	}

	interface WorkerProviderDialogProps {
		open: boolean;
		workerName: string;
		onClose: () => void;
		onSave: (providerName: string, config: Record<string, any>) => void;
	}

	let { open, workerName, onClose, onSave }: WorkerProviderDialogProps = $props();

	let providerName = $state('');
	let apiUrl = $state('');
	let connectionType = $state('openai');
	let strategy = $state('');

	let availableModels = $state<ModelInfo[]>([]);
	let selectedModelIds = $state<Set<string>>(new Set());
	let loading = $state(false);
	let fetchingModels = $state(false);
	let step: 'config' | 'models' = $state('config');

	const connectionTypes = [
		{ label: 'OpenAI Compatible', value: 'openai' },
		{ label: 'Anthropic', value: 'anthropic' },
		{ label: 'Ollama', value: 'ollama' },
		{ label: 'Custom', value: 'custom' },
	];

	const strategyOptions = [
		{ label: 'None', value: '' },
		{ label: 'Simple Rotate (Round Robin)', value: 'simple_rotate' },
		{ label: 'Hard Caps (Quota-based)', value: 'hard_caps' },
		{ label: 'Load Balancing (Weighted)', value: 'load_balancing' },
	];

	async function handleFetchModels() {
		if (!apiUrl) {
			toastStore.error('Validation Error', 'API URL is required');
			return;
		}

		fetchingModels = true;
		try {
			const app = (window as any).go?.main?.App;
			const list = await app.FetchProviderModels(
				providerName || 'provider',
				connectionType,
				apiUrl,
				'' // No API key needed for worker providers
			);

			availableModels = (list || []).map((m: any) => ({
				id: m.id,
				name: m.name,
				free: m.free || false,
				thinking: m.thinking || false,
				vision: m.vision || false,
				embedding: m.embedding || false,
				tools: m.tools || false,
				context_size: m.context_size || 128000,
			}));
			step = 'models';
		} catch (e) {
			toastStore.error('Fetch Models Error', String(e));
		} finally {
			fetchingModels = false;
		}
	}

	function toggleModel(id: string) {
		const next = new Set(selectedModelIds);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		selectedModelIds = next;
	}

	function toggleSelectAll() {
		if (selectedModelIds.size === availableModels.length) {
			selectedModelIds = new Set();
		} else {
			selectedModelIds = new Set(availableModels.map(m => m.id));
		}
	}

	function handleConfirm() {
		const models: Record<string, any> = {};
		availableModels.forEach(m => {
			if (selectedModelIds.has(m.id)) {
				models[m.id] = {
					free: m.free,
					thinking: m.thinking,
					vision: m.vision,
					embedding: m.embedding,
					tools: m.tools,
					context_size: m.context_size,
				};
			}
		});

		const providerConfig = {
			api_url: apiUrl,
			type_connection: connectionType,
			strategy: strategy,
			icon: '🔌',
			color: '#6366f1',
			models,
		};

		onSave(providerName || 'worker-provider', providerConfig);
		onClose();
	}

	const inputBase = 'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<Dialog bind:open>
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
						<h3 class="text-base font-bold" style="color: var(--text-primary)">
							{step === 'models' ? 'Select Models' : 'Configure Provider'}
						</h3>
						<p class="text-xs" style="color: var(--text-muted)">Worker: {workerName}</p>
					</div>
					<button type="button" onclick={onClose} class="h-8 w-8 rounded-lg hover:bg-black/10 flex items-center justify-center transition-colors cursor-pointer">
						<Icon name="x" size={16} />
					</button>
				</div>
			</div>

			{#if step === 'config'}
				<!-- Step 1: Provider Configuration -->
				<div class="flex-1 overflow-y-auto p-6">
					<div class="flex flex-col gap-4">
						<div>
							<label class="text-[11px] font-medium mb-1.5 block" style="color: var(--text-secondary)">Provider Name</label>
							<input
								type="text"
								bind:value={providerName}
								placeholder="e.g., my-worker-provider"
								class={cn(inputBase, 'w-full')}
							/>
						</div>

						<div>
							<label class="text-[11px] font-medium mb-1.5 block" style="color: var(--text-secondary)">API URL</label>
							<input
								type="text"
								bind:value={apiUrl}
								placeholder="https://api.example.com/v1"
								class={cn(inputBase, 'w-full')}
							/>
						</div>

						<div>
							<label class="text-[11px] font-medium mb-1.5 block" style="color: var(--text-secondary)">Connection Type</label>
							<select
								bind:value={connectionType}
								class={cn(inputBase, 'w-full cursor-pointer')}
							>
								{#each connectionTypes as opt}
									<option value={opt.value}>{opt.label}</option>
								{/each}
							</select>
						</div>

						<div>
							<label class="text-[11px] font-medium mb-1.5 block" style="color: var(--text-secondary)">Strategy (optional)</label>
							<select
								bind:value={strategy}
								class={cn(inputBase, 'w-full cursor-pointer')}
							>
								{#each strategyOptions as opt}
									<option value={opt.value}>{opt.label}</option>
								{/each}
							</select>
						</div>
					</div>
				</div>

				<!-- Footer -->
				<div class="px-6 py-4 border-t border-[var(--border-primary)] bg-[var(--surface-form)] flex items-center justify-end gap-3">
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
						onclick={handleFetchModels}
						disabled={fetchingModels || !apiUrl}
						class="px-6 py-2 bg-[var(--accent-primary)] text-[var(--accent-primary-fg)] rounded-lg text-[11px] font-bold uppercase tracking-wider hover:brightness-110 active:scale-[0.97] transition-all shadow-lg cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2"
					>
						{#if fetchingModels}
							<Icon name="loader" size={14} class="animate-spin" />
							Fetching Models...
						{:else}
							<Icon name="search" size={14} />
							Fetch Models
						{/if}
					</button>
				</div>

			{:else if step === 'models'}
				<!-- Step 2: Model Selection -->
				<div class="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
					<button
						type="button"
						onclick={() => { step = 'config'; }}
						class="flex items-center gap-1 text-[11px] font-medium hover:underline cursor-pointer"
						style="color: var(--text-muted)"
					>
						<Icon name="chevron-left" size={14} />
						Back to config
					</button>
					<span class="text-[11px]" style="color: var(--text-faint)">
						{selectedModelIds.size} of {availableModels.length} selected
					</span>
				</div>

				<div class="flex-1 overflow-y-auto p-2 min-h-[200px]">
					{#if availableModels.length > 0}
						<div class="flex flex-col gap-1">
							{#each availableModels as model}
								<button
									type="button"
									onclick={() => toggleModel(model.id)}
									class={cn(
										"flex items-center gap-4 px-4 py-3 rounded-xl border transition-all text-left cursor-pointer",
										selectedModelIds.has(model.id)
											? "bg-[var(--accent-primary)]/5 border-[var(--accent-primary)]/30"
											: "bg-transparent border-transparent hover:bg-[var(--surface-hover)]"
									)}
								>
									<div class={cn(
										"flex h-5 w-5 shrink-0 items-center justify-center rounded border transition-all",
										selectedModelIds.has(model.id)
											? "bg-[var(--accent-primary)] border-[var(--accent-primary)] text-white"
											: "border-[var(--border-primary)]"
									)}>
										{#if selectedModelIds.has(model.id)}
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
										{#if model.free} <span class="text-[8px] font-bold uppercase px-1.5 py-0.5 rounded bg-[var(--status-success)]/10 text-[var(--status-success)] border border-[var(--status-success)]/20">Free</span> {/if}
									</div>
								</button>
							{/each}
						</div>
					{:else}
						<div class="flex flex-col items-center justify-center py-16 opacity-40">
							<Icon name="cube" size={32} class="mb-3" />
							<p class="text-[10px] uppercase font-bold tracking-widest">No models available</p>
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
						{selectedModelIds.size === availableModels.length ? 'Deselect All' : 'Select All'}
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
							Save {selectedModelIds.size} Models
						</button>
					</div>
				</div>
			{/if}
		</DialogContent>
	</DialogPortal>
</Dialog>
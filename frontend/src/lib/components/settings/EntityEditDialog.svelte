<script lang="ts">
	import { cn } from '$lib/utils';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
	} from '$lib/components/ui/dialog';
	import { Switch } from '$lib/components/ui/switch';
	import { Icon } from '$lib/components/icon';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import ExpandableTextarea from '$lib/components/ui/ExpandableTextarea.svelte';
	import SettingRow from './SettingRow.svelte';
	import EntityHeader from './EntityHeader.svelte';
	import EnvEditor from './EnvEditor.svelte';
	import APIKeyManager from './APIKeyManager.svelte';
	import ModelListCollapsible from './ModelListCollapsible.svelte';
	import ModelManagerDialog from './ModelManagerDialog.svelte';
	import { providersStore } from '$lib/stores/providers.svelte';
	import { toastStore } from '$lib/stores/toast.svelte';

	export interface FieldConfig {
		key: string;
		label: string;
		description?: string;
		type: 'text' | 'textarea' | 'number' | 'select' | 'toggle' | 'color' | 'tags' | 'provider_select' | 'model_select';
		placeholder?: string;
		options?: { label: string; value: string }[];
		required?: boolean;
		min?: number;
		max?: number;
		step?: number;
		decimals?: boolean;
		expandable?: boolean;
		fullWidth?: boolean;
	}

	interface EntityEditDialogProps {
		open: boolean;
		onOpenChange: (open: boolean) => void;
		entity: Record<string, any> | null;
		entityType: string;
		fields: FieldConfig[];
		onSave: (data: Record<string, any>) => void;
	}

	let {
		open = $bindable(),
		onOpenChange,
		entity,
		entityType,
		fields,
		onSave,
	}: EntityEditDialogProps = $props();

	let formData = $state<Record<string, any>>({});
	let testing = $state(false);
	let managerOpen = $state(false);

	// ── Derived header state ──
	let headerIcon = $derived(formData.icon ?? '📄');
	let headerColor = $derived(formData.color ?? '#3f3f46');

	// ── Derived: available models for selected provider (cascade) ──
	let availableModels = $derived.by(() => {
		const provider = formData.provider;
		if (!provider) return [];
		return providersStore.getModels(provider).map((m) => ({
			value: m.name,
			label: m.name,
		}));
	});

	// ── Load providers when dialog opens ──
	$effect(() => {
		if (open) {
			providersStore.load();
		}
	});

	// ── When provider changes, clear model if it doesn't belong to the new provider ──
	$effect(() => {
		if (open && formData.provider && formData.model) {
			const models = providersStore.getModels(formData.provider);
			// Only clear if the model is NOT in the new list AND the list is NOT empty 
			// (empty might mean it's still loading)
			if (models.length > 0 && !models.some((m) => m.name === formData.model)) {
				formData = { ...formData, model: '' };
			}
		}
	});

	// Reset form when dialog opens with new entity data
	$effect(() => {
		if (open) {
			if (entity) {
				formData = { ...entity };
			} else {
				const defaults: Record<string, any> = {};
				for (const field of fields) {
					if (field.type === 'number') {
						defaults[field.key] = field.min ?? '';
					} else if (field.type === 'toggle') {
						defaults[field.key] = false;
					} else {
						defaults[field.key] = '';
					}
				}
				// Ensure icon/color defaults exist even though they're not in fields
				defaults.icon = '📄';
				defaults.color = '#3f3f46';
				formData = defaults;
			}
		}
	});

	function handleSave() {
		onSave(formData);
	}

	async function handleTestConnection() {
		if (!formData.nome && !formData.command && !formData.url) return;
		testing = true;
		try {
			let args: string[] = [];
			if (typeof formData.arguments === 'string' && formData.arguments.trim()) {
				try {
					const parsed = JSON.parse(formData.arguments);
					args = Array.isArray(parsed) ? parsed : [formData.arguments];
				} catch {
					args = formData.arguments.split(' ').filter(Boolean);
				}
			}

			const result = await (window as any).go.main.App.TestMCPConnection(
				formData.nome || 'Test',
				formData.command || '',
				formData.url || '',
				args
			);

			if (result.success) {
				toastStore.success('Connection Successful', `${result.message} (${result.latency_ms}ms)`);
			} else {
				toastStore.error('Connection Failed', result.message);
			}
		} catch (e) {
			toastStore.error('Test Error', String(e));
		} finally {
			testing = false;
		}
	}

	// ── Visibility helper ──
	function isFieldVisible(key: string): boolean {
		if (entityType !== 'MCP') return true;
		
		const connectType = formData.connect_type || 'stdio';
		
		if (connectType === 'stdio') {
			// Hide URL, Timeout, OAuth for stdio
			return !['url', 'timeout', 'oauth_client_id'].includes(key);
		} else if (connectType === 'sse') {
			// Hide Command, Arguments for sse
			return !['command', 'arguments'].includes(key);
		}
		
		return true;
	}

	function updateField(key: string, value: any) {
		formData = { ...formData, [key]: value };
	}

	// ── Provider select options (from backend) ──
	let providerOptions = $derived(
		providersStore.providers.map((p) => ({ value: p.name, label: p.name }))
	);

	// ── Input classes helper ──
	const inputBase = 'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<Dialog bind:open onOpenChange={onOpenChange}>
	<DialogPortal>
		<DialogContent class="sm:max-w-[610px] max-h-[85dvh] flex flex-col p-0 gap-0 overflow-hidden bg-[var(--surface-form)]" showCloseButton={false}>
			<!-- ── Custom header with color bar + icon/color pickers ── -->
			<EntityHeader
				icon={headerIcon}
				color={headerColor}
				entityType={entityType}
				isNew={!entity}
				onIconChange={(emoji) => updateField('icon', emoji)}
				onColorChange={(c) => updateField('color', c)}
				onClose={() => onOpenChange(false)}
			/>

			<!-- ── Form (row-based layout) ── -->
			<div class="flex-1 overflow-y-auto bg-[var(--surface-form)]">
				{#each fields.filter(f => isFieldVisible(f.key)) as field (field.key)}
					<!-- ── Spacer top ── -->
					<div class="px-5 w-full h-[10px]"></div>
					<!-- Full-width fields: label on top, component below -->
					{#if field.fullWidth}
						<div class="px-5 pb-3">
							{#if field.key === 'environment'}
								<EnvEditor
									label={field.label}
									value={formData[field.key] ?? '{}'}
									onchange={(v) => updateField(field.key, v)}
								/>
								{:else if field.key === 'api_keys'}
									<APIKeyManager
										label={field.label}
										value={formData[field.key] ?? '[]'}
										onchange={(v) => updateField(field.key, v)}
										providerName={formData.name}
										apiUrl={formData.api_url}
										connectionType={formData.type_connection}
									/>
								{:else if field.key === 'models'}
									<ModelListCollapsible
										label={field.label}
										value={formData[field.key] ?? '{}'}
										onOpenManager={() => (managerOpen = true)}
										onchange={(v) => updateField(field.key, v)}
									/>
							{:else}
								<ExpandableTextarea
									id="field-{field.key}"
									label={field.label}
									value={formData[field.key] ?? ''}
									oninput={(v) => updateField(field.key, v)}
									placeholder={field.placeholder}
									minHeight={80}
									class="w-full"
									textareaClass="w-full"
								/>
							{/if}
						</div>

					<!-- Row-based fields: label left, input right -->
					{:else}
						<div class="divide-y divide-[var(--border-primary)]">
							<div class="px-5">
								<SettingRow label={field.label} description={field.description}>
									{#if field.type === 'provider_select'}
										<ThemedSelect
											value={formData[field.key] ?? ''}
											onValueChange={(v) => updateField(field.key, v)}
											options={providerOptions}
											placeholder="Select provider"
											class="w-52"
										/>

									{:else if field.type === 'model_select'}
										<ThemedSelect
											value={formData[field.key] ?? ''}
											onValueChange={(v) => updateField(field.key, v)}
											options={availableModels}
											placeholder="Select model"
											class="w-52"
											disabled={!formData.provider}
										/>

									{:else if field.type === 'select'}
										<ThemedSelect
											value={formData[field.key] ?? ''}
											onValueChange={(v) => updateField(field.key, v)}
											options={field.options ?? []}
											placeholder={field.placeholder || 'Select...'}
											class="w-52"
										/>

									{:else if field.type === 'toggle'}
										<Switch
											checked={!!formData[field.key]}
											onCheckedChange={(v) => updateField(field.key, v)}
										/>

									{:else if field.type === 'number'}
										<input
											type="text"
											inputmode={field.decimals ? 'decimal' : 'numeric'}
											value={formData[field.key] ?? ''}
											onkeydown={(e) => {
												// Allow: backspace, delete, tab, escape, enter
												if (['Backspace', 'Delete', 'Tab', 'Escape', 'Enter', 'ArrowLeft', 'ArrowRight'].includes(e.key)) return;

												// Logic for decimals (allow one . or ,)
												if (field.decimals) {
													if (e.key === '.' || e.key === ',') {
														const val = String(formData[field.key] ?? '');
														if (val.includes('.')) e.preventDefault();
														return;
													}
												}

												// Allow digits
												if (/^\d$/.test(e.key)) return;

												// Block everything else
												e.preventDefault();
											}}
											oninput={(e) => {
												let raw = (e.target as HTMLInputElement).value;
												// Replace comma with dot
												raw = raw.replace(',', '.');

												// Final cleanup to ensure only one dot
												const parts = raw.split('.');
												if (parts.length > 2) {
													raw = parts[0] + '.' + parts.slice(1).join('');
												}

												// Update field
												if (raw === '') {
													updateField(field.key, 0);
												} else {
													const num = Number(raw);
													if (!isNaN(num)) {
														updateField(field.key, num);
													}
												}
											}}
											placeholder={field.placeholder}
											class={cn(inputBase, 'w-24 text-right font-mono truncate')}
											style="color: var(--text-primary)"
										/>

									{:else if field.type === 'textarea'}
										{#if field.expandable}
											<ExpandableTextarea
												id="field-{field.key}"
												value={formData[field.key] ?? ''}
												oninput={(v) => updateField(field.key, v)}
												placeholder={field.placeholder}
												minHeight={32}
												textareaClass="w-full"
											/>
										{:else}
											<textarea
												value={formData[field.key] ?? ''}
												oninput={(e) => updateField(field.key, (e.target as HTMLTextAreaElement).value)}
												placeholder={field.placeholder}
												rows={3}
												class={cn(inputBase, 'w-full resize-none')}
												style="color: var(--text-primary)"
											></textarea>
										{/if}

									{:else}
										<input
											type="text"
											value={formData[field.key] ?? ''}
											oninput={(e) => updateField(field.key, (e.target as HTMLInputElement).value)}
											placeholder={field.placeholder}
											class={cn(inputBase, 'w-full truncate')}
											style="color: var(--text-primary)"
										/>
									{/if}
								</SettingRow>
							</div>
						</div>
					{/if}
				{/each}
			<!-- ── Spacer bottom ── -->
			<div class="px-5 w-full h-[10px]"></div>
			</div>

			<!-- ── Footer actions ── -->
			<div class="flex items-center justify-between px-5 py-3 border-t border-[var(--border-primary)] bg-[var(--surface-form)]">
				<div class="flex items-center gap-2">
					{#if entityType === 'MCP'}
						<button
							type="button"
							onclick={handleTestConnection}
							disabled={testing}
							class={cn(
								'flex items-center gap-2 px-4 py-2 rounded-lg border border-[var(--border-primary)] bg-[var(--surface-input)]',
								'text-[11px] font-medium cursor-pointer transition-all',
								'hover:bg-[var(--surface-hover)] active:scale-[0.97] disabled:opacity-50 disabled:cursor-not-allowed'
							)}
							style="color: var(--text-primary)"
						>
							{#if testing}
								<Icon name="loader" size={13} class="animate-spin" />
								Testing...
							{:else}
								<Icon name="send" size={13} />
								Test Connection
							{/if}
						</button>
					{/if}
				</div>

				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={() => onOpenChange(false)}
						class={cn(
							'flex items-center px-4 py-2 rounded-lg',
							'text-[11px] font-medium cursor-pointer transition-colors',
							'hover:bg-[var(--surface-hover)]'
						)}
						style="color: var(--text-muted)"
					>
						Cancel
					</button>

					<button
						type="button"
						onclick={handleSave}
						class={cn(
							'flex items-center px-4 py-2 rounded-lg',
							'text-[11px] font-semibold cursor-pointer transition-all',
							'hover:brightness-110 active:scale-[0.97]'
						)}
						style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
					>
						{entity ? 'Save' : 'Create'}
					</button>
				</div>
			</div>
		</DialogContent>
	</DialogPortal>
</Dialog>

	<ModelManagerDialog
		open={managerOpen}
		onClose={() => (managerOpen = false)}
		providerName={formData.name || 'Provider'}
		connectionType={formData.type_connection || 'openai'}
		apiUrl={formData.api_url || ''}
		apiKeys={formData.api_keys || '[]'}
		currentModels={formData.models || {}}
		onSelect={(newModels) => updateField('models', newModels)}
	/>

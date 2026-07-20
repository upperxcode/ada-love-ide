<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		Dialog,
		DialogPortal,
		DialogContent,
		DialogOverlay,
	} from '$lib/components/ui/dialog';
	import { Switch } from '$lib/components/ui/switch';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import ExpandableTextarea from '$lib/components/ui/ExpandableTextarea.svelte';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import EntityHeader from './EntityHeader.svelte';
	import { toastStore } from '$lib/stores/toast.svelte';

	interface SpecWizardDialogProps {
		open: boolean;
		onOpenChange: (open: boolean) => void;
		entity: Record<string, any> | null;
		onSave: (data: Record<string, any>) => void;
	}

	let { open = $bindable(), onOpenChange, entity, onSave }: SpecWizardDialogProps = $props();

	// ── Wizard State ──
	let currentStep = $state(1);
	let formData = $state<Record<string, any>>({
		name: '',
		expert_language_plugin: '',
		prd: '',
		functional_requirements: [],
		non_functional_requirements: [],
		architecture: '',
		persistence: '',
		engineering_philosophies: [],
		design_patterns: [],
		data_patterns: [],
		stack_plugin: '',
		dependency_manifest: [],
		business: {
			state_management: '',
			api_contract: '',
			customization_details: '',
			final_adjustments: '',
			architecture_recommendations: '',
		},
		architecture_health: 100,
		color: '#3b82f6',
		icon: '📝',
	});

	let expertPlugins = $state<any[]>([]);
	let architectures = $state<any[]>([]);
	let stacks = $state<any[]>([]);
	let stateOptions = $state<any[]>([]);
	let domainOpen = $state(true);
	let loadingOptions = $state(false);
	let isLoadingOptions = $state(false);

	let newLibName = $state('');
	let newLibVersion = $state('');
	let newLibMandatory = $state(false);

	// Initialize when opening
	let isInitialized = $state(false);

	$effect(() => {
		if (open && !isInitialized) {
			isInitialized = true;
			initializeForm();
		} else if (!open) {
			isInitialized = false;
		}
	});

	async function initializeForm() {
		await loadExpertPlugins();
		currentStep = 1;
		
		if (entity) {
			const savedLang = entity.expert_language_plugin;
			formData = { ...entity };
			formData.expert_language_plugin = savedLang;

			if (savedLang) {
				const pluginExists = expertPlugins.some((p: any) => p.language === savedLang);
				if (!pluginExists) {
					toastStore.error(
						`Plugin "${savedLang}" não está instalado`,
						'Os campos de Architecture, Stack e State Management não poderão ser carregados.'
					);
				} else {
					await loadPluginOptions(savedLang);
				}
			}
		} else {
			formData = {
				name: '',
				expert_language_plugin: '',
				prd: '',
				functional_requirements: [],
				non_functional_requirements: [],
				architecture: '',
				persistence: '',
				engineering_philosophies: [],
				design_patterns: [],
				data_patterns: [],
				stack_plugin: '',
				dependency_manifest: [],
				business: {
					state_management: '',
					api_contract: '',
					customization_details: '',
					final_adjustments: '',
					architecture_recommendations: '',
				},
				architecture_health: 100,
				color: '#3b82f6',
				icon: '📝',
			};
		}
	}

	async function loadExpertPlugins() {
		try {
			expertPlugins = await (window as any).go.main.App.GetExperts();
		} catch (e) {
			console.error('Failed to load expert plugins:', e);
		}
	}

	async function loadPluginOptions(lang: string) {
		if (isLoadingOptions) return;
		isLoadingOptions = true;
		loadingOptions = true;
		try {
			const [archs, stks, states] = await Promise.all([
				(window as any).go.main.App.GetPatterns(lang),
				(window as any).go.main.App.GetStacks(lang),
				(window as any).go.main.App.GetStateManagement(lang)
			]);
			architectures = archs || [];
			stacks = stks || [];
			stateOptions = states || [];
		} catch (e) {
			console.error('Failed to load plugin options:', e);
		} finally {
			loadingOptions = false;
			isLoadingOptions = false;
		}
	}

	async function handleExpertChange(lang: string) {
		formData.expert_language_plugin = lang;
		if (!lang) return;

		const pluginExists = expertPlugins.some((p: any) => p.language === lang);
		if (!pluginExists) {
			toastStore.error(
				`Plugin "${lang}" não encontrado`,
				'O Expert Language Plugin selecionado não está instalado ou ativado.'
			);
			return;
		}

		await loadPluginOptions(lang);
		formData.architecture = '';
		formData.persistence = '';
		formData.stack_plugin = '';
		formData.dependency_manifest = [];
		formData.business.state_management = '';
	}

	const steps = [
		{ id: 1, label: 'Identity' },
		{ id: 2, label: 'Architecture' },
		{ id: 3, label: 'Patterns' },
		{ id: 4, label: 'Stack' },
		{ id: 5, label: 'Business' },
		{ id: 6, label: 'Advisor' },
	];

	function nextStep() {
		if (currentStep < 6) currentStep++;
	}

	function prevStep() {
		if (currentStep > 1) currentStep--;
	}

	function addLibrary() {
		if (!newLibName.trim() || !newLibVersion.trim()) return;
		formData.dependency_manifest = [...(formData.dependency_manifest || []), {
			lib: newLibName.trim(),
			ver: newLibVersion.trim(),
			mandatory: newLibMandatory
		}];
		newLibName = '';
		newLibVersion = '';
		newLibMandatory = false;
	}

	function handleSave() {
		onSave(formData);
		onOpenChange(false);
	}

	const inputBase = 'rounded-lg px-4 py-3 text-[14px] border border-[var(--border-primary)] bg-[var(--surface-input)] outline-none transition-all focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]';
</script>

<Dialog bind:open onOpenChange={onOpenChange}>
	<DialogPortal>
		<DialogOverlay class="z-[60] bg-black/40 backdrop-blur-sm" />
		<DialogContent 
			class="z-[70] sm:max-w-[1000px] w-[95vw] h-[90dvh] p-0 overflow-hidden flex flex-col bg-[var(--bg-tertiary)] rounded-2xl border border-[var(--border-primary)] shadow-2xl"
			showCloseButton={false}
		>
			<EntityHeader
				icon={formData.icon}
				color={formData.color}
				entityType="Spec Wizard"
				isNew={!entity}
				onIconChange={(emoji) => (formData.icon = emoji)}
				onColorChange={(c) => (formData.color = c)}
				onClose={() => onOpenChange(false)}
			/>

			<!-- ── Header Stepper Area ── -->
			<div class="px-8 py-6 bg-gradient-to-r from-[var(--bg-secondary)] to-[var(--bg-tertiary)] border-b border-[var(--border-primary)] relative shrink-0">
				<div class="flex items-center gap-6 px-2">
					<button 
						type="button" 
						disabled={currentStep === 1}
						onclick={prevStep}
						class={cn(
							"w-10 h-10 rounded-full flex items-center justify-center bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-primary)] transition-all shrink-0 shadow-md",
							currentStep === 1 ? "opacity-20 grayscale cursor-not-allowed" : "hover:bg-[var(--surface-hover)] cursor-pointer active:scale-90"
						)}
					>
						<Icon name="arrowUp" class="-rotate-90" size={18} />
					</button>

					<div class="flex-1 relative px-6">
						<div class="absolute top-5 left-6 right-6 h-1 bg-[var(--border-primary)] rounded-full overflow-hidden">
							<div 
								class="h-full bg-[#ec4899] transition-all duration-500" 
								style="width: {((currentStep - 1) / 5) * 100}%"
							></div>
						</div>
						<div class="flex justify-between relative z-10">
							{#each steps as step}
								<!-- svelte-ignore a11y_click_events_have_key_events -->
								<!-- svelte-ignore a11y_no_static_element_interactions -->
								<div class="flex flex-col items-center gap-2 group cursor-pointer" onclick={() => (currentStep = step.id)}>
									<div class={cn(
										"w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold transition-all border-2 z-20 bg-[var(--bg-tertiary)]",
										currentStep >= step.id 
											? "border-[#ec4899] text-[#ec4899] shadow-[0_0_15px_rgba(236,72,153,0.2)]" 
											: "border-[var(--border-primary)] text-[var(--text-muted)]"
									)}>
										{#if currentStep > step.id}
											<Icon name="check" size={16} />
										{:else}
											{step.id}
										{/if}
									</div>
									<span class={cn(
										"text-[10px] font-bold uppercase tracking-wider transition-colors whitespace-nowrap",
										currentStep === step.id ? "text-[var(--text-primary)]" : "text-[var(--text-faint)]"
									)}>{step.label}</span>
								</div>
							{/each}
						</div>
					</div>

					<button 
						type="button" 
						onclick={currentStep === 6 ? handleSave : nextStep}
						class="w-10 h-10 rounded-full flex items-center justify-center bg-[#ec4899] text-white shadow-lg hover:brightness-110 active:scale-90 transition-all shrink-0 cursor-pointer"
					>
						{#if currentStep === 6}
							<Icon name="send" size={18} />
						{:else}
							<Icon name="arrowUp" class="rotate-90" size={18} />
						{/if}
					</button>
				</div>
			</div>

			<!-- ── Step Content ── -->
			<div class="flex-1 overflow-y-auto px-10 py-8 bg-[var(--bg-tertiary)]">
				{#if currentStep === 1}
					<div class="flex flex-col gap-6">
						<div class="flex flex-col gap-1.5">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">Project Name</label>
							<input bind:value={formData.name} placeholder="Local Vault" class={inputBase} />
						</div>

						<div class="flex flex-col gap-1.5">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">Expert Language Plugin</label>
							<ThemedSelect
								value={formData.expert_language_plugin}
								onValueChange={handleExpertChange}
								options={expertPlugins.map(p => ({ label: p.name, value: p.language }))}
								class="w-full h-12"
							/>
						</div>

						<div class="mt-4 border-t border-[var(--border-primary)] pt-6 flex flex-col gap-4">
							<Collapsible.Root bind:open={domainOpen} class="w-full bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] p-4 shadow-sm">
									<Collapsible.Trigger>
										{#snippet child({ props: tp })}
										<button
											{...tp}
											class="flex w-full items-center justify-between transition-colors cursor-pointer group"
										>
											<div class="flex items-center gap-3">
												<div class="w-8 h-8 rounded-lg bg-[var(--accent-primary)]/10 flex items-center justify-center text-[var(--accent-primary)]">
													<Icon name="history" size={14} />
												</div>
												<h3 class="text-xs font-bold uppercase tracking-widest text-[var(--text-primary)]">Domain & Scope Definition</h3>
											</div>
												<Icon 
													name="chevron-down" 
													size={16} 
													class={cn('transition-transform duration-300', domainOpen && 'rotate-180')} 
													color="var(--text-faint)"
												/>
										</button>
									{/snippet}
								</Collapsible.Trigger>
								<Collapsible.Content class="pt-6 flex flex-col gap-6">
									<div class="flex flex-col gap-1.5">
										<!-- svelte-ignore a11y_label_has_associated_control -->
										<label class="text-[11px] font-bold uppercase tracking-widest text-[var(--text-faint)] px-1">PRD — Problem Definition</label>
										<ExpandableTextarea bind:value={formData.prd} minHeight={120} placeholder="Define the core problem..." />
									</div>

									<div class="flex flex-col gap-1.5">
										<!-- svelte-ignore a11y_label_has_associated_control -->
										<label class="text-[11px] font-bold uppercase tracking-widest text-[var(--text-faint)] px-1">Functional Requirements</label>
										<ExpandableTextarea 
											value={formData.functional_requirements?.join('\n')} 
											oninput={(v) => (formData.functional_requirements = v.split('\n'))}
											minHeight={120} 
											placeholder="List functional requirements..." 
										/>
									</div>

									<div class="flex flex-col gap-1.5">
										<!-- svelte-ignore a11y_label_has_associated_control -->
										<label class="text-[11px] font-bold uppercase tracking-widest text-[var(--text-faint)] px-1">Non-Functional Requirements</label>
										<ExpandableTextarea 
											value={formData.non_functional_requirements?.join('\n')} 
											oninput={(v) => (formData.non_functional_requirements = v.split('\n'))}
											minHeight={120} 
											placeholder="Performance, security, etc..." 
										/>
									</div>
								</Collapsible.Content>
							</Collapsible.Root>
						</div>
					</div>
				{:else if currentStep === 2}
					<div class="flex flex-col gap-6">
						<div class="bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-xl p-8 shadow-sm">
							<div class="flex items-center gap-3 mb-6">
								<div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
									<Icon name="cog" size={16} />
								</div>
								<h3 class="text-sm font-bold uppercase tracking-widest">Architecture</h3>
							</div>
							<div class="flex flex-col gap-1.5">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-xs font-bold uppercase text-[var(--text-faint)] px-1 mb-1">Select Base Architecture</label>
								<ThemedSelect
									value={formData.architecture}
									onValueChange={(v) => (formData.architecture = v)}
									options={architectures.map(a => ({ label: a.name, value: a.id }))}
									class="w-full h-12"
								/>
							</div>
						</div>

						<div class="bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-xl p-8 shadow-sm">
							<div class="flex items-center gap-3 mb-6">
								<div class="w-8 h-8 rounded-lg bg-orange-500/10 flex items-center justify-center text-orange-500">
									<Icon name="layers" size={16} />
								</div>
								<h3 class="text-sm font-bold uppercase tracking-widest">Persistence</h3>
							</div>
							<div class="flex flex-col gap-1.5">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-xs font-bold uppercase text-[var(--text-faint)] px-1 mb-1">Persistence Strategy</label>
								<ThemedSelect
									value={formData.persistence}
									onValueChange={(v) => (formData.persistence = v)}
									options={[{label: 'SQL (PostgreSQL, MySQL)', value: 'sql'}, {label: 'NoSQL (MongoDB, Redis)', value: 'nosql'}]}
									class="w-full h-12"
								/>
							</div>
						</div>
					</div>
				{:else if currentStep === 3}
					<div class="flex flex-col gap-8">
						<div class="flex flex-col gap-4">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center text-purple-500">
									<Icon name="sparkles" size={16} />
								</div>
								<h3 class="text-sm font-bold uppercase tracking-widest">Patterns & Philosophies</h3>
							</div>
							
							<div class="flex flex-col gap-3">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--text-faint)] px-1">Engineering Philosophies</label>
								<div class="flex flex-wrap gap-3">
									{#each ['KISS', 'DRY', 'SOLID', 'YAGNI'] as phil}
										<button 
											type="button"
											onclick={() => {
												if (formData.engineering_philosophies.includes(phil)) 
													formData.engineering_philosophies = formData.engineering_philosophies.filter((p: string) => p !== phil);
												else 
													formData.engineering_philosophies = [...formData.engineering_philosophies, phil];
											}}
											class={cn(
												"flex items-center gap-2.5 px-4 py-2.5 rounded-xl border transition-all text-xs font-bold cursor-pointer",
												formData.engineering_philosophies.includes(phil)
													? "bg-[#ec4899] border-[#ec4899] text-white shadow-lg"
													: "bg-[var(--surface-input)] border-[var(--border-primary)] text-[var(--text-muted)] hover:border-[var(--border-hover)]"
											)}
										>
											{#if formData.engineering_philosophies.includes(phil)}
												<Icon name="check" size={14} />
											{:else}
												<div class="w-3.5 h-3.5 rounded-full border-2 border-current opacity-30"></div>
											{/if}
											{phil}
										</button>
									{/each}
								</div>
							</div>

							<div class="flex flex-col gap-3 mt-4">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--text-faint)] px-1">Design Patterns (GoF)</label>
								<div class="flex flex-wrap gap-3">
									{#each ['Adapter', 'Facade', 'Builder', 'Factory', 'Singleton', 'Observer', 'Strategy'] as pat}
										<button 
											type="button"
											onclick={() => {
												if (formData.design_patterns.includes(pat)) 
													formData.design_patterns = formData.design_patterns.filter((p: string) => p !== pat);
												else 
													formData.design_patterns = [...formData.design_patterns, pat];
											}}
											class={cn(
												"flex items-center gap-2.5 px-4 py-2.5 rounded-xl border transition-all text-xs font-bold cursor-pointer",
												formData.design_patterns.includes(pat)
													? "bg-[#ec4899] border-[#ec4899] text-white shadow-lg"
													: "bg-[var(--surface-input)] border-[var(--border-primary)] text-[var(--text-muted)] hover:border-[var(--border-hover)]"
											)}
										>
											{#if formData.design_patterns.includes(pat)}
												<Icon name="check" size={14} />
											{:else}
												<div class="w-3.5 h-3.5 rounded-full border-2 border-current opacity-30"></div>
											{/if}
											{pat}
										</button>
									{/each}
								</div>
							</div>

							<div class="flex flex-col gap-3 mt-4">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--text-faint)] px-1">Data Patterns — Access</label>
								<div class="flex flex-wrap gap-3">
									{#each ['Repository', 'Active Record', 'DAO'] as dp}
										<button 
											type="button"
											onclick={() => {
												if (formData.data_patterns.includes(dp)) 
													formData.data_patterns = formData.data_patterns.filter((p: string) => p !== dp);
												else 
													formData.data_patterns = [...formData.data_patterns, dp];
											}}
											class={cn(
												"flex items-center gap-2.5 px-4 py-2.5 rounded-xl border transition-all text-xs font-bold cursor-pointer",
												formData.data_patterns.includes(dp)
													? "bg-[#ec4899] border-[#ec4899] text-white shadow-lg"
													: "bg-[var(--surface-input)] border-[var(--border-primary)] text-[var(--text-muted)] hover:border-[var(--border-hover)]"
											)}
										>
											{#if formData.data_patterns.includes(dp)}
												<Icon name="check" size={14} />
											{:else}
												<div class="w-3.5 h-3.5 rounded-full border-2 border-current opacity-30"></div>
											{/if}
											{dp}
										</button>
									{/each}
								</div>
							</div>
						</div>
					</div>
				{:else if currentStep === 4}
					<div class="flex flex-col gap-6">
						<div class="flex flex-col gap-1.5">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">Stack Plugin</label>
							<ThemedSelect
								value={formData.stack_plugin}
								onValueChange={(v) => {
									formData.stack_plugin = v;
									const selectedStack = stacks.find(s => s.id === v);
									if (selectedStack && selectedStack.libraries) {
								formData.dependency_manifest = selectedStack.libraries.map((lib: any) => ({
										lib: lib.name,
										ver: 'latest',
										mandatory: !!lib.mandatory
									}));
									}
								}}
								options={stacks.map(s => ({ label: s.name, value: s.id }))}
								class="w-full h-12"
							/>
						</div>

						<div class="bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-2xl shadow-sm overflow-hidden mt-4">
							<div class="px-8 py-5 border-b border-[var(--border-primary)] bg-[var(--surface-elevated)]/50">
								<div class="flex items-center gap-3">
									<Icon name="wrench" size={16} class="text-[var(--accent-primary)]" />
									<h3 class="text-sm font-bold uppercase tracking-widest">Dependency Manifest</h3>
								</div>
							</div>
							<div class="p-8">
								<div class="flex flex-col gap-3">
									{#if formData.dependency_manifest && formData.dependency_manifest.length > 0}
										<div class="grid grid-cols-[1fr_100px_100px_50px] gap-4 px-4 text-[10px] uppercase font-bold text-[var(--text-faint)] tracking-widest">
											<span>Library</span>
											<span class="text-center">Version</span>
											<span class="text-center">Mandatory</span>
											<span></span>
										</div>
										{#each (formData.dependency_manifest || []) as lib, i}
											<div class="grid grid-cols-[1fr_100px_100px_50px] gap-4 items-center bg-[var(--surface-input)] border border-[var(--border-primary)] rounded-xl px-6 py-3 transition-all hover:border-[var(--border-hover)]">
												<span class="text-[13px] font-mono font-bold truncate text-[var(--text-primary)]">{lib.lib || lib.name || 'Library'}</span>
												<input 
													type="text" 
													value={lib.ver || lib.version || ''} 
													oninput={(e) => {
														const val = (e.target as HTMLInputElement).value;
														if (lib.ver !== undefined) lib.ver = val;
														else lib.version = val;
													}}
													class="bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-lg px-2 py-1.5 text-xs text-center font-mono" 
												/>
												<div class="flex justify-center">
													<Switch 
														checked={!!lib.mandatory} 
														onCheckedChange={(v) => lib.mandatory = v}
														size="sm" 
													/>
												</div>
												<button type="button" onclick={() => (formData.dependency_manifest = formData.dependency_manifest.filter((_: unknown, idx: number) => idx !== i))} class="text-[var(--text-faint)] hover:text-red-500 p-2 transition-colors cursor-pointer">
													<Icon name="x" size={16} />
												</button>
											</div>
										{/each}
									{:else}
										<div class="py-8 text-center border-2 border-dashed border-[var(--border-primary)] rounded-2xl opacity-40">
											<p class="text-xs uppercase font-bold tracking-widest">No dependencies selected</p>
										</div>
									{/if}
									<div class="grid grid-cols-[1fr_100px_100px_auto] gap-4 items-end mt-4">
										<div class="flex flex-col gap-1">
											<label class="text-[10px] uppercase font-bold tracking-widest text-[var(--text-faint)]">Library</label>
											<input
												type="text"
												bind:value={newLibName}
												placeholder="library name"
												class="bg-[var(--surface-input)] border border-[var(--border-primary)] rounded-lg px-3 py-2 text-xs font-mono outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
											/>
										</div>
										<div class="flex flex-col gap-1">
											<label class="text-[10px] uppercase font-bold tracking-widest text-[var(--text-faint)]">Version</label>
											<input
												type="text"
												bind:value={newLibVersion}
												placeholder="e.g. 1.0.0"
												class="bg-[var(--surface-input)] border border-[var(--border-primary)] rounded-lg px-3 py-2 text-xs font-mono outline-none focus:ring-1 focus:ring-[var(--accent-primary)]/30 focus:border-[var(--accent-primary)]"
											/>
										</div>
										<div class="flex flex-col gap-1 items-center">
											<label class="text-[10px] uppercase font-bold tracking-widest text-[var(--text-faint)]">Mandatory</label>
											<Switch checked={newLibMandatory} onCheckedChange={(v) => newLibMandatory = v} size="sm" />
										</div>
										<button
											type="button"
											disabled={!newLibName.trim() || !newLibVersion.trim()}
											onclick={addLibrary}
											class="flex items-center justify-center gap-2 px-6 py-2 rounded-xl text-[11px] font-bold uppercase tracking-[0.2em] transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed disabled:pointer-events-none {newLibName.trim() && newLibVersion.trim() ? 'bg-[var(--accent-primary)] text-white shadow-lg hover:brightness-110 active:scale-90' : 'bg-[var(--surface-input)] border border-[var(--border-primary)] text-[var(--text-muted)]'}"
										>
											<Icon name="plus" size={14} /> Add
										</button>
									</div>
								</div>
							</div>
						</div>
					</div>
				{:else if currentStep === 5}
					<div class="flex flex-col gap-8">
						<div class="flex flex-col gap-1.5">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">State Management</label>
							<ThemedSelect
								value={formData.business.state_management}
								onValueChange={(v) => (formData.business.state_management = v)}
								options={stateOptions.map(s => ({ label: s.name, value: s.id }))}
								class="w-full h-12"
							/>
						</div>

						<div class="flex flex-col gap-2">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">API Contract / Communication</label>
							<ExpandableTextarea bind:value={formData.business.api_contract} minHeight={120} placeholder="Describe the API surface..." />
						</div>

						<div class="flex flex-col gap-2">
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<label class="text-xs font-bold uppercase text-[var(--text-muted)] px-1">Customization Details & Nuances</label>
							<ExpandableTextarea bind:value={formData.business.customization_details} minHeight={120} placeholder="Detail edge cases and nuances..." />
						</div>
					</div>
				{:else if currentStep === 6}
					<div class="flex flex-col gap-6">
						<div class="flex flex-col gap-2">
							<h3 class="text-sm font-bold uppercase tracking-widest px-1">Final Adjustments & Advisor</h3>
							<div class="flex flex-col gap-2 mt-2">
								<!-- svelte-ignore a11y_label_has_associated_control -->
								<label class="text-[11px] font-bold uppercase text-[var(--text-faint)] px-1">Implementation Instructions</label>
								<ExpandableTextarea bind:value={formData.business.final_adjustments} minHeight={150} placeholder="Any specific instructions for the AI advisor..." />
							</div>
						</div>

						<div class="flex flex-col gap-4 mt-4">
							<h3 class="text-xs font-bold uppercase tracking-[0.2em] text-[var(--text-faint)] px-1 mb-1">Architecture Recommendations</h3>
							
							<div class="p-6 rounded-2xl border border-green-500/30 bg-green-500/5 flex flex-col gap-2 shadow-sm">
								<div class="flex items-center gap-2 text-green-500">
									<Icon name="check" size={16} />
									<span class="text-xs font-bold uppercase tracking-widest">High Testability</span>
								</div>
								<p class="text-[12px] leading-relaxed text-[var(--text-muted)]">Repositories with SOLID facilitate Mocking and Dependency Injection. Your structure allows for easy unit and integration testing.</p>
							</div>

							<div class="p-6 rounded-2xl border border-yellow-500/30 bg-yellow-500/5 flex flex-col gap-2 shadow-sm">
								<div class="flex items-center gap-2 text-yellow-500">
									<Icon name="cog" size={16} />
									<span class="text-xs font-bold uppercase tracking-widest">'Fat' Model Danger</span>
								</div>
								<p class="text-[12px] leading-relaxed text-[var(--text-muted)]">Business rules and persistence seem tightly coupled. Watch out for giant classes that violate the Single Responsibility Principle.</p>
							</div>

							<div class="p-6 rounded-2xl border border-red-500/30 bg-red-500/5 flex flex-col gap-2 shadow-sm">
								<div class="flex items-center gap-2 text-red-500">
									<Icon name="loader" size={16} />
									<span class="text-xs font-bold uppercase tracking-widest">CRITICAL: Architecture Overload</span>
								</div>
								<p class="text-[12px] leading-relaxed text-[var(--text-muted)]">Your architecture is extremely overloaded. The number of patterns selected for a project of this size will cause significant slowdown in development and maintenance.</p>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- ── Health Bar Footer (Fixed at bottom) ── -->
			<div class="px-8 py-5 bg-[var(--surface-elevated)] border-t border-[var(--border-primary)] flex flex-col items-center gap-3 shrink-0">
				<div class="flex items-center gap-2">
					<div class="flex h-5 w-5 items-center justify-center rounded-full bg-green-500/10 text-green-500 animate-pulse">
						<div class="w-1.5 h-1.5 rounded-full bg-current"></div>
					</div>
					<span class="text-[11px] font-bold uppercase tracking-[0.2em] text-[var(--text-secondary)]">Architecture Health {formData.architecture_health}%</span>
				</div>
				<div class="w-64 h-2 bg-[var(--border-primary)] rounded-full overflow-hidden shadow-inner">
					<div 
						class="h-full bg-green-500 shadow-[0_0_15px_rgba(34,197,94,0.5)] transition-all duration-1000 ease-out" 
						style="width: {formData.architecture_health}%"
					></div>
				</div>
			</div>
		</DialogContent>
	</DialogPortal>
</Dialog>

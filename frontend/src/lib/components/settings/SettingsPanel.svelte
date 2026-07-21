<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { TooltipProvider } from '$lib/components/ui/tooltip';
	import CardList from './CardList.svelte';
	import EntityEditDialog from './EntityEditDialog.svelte';
	import SpecWizardDialog from './SpecWizardDialog.svelte';
import WorkspaceDialog from './WorkspaceDialog.svelte';
	import GeneralSettings from './GeneralSettings.svelte';
	import ConfirmDeleteDialog from '$lib/components/ui/ConfirmDeleteDialog.svelte';
	import { toastStore } from '$lib/stores/toast.svelte';
	import { entityStore, FIELD_CONFIGS } from '$lib/stores/entities.svelte';
	import type { EntityCardData } from './EntityCard.svelte';

	interface SettingsPanelProps {
		class?: string;
		onClose?: () => void;
	}

	let { class: className, onClose }: SettingsPanelProps = $props();

	// ── Navigation categories ──
	const categories = [
		{ id: 'general', label: 'General', icon: 'cog' },
		{ id: 'agents', label: 'Agents', icon: 'bot' },
		{ id: 'skills', label: 'Skills', icon: 'sparkles' },
		{ id: 'mcp', label: 'MCP', icon: 'plug' },
		{ id: 'models', label: 'Models', icon: 'cube' },
		{ id: 'workers', label: 'Workers', icon: 'users' },
		{ id: 'workspaces', label: 'Workspaces', icon: 'folder' },
		{ id: 'spec-wizard', label: 'Spec Wizard', icon: 'wand' },
		{ id: 'tools', label: 'Tools', icon: 'wrench' },
	] as const;

	type CategoryId = (typeof categories)[number]['id'];
	let activeCategory = $state<CategoryId>('general');

	// ── Edit dialog state ──
	let dialogOpen = $state(false);
	let dialogEntity = $state<Record<string, any> | null>(null);
	let deleteConfirmOpen = $state(false);
	let itemToDelete = $state<EntityCardData | null>(null);

	// ── Load data when category changes ──
	let loadedCategories = $state<Set<string>>(new Set());

	$effect(() => {
		if (activeCategory !== 'general' && !loadedCategories.has(activeCategory)) {
			loadedCategories.add(activeCategory);
			entityStore.load(activeCategory);
		}
	});

	function openNew() {
		dialogEntity = null;
		dialogOpen = true;
	}

	function openEdit(item: EntityCardData) {
		dialogEntity = { ...item };
		dialogOpen = true;
	}

	async function handleSave(data: Record<string, any>) {
		try {
			await entityStore.save(activeCategory, data);
			dialogOpen = false;
			toastStore.success('Success! Your changes have been saved');
		} catch (e) {
			console.error('[SettingsPanel] Save failed:', e);
			toastStore.error('Unable to save changes', String(e));
		}
	}

	function confirmDelete(item: EntityCardData) {
		itemToDelete = item;
		deleteConfirmOpen = true;
	}

	async function handleDelete() {
		if (!itemToDelete) return;
		try {
			await entityStore.remove(activeCategory, { ...itemToDelete });
			toastStore.success('Item deleted successfully');
			itemToDelete = null;
		} catch (e) {
			console.error('[SettingsPanel] Delete failed:', e);
			toastStore.error('Failed to delete item', String(e));
		}
	}

	const currentItems = $derived(entityStore.getItems(activeCategory));
	const isLoading = $derived(entityStore.isLoading(activeCategory));
</script>

<TooltipProvider>
	<aside
		class={cn(
			'flex h-full shrink-0',
			'border-l border-[var(--border-primary)]',
			'bg-[var(--bg-secondary)]',
			'w-[525px]',
			className
		)}
	>
		<!-- ── Settings Nav (left column) ── -->
		<nav class="flex w-[140px] shrink-0 flex-col border-r border-[var(--border-primary)]">
			<!-- Header -->
			<div class="px-4 pt-5 pb-3">
				<h2
					class="text-xs font-semibold tracking-[0.2em] uppercase"
					style="color: var(--text-muted); font-family: var(--font-display)"
				>
					Settings
				</h2>
			</div>

			<!-- Nav items -->
			<div class="flex-1 overflow-y-auto px-2 py-1">
				{#each categories as cat}
					<button
						type="button"
						onclick={() => (activeCategory = cat.id)}
						class={cn(
							'flex w-full items-center gap-2.5 rounded-lg px-3 py-2',
							'text-xs font-medium cursor-pointer transition-colors',
							activeCategory === cat.id
								? 'bg-[var(--surface-hover)]'
								: 'hover:bg-[var(--surface-hover)]'
						)}
						style={
							activeCategory === cat.id
								? 'color: var(--text-primary)'
								: 'color: var(--text-muted)'
						}
					>
						<Icon name={cat.icon} size={14} />
						<span>{cat.label}</span>
					</button>
				{/each}
			</div>
		</nav>

		<!-- ── Settings Content (right area) ── -->
		<div class="flex flex-1 flex-col min-w-0">
			<!-- Content header -->
			<div class="flex items-center justify-between border-b border-[var(--border-primary)] px-5 pt-5 pb-3">
				<h3 class="text-sm font-semibold" style="color: var(--text-primary)">
					{categories.find((c) => c.id === activeCategory)?.label}
				</h3>

				<button
					type="button"
					onclick={onClose}
					class={cn(
						'flex items-center justify-center w-7 h-7 rounded-md',
						'transition-colors cursor-pointer',
						'hover:bg-[var(--surface-hover)]'
					)}
					style="color: var(--text-muted)"
					title="Close settings"
				>
					<Icon name="x" size={15} />
				</button>
			</div>

			<!-- ── Content body ── -->
			{#if activeCategory === 'general'}
				<GeneralSettings />
			{:else if isLoading}
				<!-- Loading state -->
				<div class="flex flex-1 items-center justify-center">
					<div
						class="h-5 w-5 rounded-full border-2 border-[var(--border-primary)] border-t-[var(--accent-primary)] animate-spin"
					></div>
				</div>
			{:else}
				<!-- Card list for entity categories -->
					<CardList
						items={currentItems}
						entityType={categories.find((c) => c.id === activeCategory)?.label ?? 'item'}
						onNew={openNew}
						onEdit={openEdit}
						onDelete={confirmDelete}
						emptyMessage={`No ${categories.find((c) => c.id === activeCategory)?.label?.toLowerCase() ?? 'items'} yet`}
					/>
				{/if}
			</div>

			<!-- ── Delete confirmation ── -->
			<ConfirmDeleteDialog
				bind:open={deleteConfirmOpen}
				onOpenChange={(val) => (deleteConfirmOpen = val)}
				title="Delete Item"
				description={`Are you sure you want to delete ${itemToDelete?.name}? This action cannot be undone.`}
				onConfirm={handleDelete}
			/>

<!-- ── Edit/Create Dialog ── -->
				{#if activeCategory !== 'general'}
					{#if activeCategory === 'spec-wizard'}
						<SpecWizardDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							onSave={handleSave}
						/>
					{:else if activeCategory === 'workspaces'}
						<WorkspaceDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							onSave={handleSave}
						/>
					{:else if FIELD_CONFIGS[activeCategory]}
						<EntityEditDialog
							bind:open={dialogOpen}
							onOpenChange={(val) => (dialogOpen = val)}
							entity={dialogEntity}
							entityType={categories.find((c) => c.id === activeCategory)?.label ?? 'item'}
							fields={FIELD_CONFIGS[activeCategory]}
							onSave={handleSave}
						/>
					{/if}
				{/if}
	</aside>
</TooltipProvider>

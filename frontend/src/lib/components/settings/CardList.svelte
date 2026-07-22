<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import EntityCard from './EntityCard.svelte';
	import type { EntityCardData } from './EntityCard.svelte';

	interface CardListProps {
		items: EntityCardData[];
		entityType: string;
		onNew?: () => void;
		onEdit?: (item: EntityCardData) => void;
		onDelete?: (item: EntityCardData) => void;
		emptyMessage?: string;
		class?: string;
	}

	let {
		items,
		entityType,
		onNew,
		onEdit,
		onDelete,
		emptyMessage = 'No items yet',
		class: className,
	}: CardListProps = $props();
</script>

<div class={cn('flex flex-col h-full', className)}>
	<!-- ── Card grid ── -->
	<div class="flex-1 overflow-y-auto px-4 pt-4 pb-8 bg-[var(--surface-form)]">
		{#if items.length > 0}
			<div class="flex flex-col gap-2">
				{#each items as item (item.id)}
					<EntityCard
						{item}
						onEdit={onEdit}
						onDelete={onDelete}
					/>
				{/each}
			</div>
		{:else}
			<div class="flex items-center justify-center py-12">
				<p class="text-xs text-center" style="color: var(--text-faint)">
					{emptyMessage}
				</p>
			</div>
		{/if}
	</div>
</div>

<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { Tooltip, TooltipContent, TooltipTrigger } from '$lib/components/ui/tooltip';
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
	<!-- ── Toolbar: count + new button ── -->
	<div class="flex items-center justify-between px-4 py-2.5">
		<span class="text-[11px] font-medium" style="color: var(--text-muted)">
			{items.length} {items.length === 1 ? 'item' : 'items'}
		</span>

		{#if onNew}
				<Tooltip>
					<TooltipTrigger>
						{#snippet child({ props })}
							<button
								{...props}
								type="button"
								onclick={onNew}
								class={cn(
									'flex items-center gap-1.5 px-2.5 py-1 rounded-lg',
									'text-[11px] font-medium cursor-pointer',
									'transition-colors',
									'hover:bg-[var(--surface-hover)]'
								)}
								style="color: var(--text-secondary)"
							>
								<Icon name="plus" size={12} />
								new
							</button>
						{/snippet}
					</TooltipTrigger>
				<TooltipContent side="top">
					Add new {entityType}
				</TooltipContent>
			</Tooltip>
		{/if}
	</div>

	<!-- ── Card grid ── -->
	<div class="flex-1 overflow-y-auto px-4 pb-4">
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

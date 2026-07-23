<script lang="ts">
	import { cn } from '$lib/utils';
	import { Tooltip, TooltipContent, TooltipTrigger } from '$lib/components/ui/tooltip';
	import { Icon } from '$lib/components/icon';

	export interface EntityCardData {
		id: string | number;
		name: string;
		description?: string;
		icon: string;
		color: string;
		[key: string]: any;
	}

	interface EntityCardProps {
		item: EntityCardData;
		onEdit?: (item: EntityCardData) => void;
		onDelete?: (item: EntityCardData) => void;
		class?: string;
	}

	let { item, onEdit, onDelete, class: className }: EntityCardProps = $props();

	function handleDelete(e: MouseEvent) {
		e.stopPropagation();
		onDelete?.(item);
	}

	function handleEditBtn(e: MouseEvent) {
		e.stopPropagation();
		onEdit?.(item);
	}
</script>

<div
	class={cn(
		'group/card relative w-full text-left',
		'rounded-xl overflow-hidden transition-all duration-150',
		'hover:brightness-110',
		'ring-1 ring-[var(--border-primary)]',
		'hover:ring-[var(--border-hover)]',
		'bg-[var(--bg-secondary)]',
		'shadow-sm',
		className
	)}
>
	<!-- ── Color header bar ── -->
	<div
		class="flex items-center gap-2.5 px-3 py-2"
		style="background: linear-gradient(135deg, {item.color}66, {item.color}33)"
	>
		<!-- Icon (emoji from backend) -->
		<span class="flex items-center justify-center text-xl leading-none select-none">
			{item.icon}
		</span>

		<!-- Name -->
		<span
			class="flex-1 text-xs font-semibold truncate"
			style="color: var(--text-primary)"
		>
			{item.name}
		</span>

		<!-- Action buttons (visible on hover) -->
		<div class="flex items-center gap-0.5 opacity-0 group-hover/card:opacity-100 transition-opacity">
			{#if onEdit}
				<div data-card-action>
					<Tooltip>
						<TooltipTrigger>
							{#snippet child({ props })}
								<button
									{...props}
									type="button"
									onclick={handleEditBtn}
									class={cn(
										'flex items-center justify-center w-6 h-6 rounded-md',
										'cursor-pointer transition-colors',
										'hover:bg-black/20'
									)}
									style="color: var(--text-primary)"
								>
									<Icon name="pencil" size={12} />
								</button>
							{/snippet}
						</TooltipTrigger>
						<TooltipContent side="top">
							Edit
						</TooltipContent>
					</Tooltip>
				</div>
			{/if}

			{#if onDelete}
				<div data-card-action>
					<Tooltip>
						<TooltipTrigger>
							{#snippet child({ props })}
								<button
									{...props}
									type="button"
									onclick={handleDelete}
									class={cn(
										'flex items-center justify-center w-6 h-6 rounded-md',
										'cursor-pointer transition-colors',
										'hover:bg-black/20'
									)}
									style="color: var(--text-primary)"
								>
									<Icon name="trash-2" size={12} />
								</button>
							{/snippet}
						</TooltipTrigger>
						<TooltipContent side="top">
							Delete
						</TooltipContent>
					</Tooltip>
				</div>
			{/if}
		</div>
	</div>

		<!-- ── Body: up to 3 info lines ── -->
		{#if item.description || item.provider || item.model || item.tags || item.active !== undefined || item.connect_type || item.enabled !== undefined || item.type_connection || item.api_url || item.strategy || item.language || item.connection_type || item.expert_language_plugin || item.architecture || item.stack_plugin || item.command}
			<div class="px-3 py-2 flex flex-col gap-1">
				{#if item.description}
					<p class="text-[11px] leading-relaxed line-clamp-1" style="color: var(--text-muted)">
						{item.description}
					</p>
				{/if}

				<!-- Spec Wizard specific -->
				{#if item.expert_language_plugin}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Expert Language Plugin
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.expert_language_plugin}
						</span>
					</div>
				{/if}
				{#if item.architecture}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Select Base Architecture
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.architecture}
						</span>
					</div>
				{/if}
				{#if item.stack_plugin}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Stack Plugin
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.stack_plugin}
						</span>
					</div>
				{/if}

				<!-- Agents specific -->
				{#if item.provider}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Provider
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.provider}
						</span>
					</div>
				{/if}
				{#if item.model}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Model
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.model}
						</span>
					</div>
				{/if}

				<!-- Skills specific -->
				{#if item.tags}
					<div class="flex items-center gap-1.5">
						<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
							Tags
						</span>
						<span class="text-[10px] truncate" style="color: var(--text-muted)">
							{item.tags}
						</span>
					</div>
				{/if}
					<!-- Workers specific -->
					{#if item.language}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Language
							</span>
							<span class="text-[10px] truncate capitalize" style="color: var(--text-muted)">
								{item.language}
							</span>
						</div>
					{/if}
						{#if item.connection_type && !item.connect_type}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Bridge
							</span>
							<span class="text-[10px] truncate uppercase" style="color: var(--text-muted)">
								{item.connection_type}
							</span>
						</div>
					{/if}

					<!-- CLI worker specific -->
					{#if item.connection_type === 'cli' && item.command}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Command
							</span>
							<span class="text-[10px] truncate font-mono" style="color: var(--text-muted)">
								{item.command}
							</span>
						</div>
					{/if}
					{#if item.connection_type === 'cli' && item.arguments}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Args
							</span>
							<span class="text-[10px] truncate font-mono" style="color: var(--text-muted)">
								{item.arguments}
							</span>
						</div>
					{/if}

					<!-- Providers/Models specific -->
					{#if item.type_connection}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Type
							</span>
							<span class="text-[10px] truncate uppercase" style="color: var(--text-muted)">
								{item.type_connection}
							</span>
						</div>
					{/if}
					{#if item.api_url}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								API URL
							</span>
							<span class="text-[10px] truncate font-mono" style="color: var(--text-muted)">
								{item.api_url}
							</span>
						</div>
					{/if}
					{#if item.strategy}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Strategy
							</span>
							<span class="text-[10px] truncate italic" style="color: var(--text-muted)">
								{item.strategy.replace('_', ' ')}
							</span>
						</div>
					{/if}

					<!-- MCP specific -->
					{#if item.connect_type}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Type
							</span>
							<span class="text-[10px] truncate uppercase" style="color: var(--text-muted)">
								{item.connect_type}
							</span>
						</div>
					{/if}
					{#if item.command || item.url}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								{item.connect_type === 'sse' ? 'URL' : 'Command'}
							</span>
							<span class="text-[10px] truncate font-mono" style="color: var(--text-muted)">
								{item.connect_type === 'sse' ? item.url : item.command}
							</span>
						</div>
					{/if}

					<!-- Status for Skills (Active) or MCP (Enabled) -->
					{#if item.active !== undefined || item.enabled !== undefined}
						{@const isActive = item.active ?? item.enabled}
						<div class="flex items-center gap-1.5">
							<span class="text-[9px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">
								Status
							</span>
							<div class="flex items-center gap-1">
								<div 
									class="w-1.5 h-1.5 rounded-full" 
									style="background-color: {isActive ? 'var(--status-success)' : 'var(--text-faint)'}"
								></div>
								<span class="text-[10px]" style="color: var(--text-muted)">
									{isActive ? (item.active !== undefined ? 'Active' : 'Enabled') : (item.active !== undefined ? 'Inactive' : 'Disabled')}
								</span>
							</div>
						</div>
					{/if}
			</div>
		{/if}
</div>

<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import {
		DropdownMenu,
		DropdownContent,
		DropdownItem,
	} from '$lib/components/ui/dropdown';

	interface ModelOption {
		id: string;
		name: string;
	}

	interface ChatToolbarProps {
		isLoading: boolean;
		models: ModelOption[];
		selectedModel: ModelOption;
		selectedMode: 'ASK' | 'EDIT' | 'PLAN';
		modes: readonly ('ASK' | 'EDIT' | 'PLAN')[];
		onSend: () => void;
		onSelectedModelChange?: (model: ModelOption) => void;
		onSelectedModeChange?: (mode: 'ASK' | 'EDIT' | 'PLAN') => void;
	}

	let {
		isLoading,
		models,
		selectedModel,
		selectedMode,
		modes,
		onSend,
		onSelectedModelChange,
		onSelectedModeChange,
	}: ChatToolbarProps = $props();

	function handleSelectModel(model: ModelOption) {
		onSelectedModelChange?.(model);
	}

	function handleSelectMode(mode: 'ASK' | 'EDIT' | 'PLAN') {
		onSelectedModeChange?.(mode);
	}

	function cycleMode() {
		const idx = modes.indexOf(selectedMode);
		const next = modes[(idx + 1) % modes.length];
		handleSelectMode(next);
	}
</script>

<div class="flex items-center justify-between px-6 py-2.5">
	<!-- ── Left: Utility Icons ── -->
	<div class="flex items-center gap-1">
		<button
			type="button"
			title="History"
			class={cn(
				'flex items-center justify-center w-8 h-8 rounded-lg',
				'transition-colors cursor-pointer',
				'hover:bg-[var(--surface-hover)]'
			)}
			style="color: var(--text-muted)"
		>
			<Icon name="history" size={15} />
		</button>

		<button
			type="button"
			title="Attach file"
			class={cn(
				'flex items-center justify-center w-8 h-8 rounded-lg',
				'transition-colors cursor-pointer',
				'hover:bg-[var(--surface-hover)]'
			)}
			style="color: var(--text-muted)"
		>
			<Icon name="attachment" size={15} />
		</button>

		<button
			type="button"
			title="Logs"
			class={cn(
				'flex items-center justify-center w-8 h-8 rounded-lg',
				'transition-colors cursor-pointer',
				'hover:bg-[var(--surface-hover)]'
			)}
			style="color: var(--text-muted)"
		>
			<Icon name="log" size={15} />
		</button>
	</div>

	<!-- ── Right: Status · Mode · Model · Send ── -->
	<div class="flex items-center gap-2">
		<!-- Loading / Status indicator -->
		<button
			type="button"
			class={cn(
				'flex items-center justify-center w-8 h-8 rounded-full',
				'transition-colors cursor-pointer',
				isLoading ? 'animate-spin' : 'hover:bg-[var(--surface-hover)]'
			)}
			style="color: var(--text-muted)"
			title={isLoading ? 'Generating...' : 'Idle'}
		>
			<Icon name="loader" size={15} />
		</button>

		<!-- Mode Selector (ASK / EDIT / PLAN) -->
		<button
			type="button"
			onclick={cycleMode}
			class={cn(
				'flex items-center justify-center h-7 px-3 rounded-md',
				'text-[11px] font-semibold tracking-wider cursor-pointer',
				'border transition-colors',
				'border-[var(--border-primary)]',
				'hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]'
			)}
			style="color: var(--text-secondary)"
		>
			{selectedMode}
		</button>

		<!-- Model Dropdown -->
		<DropdownMenu>
			{#snippet trigger({ toggle })}
				<button
					type="button"
					onclick={toggle}
					class={cn(
						'flex items-center gap-1.5 h-7 px-3 rounded-md',
						'text-[11px] font-mono cursor-pointer',
						'border transition-colors',
						'border-[var(--border-primary)]',
						'hover:bg-[var(--surface-hover)] hover:border-[var(--border-hover)]'
					)}
					style="color: var(--text-secondary)"
				>
					{selectedModel.name}
					<Icon name="chevron-down" size={11} />
				</button>
			{/snippet}

			{#snippet content({ close })}
				<DropdownContent align="end" class="min-w-[200px]">
					{#each models as model}
						<button
							type="button"
							class={cn(
								'relative flex w-full items-center rounded-sm px-2 py-1.5 text-xs cursor-pointer',
								'font-mono transition-colors',
								'hover:bg-[var(--surface-hover)]',
								model.id === selectedModel.id
									? 'font-semibold'
									: ''
							)}
							style={
								model.id === selectedModel.id
									? 'color: var(--accent-primary)'
									: 'color: var(--text-secondary)'
							}
							onclick={() => {
								handleSelectModel(model);
								close();
							}}
						>
							{model.name}
						</button>
					{/each}
				</DropdownContent>
			{/snippet}
		</DropdownMenu>

		<!-- Send Button -->
		<button
			type="button"
			onclick={onSend}
			disabled={isLoading}
			class={cn(
				'flex items-center justify-center w-8 h-8 rounded-lg',
				'cursor-pointer transition-all duration-150',
				'disabled:opacity-30 disabled:cursor-not-allowed',
				'hover:brightness-125 active:scale-95'
			)}
			style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
			title="Send message"
		>
			<Icon name="arrowUp" size={16} />
		</button>
	</div>
</div>

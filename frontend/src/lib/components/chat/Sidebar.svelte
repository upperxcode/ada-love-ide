<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import { Separator } from '$lib/components/ui/separator';

	interface SidebarProps {
		class?: string;
		onOpenSettings?: () => void;
	}

	let { class: className, onOpenSettings }: SidebarProps = $props();

	function handleNewSession() {
		// Future: call Wails binding CreateSession()
		console.log('[Sidebar] New session clicked');
	}

	function handleOpenSettings() {
		// Future: navigate to settings or open settings modal
		console.log('[Sidebar] Settings clicked');
	}
</script>

<aside
	class={cn(
		'flex h-full flex-col shrink-0',
		'border-r border-[var(--border-primary)]',
		'bg-[var(--bg-secondary)]',
		'w-[280px]',
		className
	)}
>
	<!-- ── Workspace Header ── -->
	<div class="px-5 pt-5 pb-3">
		<h2
			class="text-xs font-semibold tracking-[0.2em] uppercase"
			style="color: var(--text-muted); font-family: var(--font-display)"
		>
			ADA LOVE
		</h2>
	</div>

	<!-- ── New Session Button ── -->
	<div class="px-4 pb-4">
		<button
			type="button"
			onclick={handleNewSession}
			class={cn(
				'flex w-full items-center justify-center gap-2',
				'rounded-full px-4 py-2.5 text-sm font-semibold',
				'cursor-pointer transition-all duration-150',
				'hover:brightness-110 active:scale-[0.97]'
			)}
			style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
		>
			<Icon name="plus" size={16} color="var(--accent-primary-fg)" />
			new session
		</button>
	</div>

	<Separator class="mx-4 w-auto" />

	<!-- ── Session History (empty) ── -->
	<div class="flex-1 overflow-y-auto px-3 py-3">
		{#if true}
			<p class="px-2 py-8 text-center text-xs" style="color: var(--text-faint)">
				No sessions yet
			</p>
		{/if}
	</div>

	<Separator class="mx-4 w-auto" />

	<!-- ── Footer Bar ── -->
	<div class="flex items-center justify-between px-4 py-3">
		<!-- Version indicator -->
		<span class="text-[10px] font-mono" style="color: var(--text-faint)">
			v1.1.0
		</span>

		<!-- Settings gear -->
		<button
			type="button"
			onclick={onOpenSettings}
			class={cn(
				'flex items-center justify-center w-7 h-7 rounded-md',
				'transition-colors cursor-pointer',
				'hover:bg-[var(--surface-hover)]'
			)}
			style="color: var(--text-muted)"
		>
			<Icon name="cog" size={15} />
		</button>
	</div>
</aside>

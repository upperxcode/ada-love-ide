<script lang="ts">
	import { cn } from '$lib/utils';

	interface DropdownMenuProps {
		class?: string;
		trigger?: import('svelte').Snippet<[ctx: { open: boolean; toggle: () => void; close: () => void }]>;
		content?: import('svelte').Snippet<[ctx: { close: () => void }]>;
	}

	let { class: className, trigger, content }: DropdownMenuProps = $props();

	let open = $state(false);

	function toggle() {
		open = !open;
	}

	function close() {
		open = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	let containerEl: HTMLDivElement;

	$effect(() => {
		if (!open || typeof document === 'undefined') return;

		function handleClickOutside(e: MouseEvent) {
			if (containerEl && !containerEl.contains(e.target as Node)) {
				close();
			}
		}

		const id = requestAnimationFrame(() => {
			document.addEventListener('click', handleClickOutside);
		});

		return () => {
			cancelAnimationFrame(id);
			document.removeEventListener('click', handleClickOutside);
		};
	});
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	role="menu"
	tabindex="-1"
	class={cn('relative inline-flex', className)}
	bind:this={containerEl}
	onkeydown={handleKeydown}
>
	{#if trigger}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div onclick={toggle} class="contents">
			{@render trigger({ open, toggle, close })}
		</div>
	{/if}

	{#if open && content}
		{@render content({ close })}
	{/if}
</div>

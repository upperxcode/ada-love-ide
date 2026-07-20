<script lang="ts">
	interface DropdownTriggerProps {
		children?: import('svelte').Snippet;
		onclick?: (e: MouseEvent) => void;
		class?: string;
	}
	let { children, onclick, class: className }: DropdownTriggerProps = $props();
</script>

<button
	class={className}
	onclick={onclick}
>
	{#if children}
		{@render children()}
	{/if}
</button>

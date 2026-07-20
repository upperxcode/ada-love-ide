<script lang="ts">
	import { cn } from '$lib/utils';

	interface SeparatorProps {
		orientation?: 'horizontal' | 'vertical';
		class?: string;
	}

	let { orientation = 'horizontal', class: className }: SeparatorProps = $props();
</script>

<div
	role="separator"
	class={cn(
		'shrink-0 bg-[var(--border-primary)]',
		orientation === 'horizontal' ? 'h-[1px] w-full' : 'h-full w-[1px]',
		className
	)}
></div>

<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: className = '' }: { class?: string } = $props();
</script>

<div class={cn('space-y-2', className)}>
	{#each [70, 85, 60, 75] as width, i}
		<div
			class="h-3 rounded-full animate-shimmer"
			style="width: {width}%; animation-delay: {i * 150}ms; --shimmer-duration: 1.5s;"
		></div>
	{/each}
</div>

<style>
	@keyframes shimmer {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}

	:global(.animate-shimmer) {
		background: linear-gradient(
			90deg,
			var(--bg-tertiary) 25%,
			var(--border-primary) 50%,
			var(--bg-tertiary) 75%
		);
		background-size: 200% 100%;
		animation: shimmer var(--shimmer-duration, 1.5s) ease-in-out infinite;
	}
</style>
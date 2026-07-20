<script lang="ts">
	import { cn } from '$lib/utils';
	import { theme } from '$lib/stores/theme.svelte';
	import * as Lucide from 'lucide-svelte';
	import { resolveIcon } from './icon-map';

	interface IconProps {
		name: string;
		size?: number;
		strokeWidth?: number;
		class?: string;
		color?: string;
	}

	let {
		name,
		size = 18,
		strokeWidth = 1.5,
		class: className,
		color,
	}: IconProps = $props();

	let currentIconFamily = $derived(theme.iconTheme.family);

	let resolved = $derived(resolveIcon(name, currentIconFamily));

	type LucideIcon = typeof Lucide.Settings;
	let iconComponent = $derived<LucideIcon | null>(
		resolved.lucide
			? 		(Lucide as unknown as Record<string, typeof Lucide.Settings>)[resolved.lucide] ?? null
			: null,
	);
</script>

{#if currentIconFamily === 'lucide'}
	{#if iconComponent}
		{@const Ic = iconComponent}
		<span class={cn('inline-flex shrink-0', className)}>
			<Ic size={size} strokeWidth={strokeWidth} color={color} />
		</span>
	{:else}
		<span class={cn('inline-flex shrink-0 opacity-30', className)}>
			<Lucide.Circle size={size} strokeWidth={strokeWidth} color={color} />
		</span>
	{/if}
{:else if currentIconFamily === 'material'}
	{#if resolved.material}
		<svg
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			fill="currentColor"
			width={size}
			height={size}
			class={cn('inline-flex shrink-0', className)}
			style={color ? `color: ${color}` : undefined}
		>
			<path d={resolved.material} />
		</svg>
	{:else}
		<svg
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			fill="currentColor"
			width={size}
			height={size}
			class={cn('inline-flex shrink-0 opacity-30', className)}
			style={color ? `color: ${color}` : undefined}
		>
			<path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z" />
		</svg>
	{/if}
{/if}

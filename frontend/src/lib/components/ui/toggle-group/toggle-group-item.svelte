<script lang="ts">
	import { ToggleGroup as ToggleGroupPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";
	import { getContext } from "svelte";

	let { class: className, children, ...restProps }: ToggleGroupPrimitive.ItemProps = $props();

	const ctx = getContext<{ variant: string; size: string }>("toggle-group");

	const variants = {
		default: "data-[state=on]:bg-[var(--accent-primary)] data-[state=on]:text-[var(--accent-primary-fg)]",
		outline: "border border-[var(--border-primary)] bg-transparent hover:bg-[var(--surface-hover)] data-[state=on]:border-[var(--accent-primary)] data-[state=on]:bg-[var(--accent-primary)]/10"
	};

	const sizes = {
		default: "h-9 px-3",
		sm: "h-8 px-2.5 text-xs",
		lg: "h-10 px-5"
	};
</script>

<ToggleGroupPrimitive.Item
	{...restProps}
>
	{#snippet child({ props, pressed })}
		<div
			{...props}
			data-state={pressed ? "on" : "off"}
			class={cn(
				"inline-flex items-center justify-center rounded-md text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--accent-primary)] disabled:pointer-events-none disabled:opacity-50 cursor-pointer",
				variants[ctx.variant as keyof typeof variants] || variants.default,
				sizes[ctx.size as keyof typeof sizes] || sizes.default,
				className
			)}
		>
			{@render children?.({ pressed })}
		</div>
	{/snippet}
</ToggleGroupPrimitive.Item>

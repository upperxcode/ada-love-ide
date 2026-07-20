<script lang="ts">
	import { ToggleGroup as ToggleGroupPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";
	import { getContext, setContext } from "svelte";
	import type { Snippet } from "svelte";

	interface ToggleGroupProps {
		class?: string;
		variant?: "default" | "outline";
		size?: "default" | "sm" | "lg";
		spacing?: number;
		children?: Snippet;
		type?: "single" | "multiple";
		value?: any;
		onValueChange?: (value: string | string[] | undefined) => void;
		[key: string]: unknown;
	}

	let {
		class: className,
		variant = "default",
		size = "default",
		spacing = 0,
		children,
		type = "single",
		value = $bindable<any>(),
		onValueChange,
		...restProps
	}: ToggleGroupProps = $props();

	setContext("toggle-group", { variant, size });
</script>

<ToggleGroupPrimitive.Root
	bind:value
	{type}
	{onValueChange}
	class={cn("flex items-center justify-center", spacing ? `gap-${spacing}` : "gap-0", className)}
	{...restProps}
>
	{@render children?.()}
</ToggleGroupPrimitive.Root>

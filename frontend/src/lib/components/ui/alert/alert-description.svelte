<script lang="ts">
	import { cn } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props();
</script>

<div
	class={cn("text-sm [&_p]:leading-relaxed", className)}
	{...restProps}
>
	{@render children?.()}
</div>

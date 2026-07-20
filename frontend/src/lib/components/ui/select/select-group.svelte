<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let { class: className, children, ...restProps }: SelectPrimitive.GroupProps = $props();
</script>

<SelectPrimitive.Group class={cn("p-0", className)} {...restProps}>
	{@render children?.()}
</SelectPrimitive.Group>

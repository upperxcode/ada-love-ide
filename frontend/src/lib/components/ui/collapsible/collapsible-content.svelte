<script lang="ts">
	import { Collapsible as CollapsiblePrimitive } from "bits-ui";

	let { ...restProps }: CollapsiblePrimitive.ContentProps = $props();
</script>

<CollapsiblePrimitive.Content
	{...restProps}
	class="overflow-hidden data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down"
/>

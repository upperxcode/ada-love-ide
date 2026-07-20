<script lang="ts">
	import type { Snippet } from 'svelte';

	interface SettingRowProps {
		label: string;
		description?: string;
		children: Snippet;
	}

	let { label, description, children }: SettingRowProps = $props();
</script>

<div class="flex items-start gap-3 py-3">
	<!-- Left: label + description -->
	<div class="flex flex-col gap-0.5 min-w-0 flex-1">
			<span class="text-[12px] font-medium leading-tight" style="color: var(--text-primary)">
				{label}
			</span>
			{#if description}
				<span class="text-[11px] leading-tight" style="color: var(--text-muted)">
					{description}
				</span>
		{/if}
		</div>

		<!-- Right: control -->
	<div class="flex-shrink-0 mt-0">
		{@render children()}
	</div>
</div>

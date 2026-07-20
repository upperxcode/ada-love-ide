<script lang="ts">
	import { cn } from '$lib/utils';

	interface SwitchProps {
		checked?: boolean;
		onCheckedChange?: (value: boolean) => void;
		disabled?: boolean;
		class?: string;
		size?: 'default' | 'sm';
	}

	let { checked = $bindable(false), onCheckedChange, disabled = false, class: className, size = 'default' }: SwitchProps = $props();

	function toggle() {
		if (disabled) return;
		checked = !checked;
		onCheckedChange?.(checked);
	}
</script>

<button
	type="button"
	role="switch"
	aria-checked={checked}
	aria-label="Toggle"
	{disabled}
		class={cn(
			'peer inline-flex shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent',
			'shadow-sm transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-primary)]/30',
			'disabled:cursor-not-allowed disabled:opacity-50',
			checked ? 'bg-[var(--accent-primary)]' : 'bg-[var(--surface-input)] border-[var(--border-primary)]',
			size === 'sm' ? 'h-4 w-7' : 'h-6 w-11',
			className
		)}
		onclick={toggle}
	>
		<span
			class={cn(
				'pointer-events-none block rounded-full shadow-lg ring-0 transition-transform',
				'bg-white',
				size === 'sm' ? 'h-3 w-3' : 'h-5 w-5',
				checked ? (size === 'sm' ? 'translate-x-3' : 'translate-x-5') : 'translate-x-0'
			)}
		></span>
</button>

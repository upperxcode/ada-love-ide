<script lang="ts">
	import { cn } from '$lib/utils.js';

	interface ThemedSelectProps {
		value?: string;
		onValueChange?: (value: string) => void;
		placeholder?: string;
		options: { label: string; value: string; icon?: string; color?: string }[];
		class?: string;
		disabled?: boolean;
	}

	let {
		value = $bindable(),
		onValueChange,
		placeholder = 'Select...',
		options = [],
		class: className,
		disabled = false,
	}: ThemedSelectProps = $props();

	// Check if the current value exists in the options
	const valueInOptions = $derived(
		options.some(opt => opt.value === value)
	);

	// Find the label for the current value if it exists in options
	const currentLabel = $derived(
		options.find(opt => opt.value === value)?.label
	);

	// Store the fallback label when value is not in options yet (async loading)
	let fallbackLabel = $state<string | undefined>();

	// Handle value change explicitly to trigger onValueChange prop
	$effect(() => {
		if (value !== undefined) {
			onValueChange?.(value);
			// If value is not in options, try to find its label from previous options
			if (!valueInOptions && !fallbackLabel) {
				// Try to find label from previous render's options
				fallbackLabel = options.find(opt => opt.value === value)?.label;
			}
		} else {
			fallbackLabel = undefined;
		}
	});
</script>

<div class={cn("themed-select-container", className)}>
	<select 
		{disabled} 
		bind:value={value}
		class="custom-select"
	>
		<button aria-label="select" class="select-trigger">
			<selectedcontent></selectedcontent>
		</button>
		
		{#if placeholder && !value}
			<option value="" disabled selected>{placeholder}</option>
		{/if}

		<!-- Fallback option to preserve value and label while async options are loading -->
		{#if value && !valueInOptions}
			<option value={value} hidden>{fallbackLabel || value}</option>
		{/if}

		{#each options as opt}
			<option 
				value={opt.value} 
				style="--color: {opt.color || 'var(--accent-primary)'}"
			>
				{opt.label}
			</option>
		{/each}
	</select>
</div>

<style>
	.themed-select-container {
		display: grid;
		width: 100%;
	}

	.custom-select {
		--time: 0.2s;
		width: 100%;
		padding: 0.6rem 1rem;
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-primary);
		background-color: var(--surface-input);
		color: var(--text-primary);
		font-size: 14px;
		outline: 0;
		cursor: pointer;
		transition: border 0.1s, border-radius 0.1s;
		transition-delay: var(--time);
		appearance: none;

		/* The magic: Customizable Select API */
		&, &::picker(select) {
			appearance: base-select;    
		}

		&:hover {
			border-color: var(--border-hover);
		}

		&:focus {
			border-color: var(--accent-primary);
			box-shadow: 0 0 0 1px var(--accent-primary);
		}

		&:open {
			border-bottom-right-radius: 0;
			border-bottom-left-radius: 0;
			transition-delay: 0s;
		}

		&::picker(select) {
			width: var(--bits-select-trigger-width, 100%);
			background-color: var(--bg-tertiary);
			border: 1px solid var(--border-primary);
			border-top: 0;
			box-shadow: var(--shadow-2xl);
			transform-origin: top;
			transition: clip-path var(--time),
				display var(--time) allow-discrete,
				overlay var(--time) allow-discrete;
			clip-path: polygon(0 0, 100% 0, 100% 0, 0 0);
		}

		&:open::picker(select) {
			clip-path: polygon(0 0, 100% 0, 100% 100%, 0 100%);
			@starting-style {
				clip-path: polygon(0 0, 100% 0, 100% 0, 0 0);
			}
			border-bottom-right-radius: var(--radius-xl);
			border-bottom-left-radius: var(--radius-xl);
		}
	}

	.select-trigger {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: transparent;
		border: none;
		color: inherit;
		font: inherit;
		text-align: left;
		padding-right: 0.5rem;
	}

	/* Seta visual (caret) customizada */
	.select-trigger::after {
		content: "";
		width: 12px;
		height: 12px;
		background-color: var(--text-muted);
		mask: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="black" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>') no-repeat center;
		mask-size: contain;
		-webkit-mask: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="black" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>') no-repeat center;
		-webkit-mask-size: contain;
		flex-shrink: 0;
		opacity: 0.6;
		transition: transform 0.2s ease;
	}

	.custom-select:open .select-trigger::after {
		transform: rotate(180deg);
	}

	option {
		padding: 0.7rem 1rem;
		background-color: transparent;
		color: var(--text-secondary);
		transition: all 0.2s;
		cursor: pointer;
	}

	option:checked {
		background-color: var(--accent-primary);
		color: var(--accent-primary-fg);
	}

	option:hover, option:focus-visible {
		background-color: rgba(236, 72, 153, 0.1);
		color: var(--text-primary);
		outline: none;
	}

	option::checkmark {
		display: none;
	}

	select:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>

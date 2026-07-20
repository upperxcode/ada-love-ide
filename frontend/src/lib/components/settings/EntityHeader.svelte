<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';

	// ── Emoji categories ──
	const emojiCategories = [
		{
			label: 'Faces',
			emojis: ['😀', '😎', '🤖', '🧙', '👻', '🐱', '🦊', '🐸'],
		},
		{
			label: 'Objects',
			emojis: ['📄', '📋', '📁', '🔧', '🛠️', '⚙️', '🧩', '🔌'],
		},
		{
			label: 'Symbols',
			emojis: ['✨', '💡', '🔥', '⚡', '💎', '🌟', '🎯', '🚀'],
		},
		{
			label: 'Nature',
			emojis: ['🌈', '🌊', '🌿', '🌳', '🏔️', '🌙', '☀️', '❄️'],
		},
		{
			label: 'Tech',
			emojis: ['🧠', '💻', '🖥️', '📱', '🌐', '📡', '💾', '🔑'],
		},
	];

	interface EntityHeaderProps {
		icon: string;
		color: string;
		entityType: string;
		isNew: boolean;
		onIconChange: (emoji: string) => void;
		onColorChange: (color: string) => void;
		onClose: () => void;
	}

	let {
		icon,
		color,
		entityType,
		isNew,
		onIconChange,
		onColorChange,
		onClose,
	}: EntityHeaderProps = $props();

	let emojiPickerOpen = $state(false);
	let colorInputEl: HTMLInputElement | null = null;

	function handleIconClick(e: MouseEvent) {
		e.preventDefault();
		emojiPickerOpen = !emojiPickerOpen;
	}

	function handleColorBtnClick() {
		colorInputEl?.click();
	}

	function selectEmoji(emoji: string) {
		onIconChange(emoji);
		emojiPickerOpen = false;
	}

	function handleColorChange(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		onColorChange(val);
	}

	function handleClickOutside(e: MouseEvent) {
		if (!(e.target as HTMLElement).closest('[data-emoji-picker]') &&
			!(e.target as HTMLElement).closest('[data-icon-trigger]')) {
			emojiPickerOpen = false;
		}
	}

	$effect(() => {
		if (emojiPickerOpen) {
			document.addEventListener('click', handleClickOutside, true);
		} else {
			document.removeEventListener('click', handleClickOutside, true);
		}
		return () => document.removeEventListener('click', handleClickOutside, true);
	});
</script>

	<div class="relative">
		<!-- ── Header bar ── -->
		<div
			class="flex items-center gap-3 px-4 py-3 rounded-t-lg"
			style="background: linear-gradient(135deg, {color}88, {color}44)"
		>
			<!-- Icon (single-click opens emoji picker) -->
			<button
				type="button"
				onclick={handleIconClick}
				data-icon-trigger
				class="flex items-center justify-center w-9 h-9 rounded-lg bg-black/10 cursor-pointer hover:bg-black/20 transition-colors"
				title="Click to change icon"
			>
				<span class="text-2xl leading-none select-none">{icon}</span>
			</button>

			<!-- Title -->
			<div class="flex-1 min-w-0">
				<h3
					class="text-sm font-semibold truncate"
					style="color: var(--text-primary)"
				>
					{isNew ? 'New' : 'Edit'} {entityType}
				</h3>
			</div>

			<!-- Color picker button (transparent, after title) -->
			<button
				type="button"
				onclick={handleColorBtnClick}
				data-header-action
				class="flex items-center justify-center w-7 h-7 rounded-md cursor-pointer transition-colors hover:bg-black/20"
				title="Change color"
			>
				<span
					class="w-3.5 h-3.5 rounded-full ring-1 ring-white/20"
					style="background-color: {color}"
				></span>
			</button>

			<!-- Close button -->
			<button
				type="button"
				data-header-action
				onclick={onClose}
				class={cn(
					'flex items-center justify-center w-7 h-7 rounded-md',
					'transition-colors cursor-pointer',
					'hover:bg-black/20'
				)}
				style="color: var(--text-primary)"
				title="Close"
			>
				<Icon name="x" size={15} />
			</button>
		</div>

		<!-- ── Hidden color input ── -->
		<input
			type="color"
			bind:this={colorInputEl}
			value={color}
			onchange={handleColorChange}
			class="sr-only"
		/>

		<!-- ── Emoji Picker popover ── -->
		{#if emojiPickerOpen}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				data-emoji-picker
				class="absolute top-full left-2 z-50 mt-1 w-64 rounded-lg border border-[var(--border-primary)] bg-[var(--bg-tertiary)] shadow-xl p-2 animate-in fade-in-0 zoom-in-95"
			>
				{#each emojiCategories as cat}
					<p class="text-[9px] font-semibold uppercase tracking-wider px-1 pt-1 pb-0.5" style="color: var(--text-faint)">
						{cat.label}
					</p>
					<div class="grid grid-cols-8 gap-0.5">
						{#each cat.emojis as emoji}
							<button
								type="button"
								onclick={() => selectEmoji(emoji)}
								class="flex items-center justify-center w-7 h-7 rounded-md cursor-pointer hover:bg-[var(--accent-primary)]/15 transition-colors"
								title={emoji}
							>
								<span class="text-base leading-none">{emoji}</span>
							</button>
						{/each}
					</div>
				{/each}
			</div>
		{/if}
	</div>

<script lang="ts">
	import { cn } from '$lib/utils';
	import { Icon } from '$lib/components/icon';
	import ThemedSelect from '$lib/components/ui/Select.svelte';
	import { theme } from '$lib/stores/theme.svelte';
	import SettingRow from './SettingRow.svelte';

	// ── Derived options arrays for Select ──
	const colorThemeOptions = $derived(
		theme.availableColorThemes.map((ct) => ({ value: ct.id, label: ct.name }))
	);

	const iconThemeOptions = $derived(
		theme.availableIconThemes.map((it) => ({ value: it.id, label: it.name }))
	);

	const fontThemeOptions = $derived(
		theme.availableFontThemes.map((ft) => ({ value: ft.id, label: ft.name }))
	);

	const fontSizeOptions = $derived(
		theme.availableFontSizePresets.map((fs) => ({
			value: fs.id,
			label: fs.name,
		}))
	);

	const styleThemeOptions = $derived(
		theme.availableStyleThemes.map((st) => ({
			value: st.id,
			label: st.name,
			description: st.description,
		}))
	);
</script>

<div class="flex flex-col gap-3 p-4 overflow-y-auto flex-1">

	<!-- ═══════════════════════════════════════════════════════════════
	     Card: System (app-wide theming)
	     ═══════════════════════════════════════════════════════════════ -->
	<div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] overflow-hidden">
		<div class="px-4 py-3 border-b border-[var(--border-primary)] flex items-center gap-2">
			<div class="flex items-center justify-center w-5 h-5 rounded" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
				<Icon name="monitor" size={12} />
			</div>
			<h4 class="text-xs font-semibold tracking-wider" style="color: var(--text-secondary)">SYSTEM</h4>
		</div>
		<div class="divide-y divide-[var(--border-primary)]">

			<!-- Row: Color Theme -->
			<div class="px-4">
				<SettingRow
					label="Color Theme"
					description="Base color palette for the entire interface"
				>
					<ThemedSelect
						value={theme.colorThemeId}
						onValueChange={(v) => theme.setColorTheme(v)}
						options={colorThemeOptions}
						placeholder="Select theme"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Vibrance -->
			<div class="px-4">
				<SettingRow
					label="Vibrance"
					description="Adjust the color saturation intensity"
				>
					<div class="flex items-center gap-2.5">
						<input
							type="range"
							value={theme.vibrance}
							min={0}
							max={200}
							step={5}
							class={cn(
								'w-28 h-1.5 rounded-full appearance-none cursor-pointer',
								'bg-[var(--bg-elevated)]',
								'[&::-webkit-slider-thumb]:appearance-none',
								'[&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3',
								'[&::-webkit-slider-thumb]:rounded-full',
								'[&::-webkit-slider-thumb]:bg-[var(--text-primary)]',
								'[&::-webkit-slider-thumb]:cursor-pointer',
								'[&::-webkit-slider-thumb]:shadow-sm',
								'[&::-webkit-slider-thumb]:transition-transform',
								'[&::-webkit-slider-thumb]:hover:scale-125',
								'[&::-webkit-slider-thumb]:active:scale-110',
								'[&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3',
								'[&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-0',
								'[&::-moz-range-thumb]:bg-[var(--text-primary)]',
								'[&::-moz-range-thumb]:cursor-pointer'
							)}
							oninput={(e) => theme.setVibrance(Number((e.target as HTMLInputElement).value))}
						/>
						<span class="text-[10px] font-mono tabular-nums w-8 text-right" style="color: var(--text-muted)">
							{theme.vibrance}%
						</span>
					</div>
				</SettingRow>
			</div>

			<!-- Row: Icon Theme -->
			<div class="px-4">
				<SettingRow
					label="Icon Theme"
					description="Switch between icon families for the UI"
				>
					<ThemedSelect
						value={theme.iconThemeId}
						onValueChange={(v) => theme.setIconTheme(v)}
						options={iconThemeOptions}
						placeholder="Select icons"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Font Theme -->
			<div class="px-4">
				<SettingRow
					label="Font Theme"
					description="Choose the typeface for the interface"
				>
					<ThemedSelect
						value={theme.fontThemeId}
						onValueChange={(v) => theme.setFontTheme(v)}
						options={fontThemeOptions}
						placeholder="Select font"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Font Size -->
			<div class="px-4">
				<SettingRow
					label="Font Size"
					description="Set the base text size across the app"
				>
					<ThemedSelect
						value={theme.fontSizeId}
						onValueChange={(v) => theme.setFontSize(v)}
						options={fontSizeOptions}
						placeholder="Select size"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Style Theme -->
			<div class="px-4">
				<SettingRow
					label="Style Theme"
					description="Controls visual style: border radius, shadows, and overall shape"
				>
					<ThemedSelect
						value={theme.styleThemeId}
						onValueChange={(v) => theme.setStyleTheme(v)}
						options={styleThemeOptions}
						placeholder="Select style"
						class="w-36"
					/>
				</SettingRow>
			</div>

		</div>
	</div>

	<!-- ═══════════════════════════════════════════════════════════════
	     Card: Chat (chat-specific theming)
	     ═══════════════════════════════════════════════════════════════ -->
	<div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] overflow-hidden">
		<div class="px-4 py-3 border-b border-[var(--border-primary)] flex items-center gap-2">
			<div class="flex items-center justify-center w-5 h-5 rounded" style="background-color: var(--accent-primary); color: var(--accent-primary-fg)">
				<Icon name="messageSquare" size={12} />
			</div>
			<h4 class="text-xs font-semibold tracking-wider" style="color: var(--text-secondary)">CHAT</h4>
		</div>
		<div class="divide-y divide-[var(--border-primary)]">
			<!-- Row: Chat Color Theme -->
			<div class="px-4">
				<SettingRow
					label="Color Theme"
					description="Color palette exclusive to the chat area"
				>
					<ThemedSelect
						value={theme.chatColorThemeId}
						onValueChange={(v) => theme.setChatColorTheme(v)}
						options={colorThemeOptions}
						placeholder="Select theme"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Chat Font Theme -->
			<div class="px-4">
				<SettingRow
					label="Font Theme"
					description="Typeface used only inside the chat area"
				>
					<ThemedSelect
						value={theme.chatFontThemeId}
						onValueChange={(v) => theme.setChatFontTheme(v)}
						options={fontThemeOptions}
						placeholder="Select font"
						class="w-36"
					/>
				</SettingRow>
			</div>

			<!-- Row: Chat Font Size -->
			<div class="px-4">
				<SettingRow
					label="Font Size"
					description="Text size used only inside the chat area"
				>
					<ThemedSelect
						value={theme.chatFontSizeId}
						onValueChange={(v) => theme.setChatFontSize(v)}
						options={fontSizeOptions}
						placeholder="Select size"
						class="w-36"
					/>
				</SettingRow>
			</div>
		</div>
	</div>

	<!-- ═══════════════════════════════════════════════════════════════
	     Card: Shortcuts (placeholder)
	     ═══════════════════════════════════════════════════════════════ -->
	<div class="rounded-xl border border-[var(--border-primary)] bg-[var(--bg-secondary)] px-4 py-5">
		<div class="flex flex-col items-center gap-3">
			<div
				class="flex items-center justify-center w-10 h-10 rounded-xl opacity-15"
				style="border: 1px solid var(--border-primary)"
			>
				<Icon name="keyboard" size={18} />
			</div>
			<p class="text-xs text-center" style="color: var(--text-faint)">
				Keyboard shortcuts — coming soon
			</p>
		</div>
	</div>
</div>

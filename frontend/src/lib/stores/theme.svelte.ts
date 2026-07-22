// ── Theme Store (Svelte 5 Runes) ──────────────────────────────────
// Central reactive store for all theming concerns:
//   - Color theme (palette swap)
//   - Font theme (family swap)
//   - Font size (scaling)
//   - Vibrance (saturation multiplier)
//   - Icon theme (lucide / material)
//
// Mirrors the Go backend GeneralConfig fields:
//   icon_theme, theme, font_theme, font_size, vibrance

import {
	type ColorTheme,
	type FontTheme,
	type FontSizePreset,
	type StyleTheme,
	type IconThemeDef,
	colorThemes,
	fontThemes,
	fontSizePresets,
	styleThemes,
	iconThemes,
} from '$lib/themes/definitions';

// ── Persistence key ──
const STORAGE_KEY = 'ada-love-ide:theme';

// ── Saved preferences (from localStorage or defaults) ──
interface SavedPreferences {
	colorTheme: string;
	fontTheme: string;
	fontSize: string;
	styleTheme: string;
	iconTheme: string;
	vibrance: number;
	chatColorTheme: string;
	chatFontTheme: string;
	chatFontSize: string;
}

function loadPreferences(): SavedPreferences {
	if (typeof window === 'undefined') {
		return {
			colorTheme: 'zinc',
			fontTheme: 'geist',
			fontSize: 'default',
			styleTheme: 'modern',
			iconTheme: 'lucide',
			vibrance: 100,
			chatColorTheme: 'zinc',
			chatFontTheme: 'geist',
			chatFontSize: 'default',
		};
	}
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) {
			const parsed = JSON.parse(raw) as SavedPreferences;
			// Ensure backward compat: if chat fields are missing, default to app theme values
			if (!parsed.chatColorTheme) parsed.chatColorTheme = parsed.colorTheme;
			if (!parsed.chatFontTheme) parsed.chatFontTheme = parsed.fontTheme;
			if (!parsed.chatFontSize) parsed.chatFontSize = parsed.fontSize;
			if (!parsed.styleTheme) parsed.styleTheme = 'modern';
			return parsed;
		}
	} catch {
		// Ignore parse errors — use defaults
	}
	return {
		colorTheme: 'zinc',
		fontTheme: 'geist',
		fontSize: 'default',
		styleTheme: 'modern',
		iconTheme: 'lucide',
		vibrance: 100,
		chatColorTheme: 'zinc',
		chatFontTheme: 'geist',
		chatFontSize: 'default',
	};
}

function savePreferences(prefs: SavedPreferences) {
	if (typeof window === 'undefined') return;
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
	} catch {
		// Storage full or unavailable
	}
}

// ── Apply CSS variables to <html> root ──
function applyColorTheme(theme: ColorTheme) {
	const root = document.documentElement;
	for (const [key, value] of Object.entries(theme.vars)) {
		root.style.setProperty(key, value);
	}
}

function applyFontTheme(theme: FontTheme) {
	const root = document.documentElement;
	root.style.setProperty('--font-sans', theme.sans);
	root.style.setProperty('--font-mono', theme.mono);
	root.style.setProperty('--font-display', theme.display);
}

function applyFontSize(preset: FontSizePreset) {
	const root = document.documentElement;
	root.style.setProperty('--text-base', preset.base);
	root.style.setProperty('--text-sm', preset.sm);
	root.style.setProperty('--text-xs', preset.xs);
	root.style.setProperty('--text-lg', preset.lg);
	root.style.setProperty('--line-height', preset.lineHeight);
}

function applyVibrance(value: number) {
	// Vibrance controls global saturation via CSS filter on body
	const root = document.documentElement;
	const saturation = value / 100;
	root.style.setProperty('--vibrance-sat', `${saturation}`);
}

// ── Style theme: inject Tailwind overrides that reference CSS variables ──
// Tailwind generates hardcoded values for duration, shadow, border classes.
// We inject a <style> with !important rules so style themes can change them.
function applyStyleTheme(theme: StyleTheme) {
	const root = document.documentElement;
	for (const [key, value] of Object.entries(theme.vars)) {
		root.style.setProperty(key, value);
	}

	const duration = theme.vars['--default-transition-duration'] ?? '0.15s';
	const easing = theme.vars['--default-transition-timing-function'] ?? 'cubic-bezier(0.4, 0, 0.2, 1)';
	const borderW = theme.vars['--border-width'] ?? '1px';
	const sSm = theme.vars['--shadow-sm'] ?? '0 1px 3px 0 rgba(0,0,0,0.1)';
	const sMd = theme.vars['--shadow-md'] ?? '0 4px 6px -1px rgba(0,0,0,0.1)';
	const sLg = theme.vars['--shadow-lg'] ?? '0 10px 15px -3px rgba(0,0,0,0.1)';

	const css = `
.duration-100, .duration-150, .duration-200, .duration-300, .duration-500 { --tw-duration: ${duration} !important; }
.ease-out, .ease-in-out { --tw-ease: ${easing} !important; }
.border, .border-t, .border-r, .border-b, .border-l { border-width: ${borderW} !important; }
.border-2 { border-width: calc(${borderW} * 2) !important; }
.shadow-sm { --tw-shadow: ${sSm} !important; }
.shadow-md { --tw-shadow: ${sMd} !important; }
.shadow-lg { --tw-shadow: ${sLg} !important; }
/* Interactive lift/press — only active when theme sets --hover-lift or --active-press */
button:not(:disabled):hover, [role="button"]:not(:disabled):hover { translate: var(--hover-lift, 0 0); }
button:not(:disabled):active, [role="button"]:not(:disabled):active { translate: var(--active-press, 0 0); }
`;

	let el = document.getElementById('ada-style-override') as HTMLStyleElement | null;
	if (!el) {
		el = document.createElement('style');
		el.id = 'ada-style-override';
		document.head.appendChild(el);
	}
	el.textContent = css;
}

// ── The Store ──
// Exposed as a plain object with $state fields (Svelte 5 runes).
// Consumers import and read/assign directly.

const saved = loadPreferences();

class ThemeStore {
	// ── Active theme IDs ──
	colorThemeId = $state<string>(saved.colorTheme);
	chatColorThemeId = $state<string>(saved.chatColorTheme);
	chatFontThemeId = $state<string>(saved.chatFontTheme);
	chatFontSizeId = $state<string>(saved.chatFontSize);

	// Version counter bumped on every chat theme change.
	// Used as a reactivity bridge across module boundaries.
	chatThemeVersion = $state(0);

	fontThemeId = $state<string>(saved.fontTheme);
	fontSizeId = $state<string>(saved.fontSize);
	styleThemeId = $state<string>(saved.styleTheme);
	iconThemeId = $state<string>(saved.iconTheme);
	vibrance = $state<number>(saved.vibrance);

	// ── Derived: resolved theme objects ──
	get colorTheme(): ColorTheme {
		return colorThemes[this.colorThemeId] ?? colorThemes.zinc;
	}

	get chatColorTheme(): ColorTheme {
		return colorThemes[this.chatColorThemeId] ?? colorThemes.zinc;
	}

	get chatFontTheme(): FontTheme {
		return fontThemes[this.chatFontThemeId] ?? fontThemes.geist;
	}

	get chatFontSizePreset(): FontSizePreset {
		return fontSizePresets[this.chatFontSizeId] ?? fontSizePresets.default;
	}

	/**
	 * Returns a CSS string of all chat theme variables AND actual CSS properties
	 * for inline application on the chat container.
	 *
	 * Includes both the CSS variable definitions AND the properties that consume them
	 * (font-family, font-size, line-height) so they resolve correctly when the chat
	 * container inherits already-computed values from the app theme.
	 */
	get chatThemeStyle(): string {
		const colorVars = Object.entries(this.chatColorTheme.vars)
			.map(([key, value]) => `${key}: ${value}`);
		const font = this.chatFontTheme;
		const fontSize = this.chatFontSizePreset;
		const fontVars = [
			`--font-sans: ${font.sans}`,
			`--font-mono: ${font.mono}`,
			`--font-display: ${font.display}`,
			`--text-base: ${fontSize.base}`,
			`--text-sm: ${fontSize.sm}`,
			`--text-xs: ${fontSize.xs}`,
			`--text-lg: ${fontSize.lg}`,
			`--line-height: ${fontSize.lineHeight}`,
		];
		// Apply actual CSS properties so they resolve with the chat's own CSS variables,
		// overriding the inherited computed values from the app theme.
		const cssProps = [
			`font-family: var(--font-sans)`,
			`font-size: var(--text-base)`,
			`line-height: var(--line-height)`,
		];
		return [...colorVars, ...fontVars, ...cssProps].join('; ');
	}

	get fontTheme(): FontTheme {
		return fontThemes[this.fontThemeId] ?? fontThemes.geist;
	}

	get fontSizePreset(): FontSizePreset {
		return fontSizePresets[this.fontSizeId] ?? fontSizePresets.default;
	}

	get iconTheme(): IconThemeDef {
		return iconThemes[this.iconThemeId] ?? iconThemes.lucide;
	}

	get styleTheme(): StyleTheme {
		return styleThemes[this.styleThemeId] ?? styleThemes.modern;
	}

	// ── Lookup helpers ──
	get availableColorThemes(): ColorTheme[] {
		return Object.values(colorThemes);
	}

	get availableFontThemes(): FontTheme[] {
		return Object.values(fontThemes);
	}

	get availableFontSizePresets(): FontSizePreset[] {
		return Object.values(fontSizePresets);
	}

	get availableStyleThemes(): StyleTheme[] {
		return Object.values(styleThemes);
	}

	get availableIconThemes(): IconThemeDef[] {
		return Object.values(iconThemes);
	}

	// ── Mutators (write + persist + apply) ──
	setColorTheme(id: string) {
		this.colorThemeId = id;
		this._persist();
		applyColorTheme(this.colorTheme);
		// Re-apply style theme so its interaction vars override color theme defaults
		applyStyleTheme(this.styleTheme);
	}

	setChatColorTheme(id: string) {
		this.chatColorThemeId = id;
		this.chatThemeVersion++;
		this._persist();
	}

	setChatFontTheme(id: string) {
		this.chatFontThemeId = id;
		this.chatThemeVersion++;
		this._persist();
	}

	setChatFontSize(id: string) {
		this.chatFontSizeId = id;
		this.chatThemeVersion++;
		this._persist();
	}

	setFontTheme(id: string) {
		this.fontThemeId = id;
		this._persist();
		applyFontTheme(this.fontTheme);
	}

	setFontSize(id: string) {
		this.fontSizeId = id;
		this._persist();
		applyFontSize(this.fontSizePreset);
	}

	setIconTheme(id: string) {
		this.iconThemeId = id;
		this._persist();
		// Icon rendering is handled by the adapter component
	}

	setStyleTheme(id: string) {
		this.styleThemeId = id;
		this._persist();
		applyStyleTheme(this.styleTheme);
	}

	setVibrance(value: number) {
		this.vibrance = Math.max(0, Math.min(200, value));
		this._persist();
		applyVibrance(this.vibrance);
	}

	// ── Initialize (call once in root layout) ──
	init() {
		applyColorTheme(this.colorTheme);
		applyFontTheme(this.fontTheme);
		applyFontSize(this.fontSizePreset);
		applyStyleTheme(this.styleTheme);
		applyVibrance(this.vibrance);
	}

	// ── Hydrate from Wails backend config (for future Wails integration) ──
	hydrateFromBackend(config: {
		theme?: string;
		font_theme?: string;
		font_size?: string;
		icon_theme?: string;
		vibrance?: number;
	}) {
		if (config.theme && colorThemes[config.theme]) {
			this.setColorTheme(config.theme);
		}
		if (config.font_theme && fontThemes[config.font_theme]) {
			this.setFontTheme(config.font_theme);
		}
		if (config.font_size && fontSizePresets[config.font_size]) {
			this.setFontSize(config.font_size);
		}
		if (config.icon_theme && iconThemes[config.icon_theme]) {
			this.setIconTheme(config.icon_theme);
		}
		if (config.vibrance !== undefined) {
			this.setVibrance(config.vibrance);
		}
	}

	// ── Internal ──
	private _persist() {
		savePreferences({
			colorTheme: this.colorThemeId,
			chatColorTheme: this.chatColorThemeId,
			chatFontTheme: this.chatFontThemeId,
			chatFontSize: this.chatFontSizeId,
			fontTheme: this.fontThemeId,
			fontSize: this.fontSizeId,
			styleTheme: this.styleThemeId,
			iconTheme: this.iconThemeId,
			vibrance: this.vibrance,
		});
	}
}

// ── Singleton ──
export const theme = new ThemeStore();

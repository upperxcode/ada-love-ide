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
	type IconThemeDef,
	colorThemes,
	fontThemes,
	fontSizePresets,
	iconThemes,
} from '$lib/themes/definitions';

// ── Persistence key ──
const STORAGE_KEY = 'ada-love-ide:theme';

// ── Saved preferences (from localStorage or defaults) ──
interface SavedPreferences {
	colorTheme: string;
	fontTheme: string;
	fontSize: string;
	iconTheme: string;
	vibrance: number;
}

function loadPreferences(): SavedPreferences {
	if (typeof window === 'undefined') {
		return {
			colorTheme: 'zinc',
			fontTheme: 'geist',
			fontSize: 'default',
			iconTheme: 'lucide',
			vibrance: 100,
		};
	}
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) {
			return JSON.parse(raw) as SavedPreferences;
		}
	} catch {
		// Ignore parse errors — use defaults
	}
	return {
		colorTheme: 'zinc',
		fontTheme: 'geist',
		fontSize: 'default',
		iconTheme: 'lucide',
		vibrance: 100,
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

// ── The Store ──
// Exposed as a plain object with $state fields (Svelte 5 runes).
// Consumers import and read/assign directly.

const saved = loadPreferences();

class ThemeStore {
	// ── Active theme IDs ──
	colorThemeId = $state<string>(saved.colorTheme);
	fontThemeId = $state<string>(saved.fontTheme);
	fontSizeId = $state<string>(saved.fontSize);
	iconThemeId = $state<string>(saved.iconTheme);
	vibrance = $state<number>(saved.vibrance);

	// ── Derived: resolved theme objects ──
	get colorTheme(): ColorTheme {
		return colorThemes[this.colorThemeId] ?? colorThemes.zinc;
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

	get availableIconThemes(): IconThemeDef[] {
		return Object.values(iconThemes);
	}

	// ── Mutators (write + persist + apply) ──
	setColorTheme(id: string) {
		this.colorThemeId = id;
		this._persist();
		applyColorTheme(this.colorTheme);
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
			fontTheme: this.fontThemeId,
			fontSize: this.fontSizeId,
			iconTheme: this.iconThemeId,
			vibrance: this.vibrance,
		});
	}
}

// ── Singleton ──
export const theme = new ThemeStore();

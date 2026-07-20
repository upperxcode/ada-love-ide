// ── Color Theme Definitions ───────────────────────────────────────
// Each color theme defines the full palette via CSS custom properties.
// All values are valid CSS color strings applied at :root level.
// Themes can be swapped at runtime — components consume var(--xxx).

export interface ColorTheme {
	id: string;
	name: string;
	vars: Record<string, string>;
}

export const colorThemes: Record<string, ColorTheme> = {
	// ── Zinc (default dark) ──
	zinc: {
		id: 'zinc',
		name: 'Zinc',
		vars: {
			'--bg-primary': '#09090b',       // zinc-950
			'--bg-secondary': '#18181b',     // zinc-900
			'--bg-tertiary': '#27272a',      // zinc-800
			'--bg-elevated': '#3f3f46',      // zinc-700

			'--border-primary': '#27272a',   // zinc-800
			'--border-subtle': '#1c1c1f',    // zinc-900/zinc-950 blend
			'--border-hover': '#3f3f46',     // zinc-700

			'--text-primary': '#fafafa',     // zinc-50
			'--text-secondary': '#a1a1aa',   // zinc-400
			'--text-muted': '#71717a',       // zinc-500
			'--text-faint': '#52525b',       // zinc-600

			'--accent-primary': '#ec4899',   // pink-500
			'--accent-primary-fg': '#09090b',// zinc-950
			'--accent-secondary': '#a855f7', // purple-500
			'--accent-glow': 'rgba(236, 72, 153, 0.15)',

			'--surface-input': '#0f0f11',
			'--surface-hover': '#18181b',
			'--surface-active': '#27272a',

			'--status-success': '#22c55e',
			'--status-warning': '#eab308',
			'--status-error': '#ef4444',
			'--status-info': '#3b82f6',
		},
	},

	// ── Slate (softer, blue-gray undertone) ──
	slate: {
		id: 'slate',
		name: 'Slate',
		vars: {
			'--bg-primary': '#0f172a',       // slate-950
			'--bg-secondary': '#1e293b',     // slate-800
			'--bg-tertiary': '#334155',      // slate-700
			'--bg-elevated': '#475569',      // slate-600

			'--border-primary': '#334155',   // slate-700
			'--border-subtle': '#1e293b',    // slate-800
			'--border-hover': '#475569',     // slate-600

			'--text-primary': '#f8fafc',     // slate-50
			'--text-secondary': '#94a3b8',   // slate-400
			'--text-muted': '#64748b',       // slate-500
			'--text-faint': '#475569',       // slate-600

			'--accent-primary': '#f472b6',   // pink-400
			'--accent-primary-fg': '#0f172a',// slate-950
			'--accent-secondary': '#818cf8', // indigo-400
			'--accent-glow': 'rgba(244, 114, 182, 0.15)',

			'--surface-input': '#0f172a',
			'--surface-hover': '#1e293b',
			'--surface-active': '#334155',

			'--status-success': '#4ade80',
			'--status-warning': '#facc15',
			'--status-error': '#f87171',
			'--status-info': '#60a5fa',
		},
	},

	// ── Ember (warm dark with orange/amber accent) ──
	ember: {
		id: 'ember',
		name: 'Ember',
		vars: {
			'--bg-primary': '#0c0a09',       // stone-950
			'--bg-secondary': '#1c1917',     // stone-900
			'--bg-tertiary': '#292524',      // stone-800
			'--bg-elevated': '#44403c',      // stone-700

			'--border-primary': '#292524',   // stone-800
			'--border-subtle': '#1c1917',
			'--border-hover': '#44403c',     // stone-700

			'--text-primary': '#fafaf9',     // stone-50
			'--text-secondary': '#a8a29e',   // stone-400
			'--text-muted': '#78716c',       // stone-500
			'--text-faint': '#57534e',       // stone-600

			'--accent-primary': '#f97316',   // orange-500
			'--accent-primary-fg': '#0c0a09',
			'--accent-secondary': '#fbbf24', // amber-400
			'--accent-glow': 'rgba(249, 115, 22, 0.15)',

			'--surface-input': '#0c0a09',
			'--surface-hover': '#1c1917',
			'--surface-active': '#292524',

			'--status-success': '#22c55e',
			'--status-warning': '#f59e0b',
			'--status-error': '#ef4444',
			'--status-info': '#38bdf8',
		},
	},

	// ── Midnight (deep navy with cyan accent) ──
	midnight: {
		id: 'midnight',
		name: 'Midnight',
		vars: {
			'--bg-primary': '#0a0e1a',
			'--bg-secondary': '#111827',     // gray-900
			'--bg-tertiary': '#1f2937',      // gray-800
			'--bg-elevated': '#374151',      // gray-700

			'--border-primary': '#1f2937',
			'--border-subtle': '#111827',
			'--border-hover': '#374151',

			'--text-primary': '#f3f4f6',     // gray-100
			'--text-secondary': '#9ca3af',   // gray-400
			'--text-muted': '#6b7280',       // gray-500
			'--text-faint': '#4b5563',       // gray-600

			'--accent-primary': '#06b6d4',   // cyan-500
			'--accent-primary-fg': '#0a0e1a',
			'--accent-secondary': '#8b5cf6', // violet-500
			'--accent-glow': 'rgba(6, 182, 212, 0.15)',

			'--surface-input': '#0a0e1a',
			'--surface-hover': '#111827',
			'--surface-active': '#1f2937',

			'--status-success': '#34d399',
			'--status-warning': '#fbbf24',
			'--status-error': '#fb7185',
			'--status-info': '#22d3ee',
		},
	},
};

// ── Font Theme Definitions ─────────────────────────────────────────

export interface FontTheme {
	id: string;
	name: string;
	sans: string;
	mono: string;
	display: string;
	preload: { url: string; as: string }[];
}

export const fontThemes: Record<string, FontTheme> = {
	geist: {
		id: 'geist',
		name: 'Geist',
		sans: "'Geist Variable', sans-serif",
		mono: "'Geist Mono Variable', monospace",
		display: "'Geist Variable', sans-serif",
		preload: [],
	},
	inter: {
		id: 'inter',
		name: 'Inter',
		sans: "'Inter', sans-serif",
		mono: "'JetBrains Mono', monospace",
		display: "'Inter', sans-serif",
		preload: [],
	},
	ibm: {
		id: 'ibm',
		name: 'IBM Plex',
		sans: "'IBM Plex Sans', sans-serif",
		mono: "'IBM Plex Mono', monospace",
		display: "'IBM Plex Serif', serif",
		preload: [],
	},
	fira: {
		id: 'fira',
		name: 'Fira',
		sans: "'Fira Sans', sans-serif",
		mono: "'Fira Code', monospace",
		display: "'Fira Sans', sans-serif",
		preload: [],
	},
};

// ── Font Size Presets ─────────────────────────────────────────────

export interface FontSizePreset {
	id: string;
	name: string;
	base: string;       // --text-base (px)
	sm: string;         // --text-sm
	xs: string;         // --text-xs
	lg: string;         // --text-lg
	lineHeight: string; // --line-height
}

export const fontSizePresets: Record<string, FontSizePreset> = {
	compact: {
		id: 'compact',
		name: 'Compact',
		base: '13px',
		sm: '11px',
		xs: '10px',
		lg: '15px',
		lineHeight: '1.4',
	},
	default: {
		id: 'default',
		name: 'Default',
		base: '14px',
		sm: '12px',
		xs: '11px',
		lg: '16px',
		lineHeight: '1.5',
	},
	comfortable: {
		id: 'comfortable',
		name: 'Comfortable',
		base: '15px',
		sm: '13px',
		xs: '12px',
		lg: '18px',
		lineHeight: '1.6',
	},
	large: {
		id: 'large',
		name: 'Large',
		base: '16px',
		sm: '14px',
		xs: '13px',
		lg: '20px',
		lineHeight: '1.65',
	},
};

// ── Icon Theme Definitions ────────────────────────────────────────
// The frontend uses lucide-svelte directly. The backend renders via Go templates
// (Lucide or Material). This mapping syncs the two worlds.
// When the backend switches icon_theme, the frontend adapter swaps rendering.

export interface IconThemeDef {
	id: string;
	name: string;
	family: 'lucide' | 'material' | 'custom';
}

export const iconThemes: Record<string, IconThemeDef> = {
	lucide: { id: 'lucide', name: 'Lucide', family: 'lucide' },
	material: { id: 'material', name: 'Material', family: 'material' },
};

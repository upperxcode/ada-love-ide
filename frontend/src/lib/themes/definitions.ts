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

			'--text-primary': '#fafafa',     // zinc-50
			'--text-secondary': '#a1a1aa',   // zinc-400
			'--text-muted': '#71717a',       // zinc-500
			'--text-faint': '#52525b',       // zinc-600

			'--accent-primary': '#ec4899',   // pink-500
			'--accent-primary-fg': '#09090b',// zinc-950
			'--accent-secondary': '#a855f7', // purple-500
			'--accent-glow': 'rgba(236, 72, 153, 0.15)',

			'--surface-input': '#0f0f11',

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

			'--text-primary': '#f8fafc',     // slate-50
			'--text-secondary': '#94a3b8',   // slate-400
			'--text-muted': '#64748b',       // slate-500
			'--text-faint': '#475569',       // slate-600

			'--accent-primary': '#f472b6',   // pink-400
			'--accent-primary-fg': '#0f172a',// slate-950
			'--accent-secondary': '#818cf8', // indigo-400
			'--accent-glow': 'rgba(244, 114, 182, 0.15)',

			'--surface-input': '#0f172a',

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

			'--text-primary': '#fafaf9',     // stone-50
			'--text-secondary': '#a8a29e',   // stone-400
			'--text-muted': '#78716c',       // stone-500
			'--text-faint': '#57534e',       // stone-600

			'--accent-primary': '#f97316',   // orange-500
			'--accent-primary-fg': '#0c0a09',
			'--accent-secondary': '#fbbf24', // amber-400
			'--accent-glow': 'rgba(249, 115, 22, 0.15)',

			'--surface-input': '#0c0a09',

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

			'--text-primary': '#f3f4f6',     // gray-100
			'--text-secondary': '#9ca3af',   // gray-400
			'--text-muted': '#6b7280',       // gray-500
			'--text-faint': '#4b5563',       // gray-600

			'--accent-primary': '#06b6d4',   // cyan-500
			'--accent-primary-fg': '#0a0e1a',
			'--accent-secondary': '#8b5cf6', // violet-500
			'--accent-glow': 'rgba(6, 182, 212, 0.15)',

			'--surface-input': '#0a0e1a',

			'--status-success': '#34d399',
			'--status-warning': '#fbbf24',
			'--status-error': '#fb7185',
			'--status-info': '#22d3ee',
		},
	},

	// ── One Dark (Atom editor inspired) ──
	'one-dark': {
		id: 'one-dark',
		name: 'One Dark',
		vars: {
			'--bg-primary': '#282c34',
			'--bg-secondary': '#21252b',
			'--bg-tertiary': '#383d47',
			'--bg-elevated': '#4b5263',

			'--border-primary': '#383d47',
			'--border-subtle': '#21252b',

			'--text-primary': '#abb2bf',
			'--text-secondary': '#828997',
			'--text-muted': '#5c6370',
			'--text-faint': '#4b5263',

			'--accent-primary': '#61afef',   // blue
			'--accent-primary-fg': '#282c34',
			'--accent-secondary': '#c678dd', // purple
			'--accent-glow': 'rgba(97, 175, 239, 0.15)',

			'--surface-input': '#1e2229',

			'--status-success': '#98c379',
			'--status-warning': '#e5c07b',
			'--status-error': '#e06c75',
			'--status-info': '#56b6c2',
		},
	},

	// ── Dracula ──
	dracula: {
		id: 'dracula',
		name: 'Dracula',
		vars: {
			'--bg-primary': '#282a36',
			'--bg-secondary': '#1e1f29',
			'--bg-tertiary': '#343746',
			'--bg-elevated': '#44475a',

			'--border-primary': '#343746',
			'--border-subtle': '#1e1f29',

			'--text-primary': '#f8f8f2',
			'--text-secondary': '#c0c0b6',
			'--text-muted': '#908f82',
			'--text-faint': '#6272a4',

			'--accent-primary': '#bd93f9',   // purple
			'--accent-primary-fg': '#282a36',
			'--accent-secondary': '#ff79c6', // pink
			'--accent-glow': 'rgba(189, 147, 249, 0.15)',

			'--surface-input': '#1e1f29',

			'--status-success': '#50fa7b',
			'--status-warning': '#f1fa8c',
			'--status-error': '#ff5555',
			'--status-info': '#8be9fd',
		},
	},

	// ── Nord ──
	nord: {
		id: 'nord',
		name: 'Nord',
		vars: {
			'--bg-primary': '#2e3440',       // nord-0
			'--bg-secondary': '#3b4252',     // nord-1
			'--bg-tertiary': '#434c5e',      // nord-2
			'--bg-elevated': '#4c566a',      // nord-3

			'--border-primary': '#434c5e',   // nord-2
			'--border-subtle': '#3b4252',    // nord-1

			'--text-primary': '#eceff4',     // nord-4
			'--text-secondary': '#d8dee9',   // nord-4
			'--text-muted': '#81a1c1',       // nord-8
			'--text-faint': '#616e88',       // nord-3

			'--accent-primary': '#88c0d0',   // nord-7 (frost)
			'--accent-primary-fg': '#2e3440',
			'--accent-secondary': '#b48ead', // nord-15 (aurora purple)
			'--accent-glow': 'rgba(136, 192, 208, 0.15)',

			'--surface-input': '#2e3440',

			'--status-success': '#a3be8c',   // nord-14
			'--status-warning': '#ebcb8b',   // nord-13
			'--status-error': '#bf616a',     // nord-11
			'--status-info': '#81a1c1',      // nord-8
		},
	},

	// ── Catppuccin Mocha ──
	catppuccin: {
		id: 'catppuccin',
		name: 'Catppuccin',
		vars: {
			'--bg-primary': '#1e1e2e',       // base
			'--bg-secondary': '#181825',     // mantle
			'--bg-tertiary': '#2a2a3c',      // surface0
			'--bg-elevated': '#363a4f',      // surface1

			'--border-primary': '#363a4f',   // surface1
			'--border-subtle': '#1e1e2e',    // base

			'--text-primary': '#cdd6f4',     // text
			'--text-secondary': '#a6adc8',   // subtext0
			'--text-muted': '#7f849c',       // overlay2
			'--text-faint': '#585b70',       // overlay0

			'--accent-primary': '#cba6f7',   // mauve
			'--accent-primary-fg': '#1e1e2e',
			'--accent-secondary': '#f5c2e7', // pink
			'--accent-glow': 'rgba(203, 166, 247, 0.15)',

			'--surface-input': '#181825',

			'--status-success': '#a6e3a1',   // green
			'--status-warning': '#f9e2af',   // yellow
			'--status-error': '#f38ba8',     // red
			'--status-info': '#89b4fa',      // blue
		},
	},

	// ── Tokyo Night ──
	'tokyo-night': {
		id: 'tokyo-night',
		name: 'Tokyo Night',
		vars: {
			'--bg-primary': '#1a1b26',
			'--bg-secondary': '#16161e',
			'--bg-tertiary': '#24283b',
			'--bg-elevated': '#2f3548',

			'--border-primary': '#24283b',
			'--border-subtle': '#16161e',

			'--text-primary': '#a9b1d6',
			'--text-secondary': '#787c99',
			'--text-muted': '#565a73',
			'--text-faint': '#414868',

			'--accent-primary': '#7aa2f7',   // blue
			'--accent-primary-fg': '#1a1b26',
			'--accent-secondary': '#bb9af7', // purple
			'--accent-glow': 'rgba(122, 162, 247, 0.15)',

			'--surface-input': '#14151e',

			'--status-success': '#9ece6a',
			'--status-warning': '#e0af68',
			'--status-error': '#f7768e',
			'--status-info': '#7dcfff',
		},
	},

	// ── Ayu Dark ──
	'ayu-dark': {
		id: 'ayu-dark',
		name: 'Ayu Dark',
		vars: {
			'--bg-primary': '#0f1419',
			'--bg-secondary': '#131a21',
			'--bg-tertiary': '#1a232e',
			'--bg-elevated': '#27313d',

			'--border-primary': '#1a232e',
			'--border-subtle': '#131a21',

			'--text-primary': '#bfccdb',
			'--text-secondary': '#8c9bb0',
			'--text-muted': '#5c6773',
			'--text-faint': '#4a5568',

			'--accent-primary': '#f29718',   // orange
			'--accent-primary-fg': '#0f1419',
			'--accent-secondary': '#39bae6', // blue
			'--accent-glow': 'rgba(242, 151, 24, 0.15)',

			'--surface-input': '#0d1017',

			'--status-success': '#7fd962',
			'--status-warning': '#dfb645',
			'--status-error': '#ea6c6c',
			'--status-info': '#59c2ff',
		},
	},

	// ── Rose Pine ──
	'rose-pine': {
		id: 'rose-pine',
		name: 'Rose Pine',
		vars: {
			'--bg-primary': '#191724',       // base
			'--bg-secondary': '#1f1d2e',     // surface
			'--bg-tertiary': '#2a273f',      // overlay
			'--bg-elevated': '#393552',      // highlighted

			'--border-primary': '#2a273f',   // overlay
			'--border-subtle': '#1f1d2e',    // surface

			'--text-primary': '#e0def4',     // text
			'--text-secondary': '#908caa',   // muted
			'--text-muted': '#6e6a86',       // faded
			'--text-faint': '#524f67',       // faded+

			'--accent-primary': '#eb6f92',   // love (pink)
			'--accent-primary-fg': '#191724',
			'--accent-secondary': '#c4a7e7', // iris (purple)
			'--accent-glow': 'rgba(235, 111, 146, 0.15)',

			'--surface-input': '#17141f',

			'--status-success': '#9ccfd8',   // pine (teal)
			'--status-warning': '#f6c177',   // gold
			'--status-error': '#eb6f92',     // love
			'--status-info': '#3e8fb0',      // foam
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
	'dm-sans': {
		id: 'dm-sans',
		name: 'DM Sans',
		sans: "'DM Sans', sans-serif",
		mono: "'JetBrains Mono', monospace",
		display: "'DM Sans', sans-serif",
		preload: [],
	},
	'plus-jakarta': {
		id: 'plus-jakarta',
		name: 'Plus Jakarta',
		sans: "'Plus Jakarta Sans', sans-serif",
		mono: "'Fira Code', monospace",
		display: "'Plus Jakarta Sans', sans-serif",
		preload: [],
	},
	'system': {
		id: 'system',
		name: 'System',
		sans: "-apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
		mono: "'SF Mono', 'Cascadia Code', 'Consolas', monospace",
		display: "-apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
		preload: [],
	},
	outfit: {
		id: 'outfit',
		name: 'Outfit',
		sans: "'Outfit', sans-serif",
		mono: "'JetBrains Mono', monospace",
		display: "'Outfit', sans-serif",
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
	'xsmall': {
		id: 'xsmall',
		name: 'X-Small',
		base: '12px',
		sm: '10px',
		xs: '9px',
		lg: '14px',
		lineHeight: '1.3',
	},
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
	'xlarge': {
		id: 'xlarge',
		name: 'X-Large',
		base: '18px',
		sm: '15px',
		xs: '13px',
		lg: '22px',
		lineHeight: '1.7',
	},
};

// ── Style Theme Definitions ────────────────────────────────────────
// Controls visual appearance beyond colors and fonts:
// border radius, shadow intensity, border thickness, etc.
// Applied globally (SYSTEM level) — affects everything including chat.

export interface StyleTheme {
	id: string;
	name: string;
	description: string;
	vars: Record<string, string>;
}

function shadow(y: number, blur: number, spread: number, alpha: number): string {
	return `0 ${y}px ${blur}px ${spread}px rgba(0,0,0,${alpha})`;
}

export const styleThemes: Record<string, StyleTheme> = {
  // ── Modern: balanced, refined, familiar (default) ──
  modern: {
    id: 'modern',
    name: 'Modern',
    description: 'Clean and balanced — the default',
    vars: {
      '--radius': '0.625rem',
      '--border-width': '1px',
      '--default-transition-duration': '0.15s',
      '--default-transition-timing-function': 'cubic-bezier(0.4, 0, 0.2, 1)',
      '--bg-secondary': 'color-mix(in oklab, var(--bg-primary), white 6%)',
      '--bg-tertiary': 'color-mix(in oklab, var(--bg-primary), white 12%)',
      '--bg-elevated': 'color-mix(in oklab, var(--bg-primary), white 18%)',
      '--border-subtle': 'color-mix(in oklab, var(--bg-primary), white 10%)',
      '--border-primary': 'color-mix(in oklab, var(--bg-primary), white 20%)',
      '--surface-hover': 'rgba(255,255,255,0.06)',
      '--surface-active': 'rgba(255,255,255,0.10)',
      '--border-hover': 'rgba(255,255,255,0.18)',
      '--shadow-sm': '0 1px 2px 0 rgba(0,0,0,0.05)',
      '--shadow-md': '0 4px 6px -1px rgba(0,0,0,0.10), 0 2px 4px -2px rgba(0,0,0,0.10)',
      '--shadow-lg': '0 10px 15px -3px rgba(0,0,0,0.10), 0 4px 6px -4px rgba(0,0,0,0.10)',
    },
  },

  // ── Sharp: precise, thin, minimal ──
  sharp: {
    id: 'sharp',
    name: 'Sharp',
    description: 'Minimal radius, thin borders, precise',
    vars: {
      '--radius': '0.125rem',
      '--border-width': '0.5px',
      '--default-transition-duration': '0.1s',
      '--default-transition-timing-function': 'ease',
      '--bg-secondary': 'color-mix(in oklab, var(--bg-primary), white 4%)',
      '--bg-tertiary': 'color-mix(in oklab, var(--bg-primary), white 8%)',
      '--bg-elevated': 'color-mix(in oklab, var(--bg-primary), white 12%)',
      '--border-subtle': 'color-mix(in oklab, var(--bg-primary), white 6%)',
      '--border-primary': 'color-mix(in oklab, var(--bg-primary), white 14%)',
      '--surface-hover': 'rgba(255,255,255,0.03)',
      '--surface-active': 'rgba(255,255,255,0.06)',
      '--border-hover': 'rgba(255,255,255,0.10)',
      '--shadow-sm': '0 1px 2px 0 rgba(0,0,0,0.02)',
      '--shadow-md': '0 2px 4px 0 rgba(0,0,0,0.03)',
      '--shadow-lg': '0 4px 8px 0 rgba(0,0,0,0.04)',
    },
  },

  // ── Flat: zero depth, hierarchy by proximity ──
  flat: {
    id: 'flat',
    name: 'Flat',
    description: 'No depth — hierarchy by subtle brightness shifts only',
    vars: {
      '--radius': '0.375rem',
      '--border-width': '0px',
      '--default-transition-duration': '0.1s',
      '--default-transition-timing-function': 'ease',
      '--bg-secondary': 'color-mix(in oklab, var(--bg-primary), white 3%)',
      '--bg-tertiary': 'color-mix(in oklab, var(--bg-primary), white 6%)',
      '--bg-elevated': 'color-mix(in oklab, var(--bg-primary), white 9%)',
      '--border-subtle': 'transparent',
      '--border-primary': 'transparent',
      '--surface-hover': 'rgba(255,255,255,0.04)',
      '--surface-active': 'rgba(255,255,255,0.08)',
      '--border-hover': 'transparent',
      '--accent-primary': 'color-mix(in oklab, var(--accent-primary), transparent 30%)',
      '--shadow-sm': 'none',
      '--shadow-md': 'none',
      '--shadow-lg': 'none',
    },
  },

  // ── Glass: translucent, frosted, ethereal ──
  glass: {
    id: 'glass',
    name: 'Glass',
    description: 'Glassmorphism — translucent surfaces, soft glow',
    vars: {
      '--radius': '0.75rem',
      '--border-width': '0.5px',
      '--default-transition-duration': '0.2s',
      '--default-transition-timing-function': 'cubic-bezier(0.4, 0, 0.2, 1)',
      '--bg-secondary': 'rgba(255,255,255,0.04)',
      '--bg-tertiary': 'rgba(255,255,255,0.07)',
      '--bg-elevated': 'rgba(255,255,255,0.10)',
      '--border-subtle': 'rgba(255,255,255,0.06)',
      '--border-primary': 'rgba(255,255,255,0.10)',
      '--surface-hover': 'rgba(255,255,255,0.08)',
      '--surface-active': 'rgba(255,255,255,0.14)',
      '--border-hover': 'rgba(255,255,255,0.22)',
      '--accent-primary': 'color-mix(in oklab, var(--accent-primary), transparent 50%)',
      '--accent-primary-fg': 'var(--text-primary)',
      '--accent-glow': 'color-mix(in oklab, var(--accent-primary), transparent 70%)',
      '--shadow-sm': '0 1px 3px 0 rgba(0,0,0,0.04), 0 0 0 1px rgba(255,255,255,0.03)',
      '--shadow-md': '0 4px 12px 0 rgba(0,0,0,0.06), 0 0 0 1px rgba(255,255,255,0.04)',
      '--shadow-lg': '0 10px 24px -4px rgba(0,0,0,0.08), 0 0 0 1px rgba(255,255,255,0.05)',
    },
  },

  // ── Elite: premium, dark, glass-like — inspired by Linear/Vercel ──
  elite: {
    id: 'elite',
    name: 'Elite',
    description: 'Premium dark — glass borders, subtle glow, Linear vibe',
    vars: {
      '--radius': '0.625rem',
      '--border-width': '1px',
      '--default-transition-duration': '0.15s',
      '--default-transition-timing-function': 'cubic-bezier(0.4, 0, 0.2, 1)',
      '--bg-secondary': 'color-mix(in oklab, var(--bg-primary), white 5%)',
      '--bg-tertiary': 'color-mix(in oklab, var(--bg-primary), white 10%)',
      '--bg-elevated': 'color-mix(in oklab, var(--bg-primary), white 16%)',
      '--border-subtle': 'rgba(255,255,255,0.06)',
      '--border-primary': 'rgba(255,255,255,0.10)',
      '--surface-hover': 'rgba(255,255,255,0.05)',
      '--surface-active': 'rgba(255,255,255,0.09)',
      '--border-hover': 'rgba(255,255,255,0.15)',
      '--accent-primary': 'color-mix(in oklab, var(--accent-primary), transparent 25%)',
      '--accent-glow': 'color-mix(in oklab, var(--accent-primary), transparent 50%)',
      '--shadow-sm': '0 1px 3px 0 rgba(0,0,0,0.08), 0 0 0 1px rgba(255,255,255,0.03)',
      '--shadow-md': '0 4px 14px -2px rgba(0,0,0,0.12), 0 0 0 1px rgba(255,255,255,0.04)',
      '--shadow-lg': '0 12px 28px -4px rgba(0,0,0,0.16), 0 0 0 1px rgba(255,255,255,0.05)',
    },
  },

  // ── Clay: soft, sculpted, inflated — Claymorphism ──
  // Features bold solid-offset shadows, hover lift, active press-down,
  // inflated opaque surfaces (no transparency), and minimum 16px radius.
  clay: {
    id: 'clay',
    name: 'Clay',
    description: 'Sculpted 3D clay — inflated, tactile, bold shadows',
    vars: {
      '--radius': '1rem',
      '--border-width': '0px',
      '--default-transition-duration': '0.1s',
      '--default-transition-timing-function': 'ease-out',
      '--bg-secondary': 'color-mix(in oklab, var(--bg-primary), white 14%)',
      '--bg-tertiary': 'color-mix(in oklab, var(--bg-primary), white 24%)',
      '--bg-elevated': 'color-mix(in oklab, var(--bg-primary), white 32%)',
      '--border-subtle': 'transparent',
      '--border-primary': 'transparent',
      '--surface-hover': 'rgba(255,255,255,0.10)',
      '--surface-active': 'rgba(255,255,255,0.18)',
      '--border-hover': 'transparent',
      '--accent-primary': 'color-mix(in oklab, var(--accent-primary), transparent 15%)',
      '--accent-primary-fg': 'var(--text-primary)',
      '--accent-glow': 'rgba(255,255,255,0.10)',
      '--hover-lift': '0 -2px',
      '--active-press': '0 4px',
      '--shadow-sm': '0 4px 0 0 rgba(0,0,0,0.30), 0 10px 16px -2px rgba(0,0,0,0.18)',
      '--shadow-md': '0 6px 0 0 rgba(0,0,0,0.32), 0 16px 24px -4px rgba(0,0,0,0.20)',
      '--shadow-lg': '0 10px 0 0 rgba(0,0,0,0.35), 0 24px 40px -6px rgba(0,0,0,0.22)',
    },
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
// ── Providers Store (Svelte 5 Runes) ──────────────────────────────────
// Loads provider configs from the Wails backend and provides
// derived lists for cascading Provider → Model selects.

import { GetProvidersConfig } from '../../../wailsjs/go/main/App';

export interface ProviderInfo {
	name: string;
	icon: string;
	color: string;
	typeConnection: string;
}

export interface ModelInfo {
	name: string;
	providerName: string;
}

class ProvidersStore {
	// ── Raw data ──
	providersConfig = $state<Record<string, any>>({});
	loaded = $state(false);
	loading = $state(false);

	// ── Derived: provider list ──
	get providers(): ProviderInfo[] {
		return Object.entries(this.providersConfig).map(([name, cfg]) => ({
			name,
			icon: cfg.icon ?? '🔌',
			color: cfg.color ?? '#3f3f46',
			typeConnection: cfg.type_connection ?? 'openai_compatible',
		}));
	}

	// ── Derived: models for a specific provider ──
	getModels(providerName: string): ModelInfo[] {
		const cfg = this.providersConfig[providerName];
		if (!cfg?.models) return [];
		return Object.keys(cfg.models).map((name) => ({
			name,
			providerName,
		}));
	}

	// ── Load from backend ──
	async load() {
		if (this.loaded || this.loading) return;
		this.loading = true;
		try {
			this.providersConfig = await GetProvidersConfig();
			this.loaded = true;
		} catch (e) {
			console.error('[ProvidersStore] Failed to load:', e);
		} finally {
			this.loading = false;
		}
	}

	// ── Force reload ──
	async refresh() {
		this.loaded = false;
		await this.load();
	}
}

export const providersStore = new ProvidersStore();

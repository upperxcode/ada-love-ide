import type { EntityCardData } from '$lib/components/settings/EntityCard.svelte';
import type { FieldConfig } from '$lib/components/settings/EntityEditDialog.svelte';

// ── Wails bindings (typed minimally for what the frontend needs) ──

interface WailsApp {
	GetAgents(): Promise<any[]>;
	SetAgents(list: any[]): Promise<void>;
	GetWorkers(): Promise<any[]>;
	SetWorkers(list: any[]): Promise<void>;
	GetSkills(): Promise<any[]>;
	InstallSkill(registryName: string, slug: string, version: string): Promise<void>;
	GetInstalledSkills(): Promise<string[]>;
		GetSkillFullInfo(name: string): Promise<any>;
		UpdateSkillConfig(cfg: any): Promise<void>;
		UninstallSkill(name: string): Promise<void>;
	SaveCustomSkill(name: string, description: string, tagsCSV: string, content: string): Promise<void>;
	GetWorkspaces(): Promise<any[]>;
	SaveWorkspace(ws: any): Promise<void>;
	DeleteWorkspace(path: string): Promise<void>;
	GetSpecWizards(): Promise<any[]>;
	SaveSpecWizard(w: any): Promise<void>;
	DeleteSpecWizard(id: string): Promise<void>;
	GetToolProfiles(): Promise<any[]>;
	CreateToolProfile(name: string, color: string, icon: string): Promise<any>;
	DeleteToolProfile(id: number): Promise<boolean>;
	GetProvidersConfig(): Promise<Record<string, any>>;
	SaveDBProvider(name: string, cfg: any): Promise<void>;
		DeleteDBProvider(name: string): Promise<void>;
		GetAdaConfig(): Promise<any>;
		SetAdaConfig(cfg: any): Promise<void>;
		TestMCPConnection(name: string, command: string, url: string, args: string[]): Promise<any>;
	}

function getApp(): WailsApp {
	return (window as any).go?.main?.App ?? ({} as WailsApp);
}

// ── Entity type definitions ──

interface AgentConfig {
	id: number;
	name: string;
	description: string;
	provider: string;
	model: string;
	type: string;
	icon: string;
	color: string;
	max_iterations: number;
	temperature: number;
	system_prompt: string;
}

interface WorkerConfig {
	name: string;
	persona: string;
	language: string;
	icon: string;
	color: string;
	connection_type: string;
	connection_name: string;
	connection_config: string;
	inherit_folders: boolean;
	inherit_knowledge: boolean;
	inherit_skills: boolean;
	inherit_tools: boolean;
	inherit_persona: boolean;
}

interface SkillConfig {
	id: number;
	name: string;
	description: string;
	tags: string;
	content: string;
	color: string;
	icon: string;
	active: boolean;
}

interface WorkspaceConfig {
	id: number;
	title: string;
	description: string;
	path: string;
	folders: string[];
	personality: string;
	routing_rules: string;
	knowledge: string[];
	worker_names: string[];
	skills: string[];
	tools: string[];
	enabled: boolean;
	color: string;
	icon: string;
	max_prompt_send: number;
	commit_changes: boolean;
	max_context_length: number;
	agents: string[];
}

interface SpecWizardConfig {
	id: string;
	name: string;
	description: string;
	color: string;
	icon: string;
}

interface ToolProfile {
	id: number;
	name: string;
	color: string;
	icon: string;
	tools: string[];
}

interface ProviderConfig {
	icon: string;
	color: string;
	api_url: string;
	type_connection: string;
	models: Record<string, any>;
}

// ── Field configs per entity type ──

export const FIELD_CONFIGS: Record<string, FieldConfig[]> = {
	agents: [
			{ key: 'name', label: 'Name', description: 'Identifier for this agent', type: 'text', required: true, placeholder: 'Agent name' },
			{ key: 'description', label: 'Description', description: 'Brief summary of what this agent does', type: 'text', placeholder: 'What does this agent do?' },
			{ key: 'provider', label: 'Provider', description: 'AI service provider', type: 'provider_select' },
			{ key: 'model', label: 'Model', description: 'Model identifier to use', type: 'model_select' },
				{ key: 'type', label: 'Type', description: 'Agent specialization mode', type: 'select', options: [
					{ label: 'Executor (Code)', value: 'executor' },
					{ label: 'Delegator (Routing)', value: 'delegator' },
					{ label: 'Reviewer (Critique)', value: 'reviewer' },
					{ label: 'Research (Investigation)', value: 'research' },
				]},
				{ key: 'temperature', label: 'Temperature', description: 'Creativity level (0 = precise, 2 = creative)', type: 'number', placeholder: '0.7', decimals: true },
				{ key: 'max_iterations', label: 'Max Iterations', description: 'Maximum reasoning steps per request', type: 'number', placeholder: '10', decimals: false },
			{ key: 'system_prompt', label: 'System Prompt', description: 'Instructions that define the agent behavior', type: 'textarea', placeholder: 'You are a helpful assistant...', fullWidth: true, expandable: true },
		],
		workers: [
			{ key: 'name', label: 'Name', description: 'Worker identifier', type: 'text', required: true, placeholder: 'Worker name' },
			{ key: 'persona', label: 'Persona', description: 'Worker behavior instructions', type: 'textarea', placeholder: 'You are a helpful worker...', fullWidth: true, expandable: true },
			{ key: 'language', label: 'Response Language', description: 'Instruct the model to answer in this language', type: 'select', options: [
				{ label: 'Português', value: 'portuguese' },
				{ label: 'English', value: 'english' },
				{ label: 'Español', value: 'spanish' },
				{ label: 'Others', value: 'others' },
			]},
			{ key: 'connection_type', label: 'Connection Type', description: 'Underlying bridge protocol', type: 'select', options: [
				{ label: 'Ada', value: 'ada' },
				{ label: 'CLI', value: 'cli' },
				{ label: 'Url/Api', value: 'url' },
			]},
			{ key: 'inherit_folders', label: 'Inherit Folders', description: 'Inherit workspace file structure', type: 'toggle' },
			{ key: 'inherit_knowledge', label: 'Inherit Knowledge', description: 'Inherit local knowledge base', type: 'toggle' },
			{ key: 'inherit_skills', label: 'Inherit Skills', description: 'Inherit system skills/commands', type: 'toggle' },
			{ key: 'inherit_tools', label: 'Inherit Tools', description: 'Inherit external tool definitions', type: 'toggle' },
			{ key: 'inherit_persona', label: 'Inherit Persona', description: 'Combine with global persona settings', type: 'toggle' },
		],
		skills: [
			{ key: 'name', label: 'Name', description: 'Unique identifier for this skill', type: 'text', required: true, placeholder: 'Skill name' },
			{ key: 'description', label: 'Description', description: 'Brief summary of what this skill does', type: 'text', placeholder: 'What does this skill do?' },
			{ key: 'tags', label: 'Tags', description: 'Comma-separated keywords', type: 'text', placeholder: 'comma, separated, tags' },
			{ key: 'active', label: 'Active', description: 'Whether this skill is enabled', type: 'toggle' },
			{ key: 'content', label: 'Content', description: 'Core logic/instructions of the skill', type: 'textarea', placeholder: 'Skill prompt/instructions...', fullWidth: true, expandable: true },
		],
		workspaces: [
			{ key: 'title', label: 'Title', description: 'Display name of the workspace', type: 'text', required: true, placeholder: 'Workspace title' },
			{ key: 'path', label: 'Path', description: 'Local filesystem path', type: 'text', placeholder: '/path/to/project' },
			{ key: 'description', label: 'Description', description: 'Brief summary of the project', type: 'text', placeholder: 'Workspace description' },
			{ key: 'personality', label: 'Personality', description: 'Custom traits for AI in this context', type: 'textarea', placeholder: 'AI personality traits', fullWidth: true, expandable: true },
			{ key: 'enabled', label: 'Enabled', description: 'Enable or disable this workspace', type: 'toggle' },
		],
		'spec-wizard': [
			{ key: 'name', label: 'Name', description: 'Wizard identifier', type: 'text', required: true, placeholder: 'Spec Wizard name' },
			{ key: 'description', label: 'Description', description: 'Detailed instructions for this wizard', type: 'textarea', placeholder: 'What is this spec for?', fullWidth: true, expandable: true },
		],
		tools: [
			{ key: 'name', label: 'Name', description: 'Tool profile identifier', type: 'text', required: true, placeholder: 'Tool profile name' },
		],
			models: [
				{ key: 'name', label: 'Provider Name', description: 'Unique identifier for this provider', type: 'text', required: true, placeholder: 'openai, anthropic, ollama...' },
				{ key: 'api_url', label: 'API URL', description: 'Base URL for the provider API', type: 'text', placeholder: 'https://api.example.com/v1' },
				{ key: 'type_connection', label: 'Connection Type', description: 'Underlying protocol/provider', type: 'select', options: [
					{ label: 'OpenAI Compatible', value: 'openai_compatible' },
					{ label: 'Anthropic', value: 'anthropic' },
					{ label: 'Ollama', value: 'ollama' },
					{ label: 'Custom', value: 'custom' },
				]},
				{ key: 'strategy', label: 'Rotation Strategy', description: 'How to rotate between multiple API keys', type: 'select', options: [
					{ label: 'Simple Rotate (Round Robin)', value: 'simple_rotate' },
					{ label: 'Hard Caps (Quota-based)', value: 'hard_caps' },
					{ label: 'Load Balancing (Weighted)', value: 'load_balancing' },
					{ label: 'Tiered Rotation (Priority)', value: 'tiered' },
					{ label: 'Rate Limit Evasion (429-based)', value: 'rate_limit' },
				]},
				{ key: 'api_keys', label: 'API Keys', description: 'Manage multiple keys for rotation', type: 'textarea', fullWidth: true },
				{ key: 'models', label: 'Models', description: 'List of supported models for this provider', type: 'textarea', fullWidth: true },
			],
		mcp: [
			{ key: 'nome', label: 'Name', description: 'Server identifier', type: 'text', required: true, placeholder: 'My MCP Server' },
			{ key: 'connect_type', label: 'Connect Type', description: 'Underlying protocol (stdio, sse)', type: 'select', options: [
				{ label: 'STDIO', value: 'stdio' },
				{ label: 'SSE', value: 'sse' },
			]},
			{ key: 'command', label: 'Command', description: 'Executable command (npx, uvx, node...)', type: 'text', placeholder: 'npx' },
			{ key: 'arguments', label: 'Arguments', description: 'Command line arguments', type: 'text', placeholder: '-y @mcp-server/example' },
			{ key: 'url', label: 'URL', description: 'Remote server endpoint (for SSE)', type: 'text', placeholder: 'https://...' },
			{ key: 'enabled', label: 'Enabled', description: 'Whether this MCP server is active', type: 'toggle' },
			{ key: 'timeout', label: 'Timeout', description: 'Response timeout in seconds', type: 'number', placeholder: '30' },
			{ key: 'oauth_client_id', label: 'OAuth Client ID', description: 'Optional OAuth2 client identifier', type: 'text', placeholder: 'client_...' },
			{ key: 'environment', label: 'Environment', description: 'Custom environment variables (JSON)', type: 'textarea', placeholder: '{"API_KEY": "..."}', fullWidth: true, expandable: true },
		],
};

// ── Map any backend entity to EntityCardData ──

function toCardData(raw: any, nameKey = 'name', idKey = 'id'): EntityCardData {
	// Svelte 5 each block requires unique non-null keys.
	let id = raw[idKey] ?? raw.id;

	// If ID is missing, 0, or '0', generate a unique one
	if (id === undefined || id === null || id === 0 || id === '0') {
		const nameBase = raw[nameKey] ?? raw.name ?? raw.title ?? raw.nome ?? 'item';
		id = `${nameBase}-${Math.random().toString(36).substring(2, 9)}`;
	}

	// Handle MCP specific mappings for the frontend
	if (raw.url !== undefined) {
		// Based on Go logic: if URL is not empty, it's SSE, otherwise STDIO
		if (!raw.connect_type) {
			raw.connect_type = raw.url ? 'sse' : 'stdio';
		}
	}

	if (raw.args && !raw.arguments) {
		raw.arguments = Array.isArray(raw.args) ? raw.args.join(' ') : String(raw.args);
	}
	if (raw.env && !raw.environment) {
		raw.environment = JSON.stringify(raw.env);
	}

	// Handle Provider specific mappings
	if (raw.api_keys && !raw.api_keys_raw) {
		const keys = Array.isArray(raw.api_keys) 
			? raw.api_keys.map((k: any) => (typeof k === 'object' ? k.key : k))
			: [];
		raw.api_keys = JSON.stringify(keys);
	}

	return {
		id,
		name: raw[nameKey] ?? raw.title ?? raw.nome ?? raw.name ?? 'Untitled',
		description: raw.description ?? '',
		icon: raw.icon ?? '📄',
		color: raw.color ?? '#3f3f46',
		...raw,
	};
}

// ── Entity Store ──

class EntityStore {
	// Reactive state per entity type
	agents = $state<EntityCardData[]>([]);
	workers = $state<EntityCardData[]>([]);
	skills = $state<EntityCardData[]>([]);
	workspaces = $state<EntityCardData[]>([]);
	specWizards = $state<EntityCardData[]>([]);
	toolProfiles = $state<EntityCardData[]>([]);
	providers = $state<EntityCardData[]>([]);
	mcpServers = $state<EntityCardData[]>([]);

	loading = $state<Record<string, boolean>>({});

	// ── Loaders ──

	async loadAgents() {
		this.loading.agents = true;
		try {
			const list = await getApp().GetAgents();
			this.agents = list.map((a: any) => toCardData(a));
		} catch (e) {
			console.error('[EntityStore] Failed to load agents:', e);
		} finally {
			this.loading.agents = false;
		}
	}

	async loadWorkers() {
		this.loading.workers = true;
		try {
			const list = await getApp().GetWorkers();
			this.workers = list.map((w: any) => toCardData(w, 'name', 'id'));
		} catch (e) {
			console.error('[EntityStore] Failed to load workers:', e);
		} finally {
			this.loading.workers = false;
		}
	}

	async loadSkills() {
		this.loading.skills = true;
		try {
			const list = await getApp().GetSkills();
			this.skills = list.map((s: any) => toCardData(s));
		} catch (e) {
			console.error('[EntityStore] Failed to load skills:', e);
		} finally {
			this.loading.skills = false;
		}
	}

	async loadWorkspaces() {
		this.loading.workspaces = true;
		try {
			const list = await getApp().GetWorkspaces();
			this.workspaces = list.map((w: any) => toCardData(w, 'title', 'id'));
		} catch (e) {
			console.error('[EntityStore] Failed to load workspaces:', e);
		} finally {
			this.loading.workspaces = false;
		}
	}

	async loadSpecWizards() {
		this.loading['spec-wizard'] = true;
		try {
			const list = await getApp().GetSpecWizards();
			// Ensure each spec has a unique ID from the backend
			this.specWizards = list.map((w: any) => toCardData(w, 'name', 'id'));
		} catch (e) {
			console.error('[EntityStore] Failed to load spec wizards:', e);
		} finally {
			this.loading['spec-wizard'] = false;
		}
	}

	async loadToolProfiles() {
		this.loading.tools = true;
		try {
			const list = await getApp().GetToolProfiles();
			this.toolProfiles = list.map((t: any) => toCardData(t, 'name', 'id'));
		} catch (e) {
			console.error('[EntityStore] Failed to load tool profiles:', e);
		} finally {
			this.loading.tools = false;
		}
	}

	async loadProviders() {
		this.loading.models = true;
		try {
			const map = await getApp().GetProvidersConfig();
			this.providers = Object.entries(map).map(([name, cfg]: [string, any]) =>
				toCardData({ ...cfg, name }, 'name', 'name')
			);
		} catch (e) {
			console.error('[EntityStore] Failed to load providers:', e);
		} finally {
			this.loading.models = false;
		}
	}

			async loadMCPServers() {
				this.loading.mcp = true;
				try {
					const cfg = await getApp().GetAdaConfig();
					// MCP servers come as a map in GetAdaConfig { "name": { server_data } }
					this.mcpServers = Object.entries(cfg.mcp_servers ?? {}).map(([name, s]: [string, any]) =>
						toCardData({ ...s, nome: s.nome || name }, 'nome', 'nome')
					);
				} catch (e) {
					console.error('[EntityStore] Failed to load MCP servers:', e);
				} finally {
					this.loading.mcp = false;
				}
			}

	// ── Get items by category key ──

	getItems(category: string): EntityCardData[] {
		switch (category) {
			case 'agents': return this.agents;
			case 'workers': return this.workers;
			case 'skills': return this.skills;
			case 'workspaces': return this.workspaces;
			case 'spec-wizard': return this.specWizards;
			case 'tools': return this.toolProfiles;
			case 'models': return this.providers;
			case 'mcp': return this.mcpServers;
			default: return [];
		}
	}

	isLoading(category: string): boolean {
		return !!this.loading[category];
	}

	// ── Save handlers ──

	async saveAgent(data: Record<string, any>) {
		let newList;
		if (data.id) {
			// Update existing
			newList = this.agents.map((a) => (a.id === data.id ? { ...a, ...data } : a));
		} else {
			// Create new
			const newAgent = {
				...data,
				id: 0, // Backend uses ID > 0 to decide Update vs Create
			};
			newList = [...this.agents, newAgent];
		}
		await getApp().SetAgents(newList);
		await this.loadAgents();
	}

	async saveWorker(data: Record<string, any>) {
		let newList;
		if (data.id) {
			newList = this.workers.map((w) => (w.id === data.id ? { ...w, ...data } : w));
		} else {
			const newWorker = {
				...data,
				id: 0,
			};
			newList = [...this.workers, newWorker];
		}
		await getApp().SetWorkers(newList);
		await this.loadWorkers();
	}

	async saveSkill(data: Record<string, any>) {
		await getApp().UpdateSkillConfig(data);
		await this.loadSkills();
	}

	async saveWorkspace(data: Record<string, any>) {
		await getApp().SaveWorkspace(data);
		await this.loadWorkspaces();
	}

	async saveSpecWizard(data: Record<string, any>) {
		await getApp().SaveSpecWizard(data);
		await this.loadSpecWizards();
	}

	async saveToolProfile(data: Record<string, any>) {
		const result = await getApp().CreateToolProfile(data.name, data.color, data.icon);
		await this.loadToolProfiles();
	}

	async saveProvider(data: Record<string, any>) {
		// Frontend APIKeys can be JSON string from APIKeyManager
		let apiKeys = [];
		if (typeof data.api_keys === 'string') {
			try {
				const parsed = JSON.parse(data.api_keys);
				apiKeys = Array.isArray(parsed) ? parsed : [data.api_keys];
			} catch {
				apiKeys = [data.api_keys];
			}
		} else if (Array.isArray(data.api_keys)) {
			apiKeys = data.api_keys;
		}

		// Models can be JSON string from ModelListCollapsible
		let models = {};
		if (typeof data.models === 'string') {
			try {
				models = JSON.parse(data.models);
			} catch {
				models = {};
			}
		} else if (data.models && typeof data.models === 'object') {
			models = data.models;
		}

		// Map to ProviderConfig struct for backend
		const providerData = {
			...data,
			api_keys: apiKeys.map(k => (typeof k === 'string' ? { key: k, user_key: '' } : k)),
			models: models
		};

		await getApp().SaveDBProvider(data.name, providerData);
		await this.loadProviders();
	}

	async saveMCPServer(data: Record<string, any>) {
		const cfg = await getApp().GetAdaConfig();
		const mcpServers = cfg.mcp_servers || {};
		
		// Map frontend-friendly fields back to what MCPServerUI expects
		let args: string[] = [];
		if (typeof data.arguments === 'string') {
			args = data.arguments.split(' ').filter(Boolean);
		} else if (Array.isArray(data.args)) {
			args = data.args;
		}

		let env: Record<string, string> = {};
		if (typeof data.environment === 'string') {
			try {
				env = JSON.parse(data.environment);
			} catch {
				env = {};
			}
		} else if (data.env) {
			env = data.env;
		}

		// Ensure we send ONLY what MCPServerUI (Go struct) expects
		// Backend uses 'command' and 'args' for STDIO, and 'url' for SSE.
		// Important: Go backend uses 'url != ""' to decide connect_type is 'url'.
		const serverData = {
			command: data.connect_type === 'stdio' ? (data.command || '') : '',
			args: data.connect_type === 'stdio' ? args : [],
			url: data.connect_type === 'sse' ? (data.url || '') : '',
			env: env,
			enabled: !!data.enabled,
			icon: data.icon || '🔌',
			color: data.color || '#3f3f46',
			timeout: Number(data.timeout || 30),
			oauth_client_id: data.oauth_client_id || '',
			headers: data.headers || []
		};

		mcpServers[data.nome] = serverData;
		await getApp().SetAdaConfig({ ...cfg, mcp_servers: mcpServers });
		await this.loadMCPServers();
	}

	// ── Delete handlers ──

	async deleteAgent(data: Record<string, any>) {
		await getApp().SetAgents(
			this.agents.filter((a) => a.id !== data.id)
		);
		await this.loadAgents();
	}

	async deleteWorker(data: Record<string, any>) {
		await getApp().SetWorkers(
			this.workers.filter((w) => w.id !== data.id)
		);
		await this.loadWorkers();
	}

	async deleteSkill(data: Record<string, any>) {
		await getApp().UninstallSkill(data.name);
		await this.loadSkills();
	}

	async deleteWorkspace(data: Record<string, any>) {
		await getApp().DeleteWorkspace(data.path ?? data.id);
		await this.loadWorkspaces();
	}

	async deleteSpecWizard(data: Record<string, any>) {
		await getApp().DeleteSpecWizard(data.id);
		await this.loadSpecWizards();
	}

	async deleteToolProfile(data: Record<string, any>) {
		await getApp().DeleteToolProfile(data.id);
		await this.loadToolProfiles();
	}

	async deleteProvider(data: Record<string, any>) {
		await getApp().DeleteDBProvider(data.name);
		await this.loadProviders();
	}

	async deleteMCPServer(data: Record<string, any>) {
		const cfg = await getApp().GetAdaConfig();
		const mcpServers = cfg.mcp_servers || {};
		delete mcpServers[data.nome];
		await getApp().SetAdaConfig({ ...cfg, mcp_servers: mcpServers });
		await this.loadMCPServers();
	}

	// ── Generic save/delete dispatcher ──

	async save(category: string, data: Record<string, any>) {
		switch (category) {
			case 'agents': return this.saveAgent(data);
			case 'workers': return this.saveWorker(data);
			case 'skills': return this.saveSkill(data);
			case 'workspaces': return this.saveWorkspace(data);
			case 'spec-wizard': return this.saveSpecWizard(data);
				case 'tools': return this.saveToolProfile(data);
				case 'models': return this.saveProvider(data);
				case 'mcp': return this.saveMCPServer(data);
			}
		}

		async remove(category: string, data: Record<string, any>) {
			switch (category) {
				case 'agents': return this.deleteAgent(data);
				case 'workers': return this.deleteWorker(data);
				case 'skills': return this.deleteSkill(data);
				case 'workspaces': return this.deleteWorkspace(data);
				case 'spec-wizard': return this.deleteSpecWizard(data);
				case 'tools': return this.deleteToolProfile(data);
				case 'models': return this.deleteProvider(data);
				case 'mcp': return this.deleteMCPServer(data);
			}
		}

	// ── Generic load dispatcher ──

	async load(category: string) {
		switch (category) {
			case 'agents': return this.loadAgents();
			case 'workers': return this.loadWorkers();
			case 'skills': return this.loadSkills();
			case 'workspaces': return this.loadWorkspaces();
			case 'spec-wizard': return this.loadSpecWizards();
			case 'tools': return this.loadToolProfiles();
			case 'models': return this.loadProviders();
			case 'mcp': return this.loadMCPServers();
		}
	}

	// ── Load all ──

	async loadAll() {
		await Promise.allSettled([
			this.loadAgents(),
			this.loadWorkers(),
			this.loadSkills(),
			this.loadWorkspaces(),
			this.loadSpecWizards(),
			this.loadToolProfiles(),
			this.loadProviders(),
			this.loadMCPServers(),
		]);
	}
}

export const entityStore = new EntityStore();

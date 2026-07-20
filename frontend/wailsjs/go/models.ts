export namespace ada {
	
	export class AdaConfig {
	    active_workspace_path: string;
	    active_workspace_index: number;
	    workspaces: workspace.WorkspaceConfig[];
	    tiny_brain_provider: string;
	    tiny_brain_model: string;
	    tiny_brain_tools: string[];
	    workers: worker.WorkerConfig[];
	    worker_categories: string[];
	    agents: agent.AgentConfig[];
	    agent_categories: string[];
	    provider_keys: Record<string, string>;
	    provider_bases: Record<string, string>;
	    model_settings: Record<string, any>;
	    model_list: any[];
	    providers: Record<string, provider.ProviderConfig>;
	    embedding_model: string;
	    embedding_provider: string;
	    image_model: string;
	    image_provider: string;
	    spec_model: string;
	    spec_provider: string;
	    spec_tools: string[];
	    mcp_servers: Record<string, mcp.MCPServerUI>;
	    active_session_id: string;
	    sidebar_visible: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AdaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active_workspace_path = source["active_workspace_path"];
	        this.active_workspace_index = source["active_workspace_index"];
	        this.workspaces = this.convertValues(source["workspaces"], workspace.WorkspaceConfig);
	        this.tiny_brain_provider = source["tiny_brain_provider"];
	        this.tiny_brain_model = source["tiny_brain_model"];
	        this.tiny_brain_tools = source["tiny_brain_tools"];
	        this.workers = this.convertValues(source["workers"], worker.WorkerConfig);
	        this.worker_categories = source["worker_categories"];
	        this.agents = this.convertValues(source["agents"], agent.AgentConfig);
	        this.agent_categories = source["agent_categories"];
	        this.provider_keys = source["provider_keys"];
	        this.provider_bases = source["provider_bases"];
	        this.model_settings = source["model_settings"];
	        this.model_list = source["model_list"];
	        this.providers = this.convertValues(source["providers"], provider.ProviderConfig, true);
	        this.embedding_model = source["embedding_model"];
	        this.embedding_provider = source["embedding_provider"];
	        this.image_model = source["image_model"];
	        this.image_provider = source["image_provider"];
	        this.spec_model = source["spec_model"];
	        this.spec_provider = source["spec_provider"];
	        this.spec_tools = source["spec_tools"];
	        this.mcp_servers = this.convertValues(source["mcp_servers"], mcp.MCPServerUI, true);
	        this.active_session_id = source["active_session_id"];
	        this.sidebar_visible = source["sidebar_visible"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace agent {
	
	export class AgentConfig {
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
	
	    static createFrom(source: any = {}) {
	        return new AgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.type = source["type"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.max_iterations = source["max_iterations"];
	        this.temperature = source["temperature"];
	        this.system_prompt = source["system_prompt"];
	    }
	}

}

export namespace command {
	
	export class SubCommandInfo {
	    name: string;
	    description: string;
	    args_usage: string;
	
	    static createFrom(source: any = {}) {
	        return new SubCommandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.args_usage = source["args_usage"];
	    }
	}
	export class CommandInfo {
	    name: string;
	    description: string;
	    usage: string;
	    aliases: string[];
	    sub_commands: SubCommandInfo[];
	
	    static createFrom(source: any = {}) {
	        return new CommandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.usage = source["usage"];
	        this.aliases = source["aliases"];
	        this.sub_commands = this.convertValues(source["sub_commands"], SubCommandInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace core {
	
	export class RawMessage {
	    role: string;
	    content: string;
	    tool_calls: any[];
	    tool_call_id: string;
	    // Go type: time
	    time: any;
	
	    static createFrom(source: any = {}) {
	        return new RawMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.tool_calls = source["tool_calls"];
	        this.tool_call_id = source["tool_call_id"];
	        this.time = this.convertValues(source["time"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Session {
	    id: string;
	    workspace_id: string;
	    worker_name: string;
	    parent_session_id: string;
	    title: string;
	    summary: string;
	    model: string;
	    provider: string;
	    mode: string;
	    thinking: string;
	    messages: RawMessage[];
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    pinned: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspace_id = source["workspace_id"];
	        this.worker_name = source["worker_name"];
	        this.parent_session_id = source["parent_session_id"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.model = source["model"];
	        this.provider = source["provider"];
	        this.mode = source["mode"];
	        this.thinking = source["thinking"];
	        this.messages = this.convertValues(source["messages"], RawMessage);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.pinned = source["pinned"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class MCPServerRegistry {
	    name: string;
	    url: string;
	    category: string;
	    description: string;
	    language: string;
	    stars: number;
	    topics: string[];
	    tags: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPServerRegistry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.language = source["language"];
	        this.stars = source["stars"];
	        this.topics = source["topics"];
	        this.tags = source["tags"];
	    }
	}

}

export namespace mcp {
	
	export class ConnectionDefinition {
	    name: string;
	    type: string;
	    command: string;
	    description: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.command = source["command"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	    }
	}
	export class ConnectionTestResult {
	    success: boolean;
	    message: string;
	    latency_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.latency_ms = source["latency_ms"];
	    }
	}
	export class HeaderEntry {
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class MCPServerUI {
	    command: string;
	    args: string[];
	    env: Record<string, string>;
	    url: string;
	    enabled: boolean;
	    icon: string;
	    color: string;
	    timeout: number;
	    oauth_client_id: string;
	    headers: HeaderEntry[];
	
	    static createFrom(source: any = {}) {
	        return new MCPServerUI(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.url = source["url"];
	        this.enabled = source["enabled"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.timeout = source["timeout"];
	        this.oauth_client_id = source["oauth_client_id"];
	        this.headers = this.convertValues(source["headers"], HeaderEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace provider {
	
	export class ModelSettings {
	    context_size?: number;
	    temperature?: number;
	    max_tokens?: number;
	    top_p?: number;
	    type?: string;
	    vision?: boolean;
	    embedding?: boolean;
	    tools?: boolean;
	    free?: boolean;
	    thinking?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.context_size = source["context_size"];
	        this.temperature = source["temperature"];
	        this.max_tokens = source["max_tokens"];
	        this.top_p = source["top_p"];
	        this.type = source["type"];
	        this.vision = source["vision"];
	        this.embedding = source["embedding"];
	        this.tools = source["tools"];
	        this.free = source["free"];
	        this.thinking = source["thinking"];
	    }
	}
	export class ProviderAPIKey {
	    key: string;
	    user_key: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderAPIKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.user_key = source["user_key"];
	    }
	}
	export class ProviderConfig {
	    icon: string;
	    color: string;
	    api_url: string;
	    api_keys: ProviderAPIKey[];
	    type_connection: string;
	    strategy: string;
	    models: Record<string, ModelSettings>;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.api_url = source["api_url"];
	        this.api_keys = this.convertValues(source["api_keys"], ProviderAPIKey);
	        this.type_connection = source["type_connection"];
	        this.strategy = source["strategy"];
	        this.models = this.convertValues(source["models"], ModelSettings, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderModel {
	    id: string;
	    name: string;
	    vision: boolean;
	    embedding: boolean;
	    tools: boolean;
	    free: boolean;
	    thinking: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.vision = source["vision"];
	        this.embedding = source["embedding"];
	        this.tools = source["tools"];
	        this.free = source["free"];
	        this.thinking = source["thinking"];
	    }
	}
	export class ProviderTestResult {
	    ok: boolean;
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}

}

export namespace skill {
	
	export class SearchResult {
	    name: string;
	    display_name: string;
	    registry_name: string;
	    summary: string;
	    description: string;
	    slug: string;
	    version: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.registry_name = source["registry_name"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.slug = source["slug"];
	        this.version = source["version"];
	        this.score = source["score"];
	    }
	}
	export class SkillConfig {
	    id: number;
	    name: string;
	    description: string;
	    tags: string;
	    content: string;
	    color: string;
	    icon: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SkillConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.content = source["content"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.active = source["active"];
	    }
	}
	export class SkillFullInfo {
	    name: string;
	    description: string;
	    version: string;
	    registry: string;
	    url: string;
	    markdown: string;
	    raw: string;
	    line_count: number;
	    char_count: number;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillFullInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.registry = source["registry"];
	        this.url = source["url"];
	        this.markdown = source["markdown"];
	        this.raw = source["raw"];
	        this.line_count = source["line_count"];
	        this.char_count = source["char_count"];
	        this.tags = source["tags"];
	    }
	}

}

export namespace specwizard {
	
	export class Business {
	    state_management: string;
	    api_contract: string;
	    customization_details: string;
	    final_adjustments: string;
	    architecture_recommendations: string;
	
	    static createFrom(source: any = {}) {
	        return new Business(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state_management = source["state_management"];
	        this.api_contract = source["api_contract"];
	        this.customization_details = source["customization_details"];
	        this.final_adjustments = source["final_adjustments"];
	        this.architecture_recommendations = source["architecture_recommendations"];
	    }
	}
	export class Dependency {
	    lib: string;
	    ver: string;
	    mandatory: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Dependency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lib = source["lib"];
	        this.ver = source["ver"];
	        this.mandatory = source["mandatory"];
	    }
	}
	export class StackItem {
	    name: string;
	    example: string;
	
	    static createFrom(source: any = {}) {
	        return new StackItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.example = source["example"];
	    }
	}
	export class SpecWizardConfig {
	    id: string;
	    name: string;
	    description: string;
	    expert_language_plugin: string;
	    prd: string;
	    functional_requirements: string[];
	    non_functional_requirements: string[];
	    persistence: string;
	    architecture: string;
	    engineering_philosophies: string[];
	    design_patterns: string[];
	    data_patterns: string[];
	    stack_config: StackItem[];
	    business: Business;
	    color: string;
	    icon: string;
	    architecture_health: number;
	    dependency_manifest: Dependency[];
	    stack_plugin: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new SpecWizardConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.expert_language_plugin = source["expert_language_plugin"];
	        this.prd = source["prd"];
	        this.functional_requirements = source["functional_requirements"];
	        this.non_functional_requirements = source["non_functional_requirements"];
	        this.persistence = source["persistence"];
	        this.architecture = source["architecture"];
	        this.engineering_philosophies = source["engineering_philosophies"];
	        this.design_patterns = source["design_patterns"];
	        this.data_patterns = source["data_patterns"];
	        this.stack_config = this.convertValues(source["stack_config"], StackItem);
	        this.business = this.convertValues(source["business"], Business);
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.architecture_health = source["architecture_health"];
	        this.dependency_manifest = this.convertValues(source["dependency_manifest"], Dependency);
	        this.stack_plugin = source["stack_plugin"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace specwizardmgr {
	
	export class Option {
	    id: string;
	    name: string;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new Option(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class Recommendation {
	    level: string;
	    title: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Recommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.title = source["title"];
	        this.description = source["description"];
	    }
	}

}

export namespace tool {
	
	export class ToolProfile {
	    id: number;
	    name: string;
	    color: string;
	    icon: string;
	    tools: string[];
	
	    static createFrom(source: any = {}) {
	        return new ToolProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.tools = source["tools"];
	    }
	}
	export class ToolUIInfo {
	    name: string;
	    description: string;
	    category: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolUIInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace worker {
	
	export class WorkerConfig {
	    id: number;
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
	
	    static createFrom(source: any = {}) {
	        return new WorkerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.persona = source["persona"];
	        this.language = source["language"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.connection_type = source["connection_type"];
	        this.connection_name = source["connection_name"];
	        this.connection_config = source["connection_config"];
	        this.inherit_folders = source["inherit_folders"];
	        this.inherit_knowledge = source["inherit_knowledge"];
	        this.inherit_skills = source["inherit_skills"];
	        this.inherit_tools = source["inherit_tools"];
	        this.inherit_persona = source["inherit_persona"];
	    }
	}

}

export namespace workspace {
	
	export class WorkspaceConfig {
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
	    spec_wizard: string;
	    spec_wizard_id: string;
	    agents: string[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.path = source["path"];
	        this.folders = source["folders"];
	        this.personality = source["personality"];
	        this.routing_rules = source["routing_rules"];
	        this.knowledge = source["knowledge"];
	        this.worker_names = source["worker_names"];
	        this.skills = source["skills"];
	        this.tools = source["tools"];
	        this.enabled = source["enabled"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.max_prompt_send = source["max_prompt_send"];
	        this.commit_changes = source["commit_changes"];
	        this.max_context_length = source["max_context_length"];
	        this.spec_wizard = source["spec_wizard"];
	        this.spec_wizard_id = source["spec_wizard_id"];
	        this.agents = source["agents"];
	    }
	}
	export class WorkspaceTemplate {
	    id: number;
	    name: string;
	    description: string;
	    personality: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.personality = source["personality"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


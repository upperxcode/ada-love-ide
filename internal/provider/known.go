// Package provider holds the catalog of known LLM providers and connection
// types. (Model fetching and capability inference will be added in later
// commits of the same feature plan.)
package provider

// KnownProvider is a preset entry users can pick from when creating a provider.
type KnownProvider struct {
	Name  string
	URL   string
	Type  string // connection type, matches a KnownConnectionTypes Value
	Color string
	Icon  string
}

// KnownProviders is the catalog ported from the React reference
// (ada-love-ai/frontend/src/components/settings/ModelsSection.tsx).
var KnownProviders = []KnownProvider{
	{"OpenAI", "https://api.openai.com/v1", "openai", "#10a37f", "🟢"},
	{"Anthropic", "https://api.anthropic.com/v1", "anthropic", "#d97757", "🟠"},
	{"Google Gemini", "https://generativelanguage.googleapis.com/v1beta", "gemini", "#4285f4", "🔵"},
	{"OpenRouter", "https://openrouter.ai/api/v1", "openrouter", "#8b5cf6", "🟣"},
	{"Groq", "https://api.groq.com/openai/v1", "groq", "#f55036", "🔴"},
	{"DeepSeek", "https://api.deepseek.com/v1", "deepseek", "#4d6bfe", "🔷"},
	{"Together AI", "https://api.together.xyz/v1", "together", "#0f6fff", "🟦"},
	{"Ollama", "http://localhost:11434/v1", "ollama", "#dba059", "🦙"},
	{"LM Studio", "http://localhost:1234/v1", "lmstudio", "#5b9bd5", "🖥️"},
	{"Cloudflare", "https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1", "cloudflare", "#f38020", "🟧"},
}

// ConnectionType is a selectable API dialect.
type ConnectionType struct {
	Value string
	Label string
}

// KnownConnectionTypes lists every connection type the app understands.
var KnownConnectionTypes = []ConnectionType{
	{"openai", "OpenAI Compatible"},
	{"anthropic", "Anthropic"},
	{"gemini", "Google Gemini"},
	{"ollama", "Ollama"},
	{"lmstudio", "LM Studio"},
	{"claude", "Claude"},
	{"deepseek", "DeepSeek"},
	{"groq", "Groq"},
	{"together", "Together AI"},
	{"custom", "Custom"},
}

// KnownProviderByName returns the preset for the given display name, or nil.
func KnownProviderByName(name string) *KnownProvider {
	for i := range KnownProviders {
		if KnownProviders[i].Name == name {
			return &KnownProviders[i]
		}
	}
	return nil
}

// keylessTypes are connection types whose /models endpoint needs no API key.
var keylessTypes = map[string]bool{
	"ollama":   true,
	"lmstudio": true,
}

// NeedsAPIKey reports whether a connection type requires an API key to fetch
// its model list.
func NeedsAPIKey(typeConnection string) bool {
	return !keylessTypes[typeConnection]
}

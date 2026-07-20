package provider

import "strings"

// Capabilities are the per-model flags inferred from its identifier.
type Capabilities struct {
	Free      bool
	Thinking  bool
	Tools     bool
	Vision    bool
	Embedding bool
}

// InferCapabilities derives capability flags from a model identifier using
// conservative substring rules. False-negatives are acceptable (the user can
// tick a flag on the card); false-positives are not (they pollute filters).
func InferCapabilities(modelID string) Capabilities {
	l := strings.ToLower(modelID)

	embedding := containsAny(l, "embed", "e5", "bge")
	vision := containsAny(l, "vision", "-vl", "vl-", "4o", "image")
	thinking := containsAny(l, "think", "reason", "o1", "o3", "deepseek-r")
	free := strings.Contains(l, "free")
	// Tools are the norm for LLMs; only disable for embedding models.
	tools := !embedding

	return Capabilities{
		Free:      free,
		Thinking:  thinking,
		Tools:     tools,
		Vision:    vision,
		Embedding: embedding,
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

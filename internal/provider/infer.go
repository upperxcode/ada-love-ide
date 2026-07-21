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
	tools := !embedding

	return Capabilities{
		Free:      free,
		Thinking:  thinking,
		Tools:     tools,
		Vision:    vision,
		Embedding: embedding,
	}
}

// InferContextSize retorna o tamanho máximo de contexto (em tokens) conhecido
// para um modelo, ou 0 se não for possível inferir. A lookup cobre os modelos
// mais comuns; modelos desconhecidos retornam 0 e o usuário pode configurar
// manualmente no provider.
func InferContextSize(modelID string) int {
	l := strings.ToLower(modelID)

	// ── Claude ──────────────────────────────────────────────────
	if strings.Contains(l, "claude") {
		if strings.Contains(l, "opus") {
			return 200000
		}
		if strings.Contains(l, "sonnet") {
			return 200000
		}
		if strings.Contains(l, "haiku") {
			return 200000
		}
		return 100000
	}

	// ── GPT-4 ───────────────────────────────────────────────────
	if strings.Contains(l, "gpt-4") || strings.Contains(l, "gpt4") {
		if strings.Contains(l, "turbo") || strings.Contains(l, "128k") {
			return 128000
		}
		if strings.Contains(l, "32k") {
			return 32000
		}
		return 8192
	}

	// ── GPT-4o ──────────────────────────────────────────────────
	if strings.Contains(l, "gpt-4o") || strings.Contains(l, "gpt4o") {
		if strings.Contains(l, "mini") {
			return 128000
		}
		return 128000
	}

	// ── o1 / o3 ─────────────────────────────────────────────────
	if strings.Contains(l, "o1") || strings.Contains(l, "o3") {
		return 200000
	}

	// ── GPT-3.5 ─────────────────────────────────────────────────
	if strings.Contains(l, "gpt-3.5") || strings.Contains(l, "gpt35") {
		if strings.Contains(l, "16k") {
			return 16000
		}
		return 4096
	}

	// ── DeepSeek ────────────────────────────────────────────────
	if strings.Contains(l, "deepseek") {
		if strings.Contains(l, "v3") || strings.Contains(l, "r1") {
			return 65536
		}
		return 8192
	}

	// ── Gemini ──────────────────────────────────────────────────
	if strings.Contains(l, "gemini") {
		if strings.Contains(l, "pro") || strings.Contains(l, "ultra") || strings.Contains(l, "flash") {
			return 128000
		}
		return 32000
	}

	// ── Llama ───────────────────────────────────────────────────
	if strings.Contains(l, "llama") || strings.Contains(l, "llm") {
		if strings.Contains(l, "70b") || strings.Contains(l, "405b") {
			return 32768
		}
		if strings.Contains(l, "8b") || strings.Contains(l, "13b") || strings.Contains(l, "3.1") || strings.Contains(l, "3.2") || strings.Contains(l, "3.3") {
			return 131072
		}
		return 8192
	}

	// ── Mistral ─────────────────────────────────────────────────
	if strings.Contains(l, "mistral") || strings.Contains(l, "mixtral") || strings.Contains(l, "codestral") {
		if strings.Contains(l, "large") || strings.Contains(l, "mixtral") || strings.Contains(l, "8x22") {
			return 65536
		}
		if strings.Contains(l, "nemo") || strings.Contains(l, "codestral") {
			return 128000
		}
		if strings.Contains(l, "small") || strings.Contains(l, "8b") {
			return 32768
		}
		return 32768
	}

	// ── Qwen ────────────────────────────────────────────────────
	if strings.Contains(l, "qwen") {
		if strings.Contains(l, "2.5") || strings.Contains(l, "2-") {
			return 131072
		}
		return 32768
	}

	// ── Command R ───────────────────────────────────────────────
	if strings.Contains(l, "command") || strings.Contains(l, "c4ai") {
		return 128000
	}

	// ── DBRX ────────────────────────────────────────────────────
	if strings.Contains(l, "dbrx") {
		return 32768
	}

	// ── Phi ─────────────────────────────────────────────────────
	if strings.Contains(l, "phi-3") || strings.Contains(l, "phi3") || strings.Contains(l, "phi-4") || strings.Contains(l, "phi4") {
		return 128000
	}

	// ── Granite ─────────────────────────────────────────────────
	if strings.Contains(l, "granite") {
		return 131072
	}

	return 0
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

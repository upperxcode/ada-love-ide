package provider

import "testing"

func TestInferCapabilities(t *testing.T) {
	cases := []struct {
		id   string
		want Capabilities
	}{
		{"gpt-4o", Capabilities{Vision: true, Tools: true}},
		{"gpt-4o-mini", Capabilities{Vision: true, Tools: true}},
		{"o1-preview", Capabilities{Thinking: true, Tools: true}},
		{"o3-mini", Capabilities{Thinking: true, Tools: true}},
		{"deepseek-r1", Capabilities{Thinking: true, Tools: true}},
		{"deepseek-chat", Capabilities{Tools: true}},
		{"claude-3-5-sonnet", Capabilities{Tools: true}},
		{"text-embedding-3-small", Capabilities{Embedding: true}},
		{"bge-large-en", Capabilities{Embedding: true}},
		{"qwen2-vl-7b", Capabilities{Vision: true, Tools: true}},
		{"llama-3.1-8b", Capabilities{Tools: true}},
		{"groq-gemma-free", Capabilities{Free: true, Tools: true}},
		{"gemini-1.5-pro", Capabilities{Tools: true}},
	}
	for _, c := range cases {
		got := InferCapabilities(c.id)
		if got != c.want {
			t.Errorf("InferCapabilities(%q) = %+v, want %+v", c.id, got, c.want)
		}
	}
}

package provider

import "testing"

func TestKnownProviderByName(t *testing.T) {
	cases := []struct {
		name string
		want string // expected URL; "" means not found
	}{
		{"OpenAI", "https://api.openai.com/v1"},
		{"Anthropic", "https://api.anthropic.com/v1"},
		{"Ollama", "http://localhost:11434/v1"},
		{"Unknown Provider", ""},
	}
	for _, c := range cases {
		got := KnownProviderByName(c.name)
		if c.want == "" {
			if got != nil {
				t.Errorf("KnownProviderByName(%q) = %+v, want nil", c.name, got)
			}
			continue
		}
		if got == nil || got.URL != c.want {
			t.Errorf("KnownProviderByName(%q).URL = %v, want %q", c.name, got, c.want)
		}
	}
}

func TestNeedsAPIKey(t *testing.T) {
	if NeedsAPIKey("ollama") {
		t.Error("ollama should not need an api key")
	}
	if NeedsAPIKey("lmstudio") {
		t.Error("lmstudio should not need an api key")
	}
	if !NeedsAPIKey("openai") {
		t.Error("openai should need an api key")
	}
	if !NeedsAPIKey("custom") {
		t.Error("custom should need an api key")
	}
}

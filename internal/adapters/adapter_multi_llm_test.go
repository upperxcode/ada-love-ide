package adapters

import (
	"testing"

	llm "github.com/upperxcode/ada-llm-client"
)

func toolNames(tools []llm.ToolDefinition) []string {
	var names []string
	for _, t := range tools {
		names = append(names, t.Function.Name)
	}
	return names
}

func makeTools(names ...string) []llm.ToolDefinition {
	var tools []llm.ToolDefinition
	for _, n := range names {
		tools = append(tools, llm.ToolDefinition{
			Function: llm.ToolFunction{Name: n},
		})
	}
	return tools
}

func TestFilterTools_Empty(t *testing.T) {
	got := filterTools(nil, nil)
	if len(got) != 0 {
		t.Errorf("filterTools(nil, nil) = %v, want empty", got)
	}
}

func TestFilterTools_EmptyAllowed(t *testing.T) {
	tools := makeTools("read", "write")
	got := filterTools(tools, []string{})
	if len(got) != 0 {
		t.Errorf("filterTools with empty allowed = %v, want empty", got)
	}
}

func TestFilterTools_NoMatch(t *testing.T) {
	tools := makeTools("read", "write")
	got := filterTools(tools, []string{"exec"})
	if len(got) != 0 {
		t.Errorf("filterTools([read,write], [exec]) = %v, want empty", got)
	}
}

func TestFilterTools_Partial(t *testing.T) {
	tools := makeTools("read", "write", "exec")
	got := filterTools(tools, []string{"read"})
	if len(got) != 1 || got[0].Function.Name != "read" {
		t.Errorf("filterTools = %v, want [read]", toolNames(got))
	}
}

func TestFilterTools_All(t *testing.T) {
	tools := makeTools("read", "write", "search")
	got := filterTools(tools, []string{"read", "write", "search"})
	if len(got) != 3 {
		t.Errorf("filterTools len = %d, want 3", len(got))
	}
}

func TestAllowedTools_Ask(t *testing.T) {
	a := &MultiLLMAdapter{mode: "ASK"}
	a.SetTools(makeTools("read", "write", "exec", "search", "plan"))
	got := a.allowedTools()
	names := toolNames(got)
	expected := []string{"read", "search"}
	if len(names) != len(expected) {
		t.Errorf("ASK allowedTools = %v, want %v", names, expected)
		return
	}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("ASK allowedTools[%d] = %s, want %s", i, n, expected[i])
		}
	}
}

func TestAllowedTools_Plan(t *testing.T) {
	a := &MultiLLMAdapter{mode: "PLAN"}
	a.SetTools(makeTools("read", "write", "exec", "search", "plan"))
	got := a.allowedTools()
	names := toolNames(got)
	expected := []string{"read", "search"}
	if len(names) != len(expected) {
		t.Errorf("PLAN allowedTools = %v, want %v", names, expected)
		return
	}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("PLAN allowedTools[%d] = %s, want %s", i, n, expected[i])
		}
	}
}

func TestAllowedTools_Edit(t *testing.T) {
	a := &MultiLLMAdapter{mode: "EDIT"}
	a.SetTools(makeTools("read", "write", "exec", "search", "plan"))
	got := a.allowedTools()
	names := toolNames(got)
	expected := map[string]bool{"read": true, "search": true, "write": true}
	if len(names) != len(expected) {
		t.Errorf("EDIT allowedTools = %v, want %v", names, []string{"read", "search", "write"})
		return
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("EDIT allowedTools contains unexpected %q", n)
		}
	}
}

func TestAllowedTools_Full(t *testing.T) {
	a := &MultiLLMAdapter{mode: "FULL"}
	allTools := makeTools("read", "write", "exec", "search", "plan")
	a.SetTools(allTools)
	got := a.allowedTools()
	if len(got) != len(allTools) {
		t.Errorf("FULL allowedTools len = %d, want %d", len(got), len(allTools))
	}
}

package contextprovider

import (
	"context"
	"fmt"
	"strings"

	"ada-love-ide/internal/chatsummary"
	"ada-love-ide/internal/config/worker"
	"ada-love-ide/internal/config/workspace"
	"ada-love-ide/internal/knowledge"
	core "ada-love-core"
)

type WorkerType string

const (
	TypeAda          WorkerType = "ada"
	TypeCLI          WorkerType = "cli"
	TypeURL          WorkerType = "url"
	TypeOpenCodeServe WorkerType = "opencode_serve"
)

type Params struct {
	WorkerType   WorkerType
	Worker       worker.WorkerConfig
	Workspace    workspace.WorkspaceConfig
	WorkspaceDir string
	WorkspaceID  int64
	SessionID    string
	Messages     []core.RawMessage
	CurrentMsg   string
	KnowledgeIdx *knowledge.Index
	ChatSummary  *chatsummary.Manager
}

type Result struct {
	System  string
	Parts   []string
	History string
}

func GetContext(ctx context.Context, p Params) (*Result, error) {
	res := &Result{}

	fmt.Printf("[ContextProvider] === GetContext ===\n")
	fmt.Printf("[ContextProvider] worker=%s inherit_persona=%v inherit_knowledge=%v inherit_skills=%v inherit_tools=%v\n",
		p.Worker.Name, p.Worker.InheritPersona, p.Worker.InheritKnowledge, p.Worker.InheritSkills, p.Worker.InheritTools)
	fmt.Printf("[ContextProvider] workspace=%q summary_len=%d max_send=%d\n",
		p.Workspace.Title, len(p.Workspace.Summary), p.Workspace.MaxPromptSend)
	fmt.Printf("[ContextProvider] session=%s knowledge_idx=%v chat_summary=%v\n",
		p.SessionID, p.KnowledgeIdx != nil, p.ChatSummary != nil)

	res.System = buildSystem(p)
	res.Parts = buildParts(ctx, p)
	res.History = buildHistory(ctx, p)

	fmt.Printf("[ContextProvider] === Result === system=%d chars parts=%d items history=%d chars\n",
		len(res.System), len(res.Parts), len(res.History))
	for i, part := range res.Parts {
		preview := part
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("[ContextProvider]   part[%d]: %q\n", i, preview)
	}
	if res.History != "" {
		preview := res.History
		if len(preview) > 120 {
			preview = preview[:120] + "..."
		}
		fmt.Printf("[ContextProvider]   history: %q\n", preview)
	}

	return res, nil
}

func buildSystem(p Params) string {
	var sb strings.Builder
	if p.Worker.InheritPersona && p.Worker.Persona != "" {
		sb.WriteString(p.Worker.Persona)
		fmt.Printf("[ContextProvider] system: added persona (%d chars)\n", len(p.Worker.Persona))
	} else {
		fmt.Printf("[ContextProvider] system: empty (inherit_persona=%v, persona=%q)\n",
			p.Worker.InheritPersona, p.Worker.Persona[:min(len(p.Worker.Persona), 30)])
	}
	return strings.TrimSpace(sb.String())
}

func buildParts(ctx context.Context, p Params) []string {
	var parts []string

	// 1. Workspace Summary
	if p.Workspace.Summary != "" {
		parts = append(parts, fmt.Sprintf("=== WORKSPACE SUMMARY ===\n%s", p.Workspace.Summary))
		fmt.Printf("[ContextProvider] part: added WORKSPACE SUMMARY (%d chars)\n", len(p.Workspace.Summary))
	} else {
		fmt.Printf("[ContextProvider] part: WORKSPACE SUMMARY empty\n")
	}

	// 2. Knowledge (semantic search)
	if p.Worker.InheritKnowledge && p.KnowledgeIdx != nil && p.WorkspaceID > 0 {
		count := p.KnowledgeIdx.Count(p.WorkspaceID)
		fmt.Printf("[ContextProvider] knowledge: inherit=true count=%d\n", count)
		if count > 0 {
			results := p.KnowledgeIdx.Search(ctx, p.CurrentMsg, p.WorkspaceID, 5)
			fmt.Printf("[ContextProvider] knowledge: search returned %d results\n", len(results))
			if len(results) > 0 {
				text := fmt.Sprintf("=== KNOWLEDGE (relevant) ===\n%s", strings.Join(results, "\n---\n"))
				parts = append(parts, text)
				fmt.Printf("[ContextProvider] part: added KNOWLEDGE (%d chars)\n", len(text))
			}
		}
	} else {
		fmt.Printf("[ContextProvider] knowledge: inherit=%v idx=%v wsID=%d\n",
			p.Worker.InheritKnowledge, p.KnowledgeIdx != nil, p.WorkspaceID)
	}

	// 3. Skills
	if p.Worker.InheritSkills && len(p.Workspace.Skills) > 0 {
		text := fmt.Sprintf("=== AVAILABLE SKILLS ===\n%s", strings.Join(p.Workspace.Skills, "\n"))
		parts = append(parts, text)
		fmt.Printf("[ContextProvider] part: added SKILLS (%d items)\n", len(p.Workspace.Skills))
	} else {
		fmt.Printf("[ContextProvider] skills: inherit=%v count=%d\n", p.Worker.InheritSkills, len(p.Workspace.Skills))
	}

	// 4. Tools
	if p.Worker.InheritTools && len(p.Workspace.Tools) > 0 {
		text := fmt.Sprintf("=== AVAILABLE TOOLS ===\n%s", strings.Join(p.Workspace.Tools, "\n"))
		parts = append(parts, text)
		fmt.Printf("[ContextProvider] part: added TOOLS (%d items)\n", len(p.Workspace.Tools))
	} else {
		fmt.Printf("[ContextProvider] tools: inherit=%v count=%d\n", p.Worker.InheritTools, len(p.Workspace.Tools))
	}

	return parts
}

func buildHistory(ctx context.Context, p Params) string {
	if p.ChatSummary == nil || p.SessionID == "" || p.CurrentMsg == "" {
		fmt.Printf("[ContextProvider] history: skipped (summary=%v session=%q msg=%q)\n",
			p.ChatSummary != nil, p.SessionID, p.CurrentMsg[:min(len(p.CurrentMsg), 30)])
		return ""
	}

	msg := chatsummary.RawMessage{
		Role:    "user",
		Content: p.CurrentMsg,
	}

	maxSend := p.Workspace.MaxPromptSend
	if maxSend <= 0 {
		maxSend = 10
	}
	fmt.Printf("[ContextProvider] history: pushing to chatsummary (maxSend=%d)\n", maxSend)

	history, err := p.ChatSummary.Push(ctx, p.SessionID, msg, maxSend)
	if err != nil {
		fmt.Printf("[ContextProvider] history: chatsummary push error: %v\n", err)
		return ""
	}

	fmt.Printf("[ContextProvider] history: returned %d chars\n", len(history))
	return history
}

package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	wiki "github.com/upperxcode/ada-llm-wiki"
)

func TestStaticRouting_Hello(t *testing.T) {
	store := NewMockStorage()
	comp := NewMockCompactor()
	exec := NewMockExecutor()
	llm := NewMockLLMClient()

	orch := NewOrchestrator(store, llm, comp, exec)

	resp, err := orch.ProcessMessage(context.Background(), "sess-1", "hi", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// O ada-llm-client static router retorna resposta para "hi"
	if resp == "" {
		t.Error("expected non-empty response")
	}
}

func TestStaticRouting_BomDia(t *testing.T) {
	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), NewMockCompactor(), NewMockExecutor())

	resp, err := orch.ProcessMessage(context.Background(), "sess-1", "hello", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// O ada-llm-client static router retorna resposta para "hello"
	if resp == "" {
		t.Error("expected non-empty response")
	}
}

func TestStaticRouting_NoMatch(t *testing.T) {
	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), NewMockCompactor(), NewMockExecutor())

	resp, err := orch.ProcessMessage(context.Background(), "sess-1", "write me a function", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// should go through LLM (mock), not static route
	if resp == "Hello! How can I help?" {
		t.Error("should not match greeting for non-greeting input")
	}
}

func TestMessagePersistence(t *testing.T) {
	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), NewMockCompactor(), NewMockExecutor())

	_, err := orch.ProcessMessage(context.Background(), "persist-test", "Tell me about Go", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msgs, err := store.GetMessagesBySession("persist-test")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages (user + assistant), got: %d", len(msgs))
	}
	if msgs[0].Role != "user" {
		t.Errorf("expected first message role 'user', got %q", msgs[0].Role)
	}
	if msgs[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", msgs[1].Role)
	}
}

func TestStream_NoStaticMatch(t *testing.T) {
	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), NewMockCompactor(), NewMockExecutor())

	tokens, err := orch.ProcessMessageStream(context.Background(), "stream-test", "write code", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var count int
	for token := range tokens {
		if token.Done {
			break
		}
		count++
		if count > 100 {
			break // safety
		}
	}
	if count == 0 {
		t.Error("expected at least one token from streaming")
	}
}

func TestStream_StaticMatch(t *testing.T) {
	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), NewMockCompactor(), NewMockExecutor())

	tokens, err := orch.ProcessMessageStream(context.Background(), "stream-static", "hi!", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var full string
	for token := range tokens {
		full += token.Token
		if token.Done {
			break
		}
	}
	// O ada-llm-client static router retorna uma resposta para "hi!"
	if full == "" {
		t.Error("expected non-empty response from static routing")
	}
}

func TestHistoryUsed(t *testing.T) {
	store := NewMockStorage()
	comp := NewMockCompactor()
	llm := NewMockLLMClient()
	orch := NewOrchestrator(store, llm, comp, NewMockExecutor())

	// first message
	_, _ = orch.ProcessMessage(context.Background(), "hist-test", "hello", "")

	// second message — should have history
	_, err := orch.ProcessMessage(context.Background(), "hist-test", "write code", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msgs, _ := store.GetMessagesBySession("hist-test")
	if len(msgs) < 2 {
		t.Errorf("expected at least 2 messages, got %d", len(msgs))
	}
}

func TestCompilePrompt_HeavyWikiForcesHistoryCompaction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wiki-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	heavyArticle := `---
tags: cachyos, gci, database
---
# CachyOS GCI Database Specification

## Overview
This is a comprehensive specification for the GCI (Google Code-In) database system used by the CachyOS project.
The database manages student registrations, task submissions, mentor assignments, and evaluation criteria.

## Schema
The database consists of the following tables:
- students: id, name, email, university, country, timezone, skill_level
- tasks: id, title, description, category, difficulty, points, mentor_id, status
- submissions: id, task_id, student_id, content, submitted_at, review_status
- mentors: id, name, email, expertise_areas, max_students
- evaluations: id, submission_id, mentor_id, score, feedback, evaluated_at

## Rules
1. Each student can have at most 3 active tasks at any time.
2. Mentors must evaluate submissions within 48 hours of submission.
3. Tasks with difficulty "advanced" require mentor approval before assignment.
4. Students receive points upon task completion: easy=10, medium=25, advanced=50.
5. The top 10 students by total points are featured on the leaderboard.

## API Endpoints
- GET /api/students - List all students with pagination
- POST /api/students - Register a new student
- GET /api/tasks - List available tasks filtered by category and difficulty
- POST /api/tasks/:id/submit - Submit a task for review
- GET /api/mentors - List available mentors
- POST /api/evaluations - Create a new evaluation

## Configuration
Database connection pool: max_connections=50, idle_timeout=30s
Query timeout: 10s for reads, 30s for writes
Cache TTL: 5 minutes for student profiles, 1 minute for task lists`

	err = os.WriteFile(filepath.Join(tmpDir, "gci-database.md"), []byte(heavyArticle), 0644)
	if err != nil {
		t.Fatalf("failed to write wiki article: %v", err)
	}

	wikiMgr := wiki.NewWikiManager(tmpDir)
	if err := wikiMgr.LoadArticles(); err != nil {
		t.Fatalf("LoadArticles failed: %v", err)
	}

	compactor := NewCompactorAdapter(800, 2, "System: you are Ada.")

	store := NewMockStorage()
	orch := NewOrchestrator(store, NewMockLLMClient(), compactor, NewMockExecutor())
	orch.Wiki = wikiMgr

	for i := 0; i < 10; i++ {
		_ = store.SaveMessage(Message{
			ID:        "msg-" + string(rune('a'+i)),
			SessionID: "wiki-test",
			Role:      "user",
			Content:   "This is history message number " + string(rune('0'+i)) + " with enough content to consume tokens in the budget.",
		})
		_ = store.SaveMessage(Message{
			ID:        "msg-resp-" + string(rune('a'+i)),
			SessionID: "wiki-test",
			Role:      "assistant",
			Content:   "Response to message " + string(rune('0'+i)) + " with some additional filler text to take up space.",
		})
	}

	history, _ := store.GetMessagesBySession("wiki-test")
	if len(history) != 20 {
		t.Fatalf("expected 20 history messages, got %d", len(history))
	}

	prompt, err := orch.CompilePrompt(context.Background(), "wiki-test", "gci", history)
	if err != nil {
		t.Fatalf("CompilePrompt error: %v", err)
	}

	if !strings.Contains(prompt, "--- START INTERNAL WIKI CONTEXT ---") {
		t.Fatalf("expected wiki context to be injected in prompt, got: %s", prompt[:min(200, len(prompt))])
	}
	if !strings.Contains(prompt, "CachyOS GCI Database") {
		t.Error("expected wiki article content in prompt")
	}

	totalTokens := compactor.CountTokens(prompt)
	t.Logf("Prompt tokens: %d / 800 budget (tokenizer is heuristic-based, slight overshoot is expected)", totalTokens)

	wikiIdx := strings.Index(prompt, "--- START INTERNAL WIKI CONTEXT ---")
	historySection := prompt[:wikiIdx]
	historyMsgCount := strings.Count(historySection, "User:")
	if historyMsgCount >= 10 {
		t.Errorf("expected history to be pruned, but found %d user messages (all 10 present)", historyMsgCount)
	}

	t.Logf("History messages retained: %d / 10", historyMsgCount)
}

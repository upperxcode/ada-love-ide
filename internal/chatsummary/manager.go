package chatsummary

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	llm "github.com/upperxcode/ada-llm-client"
)

// ── Exported types ──────────────────────────────────────────────

// RawMessage is a simplified chat message used by the summarization
// package. It carries only the fields needed for summarization and
// for building the formatted context string.
type RawMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMClient is the interface the summarizer uses to talk to an LLM.
// Implementations must call the underlying model and return the
// generated text. Tool calls are ignored — only the text response
// is used.
type LLMClient interface {
	Chat(ctx context.Context, messages []llm.Message) (string, error)
}

// ── Manager ─────────────────────────────────────────────────────

// Manager handles incremental, asynchronous summarization of
// conversation histories per session.
//
// Each session has a directory under baseDir with:
//   - messages.jsonl     – append-only log of all messages
//   - last_summary.txt   – most recent summary (overwritten)
//
// The Manager is safe for concurrent use.
type Manager struct {
	baseDir   string
	llmClient LLMClient
	wg        sync.WaitGroup
}

// NewManager creates a new Manager.
// baseDir is the root directory under which per-session
// directories are created (e.g. ~/.config/ada-love-ide/summaries/).
func NewManager(baseDir string, llmClient LLMClient) *Manager {
	return &Manager{
		baseDir:   baseDir,
		llmClient: llmClient,
	}
}

// Push appends a message to the session's history, launches an
// asynchronous summarization if the history exceeds maxSend, and
// returns the current formatted context synchronously.
//
// The returned context has the shape:
//
//	[prior summary if any]
//
//	role: content
//	role: content
//	...
//
// The goroutine runs with context.Background() and a 30 s LLM timeout.
// Errors from the goroutine are only logged — never returned to the caller.
func (m *Manager) Push(ctx context.Context, sessionID string, msg RawMessage, maxSend int) (string, error) {
	// 1. Persist the new message.
	if err := appendMessage(m.baseDir, sessionID, msg); err != nil {
		return "", fmt.Errorf("persist message: %w", err)
	}

	// 2. Build synchronous context (never waits for the goroutine).
	contextStr, err := m.buildCurrentContext(ctx, sessionID, maxSend)
	if err != nil {
		return "", err
	}

	// 3. Launch async summarization in the background.
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runSummarize(sessionID, maxSend)
	}()

	return contextStr, nil
}

// Get returns the current formatted context for a session without
// triggering any summarization.
func (m *Manager) Get(ctx context.Context, sessionID string, maxSend int) (string, error) {
	return m.buildCurrentContext(ctx, sessionID, maxSend)
}

// Close waits for all pending summarization goroutines to finish,
// with a hard timeout of 5 seconds. Any goroutines still running
// after the timeout are abandoned.
func (m *Manager) Close() error {
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for summarization goroutines after 5s")
	}
}

// ── internal helpers ────────────────────────────────────────────

// buildCurrentContext assembles the formatted context string:
//   - prior summary (if any) + blank line
//   - last maxSend messages
func (m *Manager) buildCurrentContext(ctx context.Context, sessionID string, maxSend int) (string, error) {
	summary, err := readSummary(m.baseDir, sessionID)
	if err != nil {
		return "", fmt.Errorf("read summary: %w", err)
	}

	recent, err := readRecentMessages(m.baseDir, sessionID, maxSend)
	if err != nil {
		return "", fmt.Errorf("read recent messages: %w", err)
	}

	return buildContext(summary, recent), nil
}

// runSummarize is the goroutine entry point. It reads all messages,
// decides whether summarization is needed, calls the LLM, and
// persists the result.
func (m *Manager) runSummarize(sessionID string, maxSend int) {
	all, err := readAllMessages(m.baseDir, sessionID)
	if err != nil {
		log.Printf("[chatsummary] error reading messages for session %s: %v", sessionID, err)
		return
	}

	// Skip if history still fits within the limit.
	if len(all) <= maxSend {
		return
	}

	summary, err := generateSummary(context.Background(), m.llmClient, all)
	if err != nil {
		log.Printf("[chatsummary] summary generation failed for session %s: %v", sessionID, err)
		return
	}

	if err := writeSummary(m.baseDir, sessionID, summary); err != nil {
		log.Printf("[chatsummary] error writing summary for session %s: %v", sessionID, err)
	}
}

// buildContext formats the final string returned by Push / Get.
func buildContext(summary string, recent []RawMessage) string {
	var sb strings.Builder

	if summary != "" {
		sb.WriteString(summary)
		sb.WriteString("\n\n")
	}

	for i, msg := range recent {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
	}

	return sb.String()
}

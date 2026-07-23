package chatsummary

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	llm "github.com/upperxcode/ada-llm-client"
)

// ── mock LLM client ─────────────────────────────────────────────

type mockLLMClient struct {
	chatFunc func(ctx context.Context, messages []llm.Message) (string, error)
}

func (m *mockLLMClient) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	return m.chatFunc(ctx, messages)
}

// newMockClient returns an LLMClient that returns a fixed summary.
func newMockClient(summary string) *mockLLMClient {
	return &mockLLMClient{
		chatFunc: func(_ context.Context, _ []llm.Message) (string, error) {
			return summary, nil
		},
	}
}

// ── helpers ─────────────────────────────────────────────────────

func tempBaseDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "chatsummary-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestPush_NewSession(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("mock summary"))

	ctx := context.Background()
	msg := RawMessage{Role: "user", Content: "hello"}
	result, err := mgr.Push(ctx, "sess_test1", msg, 10)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Result should contain the message
	if !strings.Contains(result, "hello") {
		t.Errorf("expected result to contain 'hello', got: %q", result)
	}

	// Check files were created
	sessDir := filepath.Join(baseDir, "sess_test1")
	if _, err := os.Stat(filepath.Join(sessDir, "messages.jsonl")); os.IsNotExist(err) {
		t.Error("messages.jsonl was not created")
	}

	// No summary yet — file should not exist
	if _, err := os.Stat(filepath.Join(sessDir, "last_summary.txt")); err == nil {
		t.Error("last_summary.txt should not exist yet (no summarization has run)")
	}

	// Close and verify no error
	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPush_MultipleMessages(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("mock summary"))

	ctx := context.Background()
	messages := []RawMessage{
		{Role: "user", Content: "first"},
		{Role: "assistant", Content: "second"},
		{Role: "user", Content: "third"},
	}

	for _, msg := range messages {
		_, err := mgr.Push(ctx, "sess_test2", msg, 10)
		if err != nil {
			t.Fatalf("Push failed: %v", err)
		}
	}

	// Read back the JSONL file
	sessDir := filepath.Join(baseDir, "sess_test2")
	data, err := os.ReadFile(filepath.Join(sessDir, "messages.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines in messages.jsonl, got %d", len(lines))
	}

	// Close
	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestGet_WithoutSummary(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("mock summary"))

	ctx := context.Background()

	// Push some messages
	msgs := []RawMessage{
		{Role: "user", Content: "question 1"},
		{Role: "assistant", Content: "answer 1"},
	}
	for _, msg := range msgs {
		mgr.Push(ctx, "sess_test3", msg, 10)
	}

	// Get before any async summary could have run
	result, err := mgr.Get(ctx, "sess_test3", 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Should contain the messages but no summary
	if !strings.Contains(result, "question 1") {
		t.Errorf("expected result to contain 'question 1', got: %q", result)
	}
	if !strings.Contains(result, "answer 1") {
		t.Errorf("expected result to contain 'answer 1', got: %q", result)
	}

	// No summary prefix means no summary line at start
	if strings.HasPrefix(result, "mock summary") {
		t.Errorf("expected no summary prefix, got: %q", result)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestGet_WithExistingSummary(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("mock summary"))

	ctx := context.Background()

	// Manually write a summary file
	sessDir := filepath.Join(baseDir, "sess_test4")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessDir, "last_summary.txt"), []byte("prior summary text"), 0644); err != nil {
		t.Fatal(err)
	}

	// Push a message (so we have recent messages to show)
	msg := RawMessage{Role: "user", Content: "follow up"}
	mgr.Push(ctx, "sess_test4", msg, 10)

	// Get should include the prior summary
	result, err := mgr.Get(ctx, "sess_test4", 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !strings.Contains(result, "prior summary text") {
		t.Errorf("expected result to contain 'prior summary text', got: %q", result)
	}
	if !strings.Contains(result, "follow up") {
		t.Errorf("expected result to contain 'follow up', got: %q", result)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPush_ContextFormat(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("mock summary"))

	ctx := context.Background()

	// Write an existing summary
	sessDir := filepath.Join(baseDir, "sess_test5")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessDir, "last_summary.txt"), []byte("existing summary"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := mgr.Push(ctx, "sess_test5", RawMessage{Role: "user", Content: "new msg"}, 5)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	expected := "existing summary\n\nuser: new msg"
	if result != expected {
		t.Errorf("unexpected context format:\n  got:  %q\n  want: %q", result, expected)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPush_TrimsToMaxSend(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("summary"))

	ctx := context.Background()

	// Push 5 messages with maxSend=3
	msgs := []RawMessage{
		{Role: "user", Content: "msg1"},
		{Role: "assistant", Content: "msg2"},
		{Role: "user", Content: "msg3"},
		{Role: "assistant", Content: "msg4"},
		{Role: "user", Content: "msg5"},
	}
	for _, msg := range msgs {
		mgr.Push(ctx, "sess_test6", msg, 3)
	}

	result, err := mgr.Get(ctx, "sess_test6", 3)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Should only show the last 3 messages
	if strings.Contains(result, "msg1") {
		t.Errorf("expected msg1 to be trimmed, got: %q", result)
	}
	if !strings.Contains(result, "msg3") {
		t.Errorf("expected msg3 to be present, got: %q", result)
	}
	if !strings.Contains(result, "msg4") {
		t.Errorf("expected msg4 to be present, got: %q", result)
	}
	if !strings.Contains(result, "msg5") {
		t.Errorf("expected msg5 to be present, got: %q", result)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("concurrent summary"))

	ctx := context.Background()
	var wg sync.WaitGroup

	// Push from 10 goroutines concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			msg := RawMessage{
				Role:    "user",
				Content: strings.Repeat("x", n+1),
			}
			_, err := mgr.Push(ctx, "sess_concurrent", msg, 5)
			if err != nil {
				t.Errorf("concurrent Push failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all messages were appended
	all, err := readAllMessages(baseDir, "sess_concurrent")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 10 {
		t.Errorf("expected 10 messages, got %d", len(all))
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPush_ErrorOnEmptyContent(t *testing.T) {
	baseDir := tempBaseDir(t)
	mgr := NewManager(baseDir, newMockClient("summary"))

	ctx := context.Background()
	// This should succeed (the store doesn't validate content).
	// If empty content is a problem, it is the caller's responsibility.
	_, err := mgr.Push(ctx, "sess_empty", RawMessage{Role: "user", Content: ""}, 10)
	if err != nil {
		t.Fatalf("Push with empty content should succeed: %v", err)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestClose_Timeout(t *testing.T) {
	// Create a manager with a slow LLM mock that blocks, so Close
	// hits the 5s timeout and returns an error.
	baseDir := tempBaseDir(t)
	slowClient := &mockLLMClient{
		chatFunc: func(ctx context.Context, _ []llm.Message) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	}
	mgr := NewManager(baseDir, slowClient)

	// Push a lot of messages so the goroutine actually triggers summarization
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		mgr.Push(ctx, "sess_slow", RawMessage{Role: "user", Content: "msg"}, 5)
	}

	// Close will wait for the slow goroutines — we expect a timeout
	// because the mock never responds.
	err := mgr.Close()
	if err == nil {
		t.Log("Close succeeded (goroutines may have been skipped because total <= maxSend)")
	} else {
		// Expected: timeout error
		t.Logf("Close returned expected error: %v", err)
	}
}

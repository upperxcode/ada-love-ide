package core

import (
	"context"
	"testing"
)

func TestMockCompactor_EmptyHistory(t *testing.T) {
	c := NewMockCompactor()
	result, err := c.Compact(context.Background(), "System prompt", nil, 1000)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "System prompt" {
		t.Errorf("expected 'System prompt', got: %q", result)
	}
}

func TestMockCompactor_WithHistory(t *testing.T) {
	c := NewMockCompactor()
	history := []string{"User: hi", "Assistant: hello!"}
	result, err := c.Compact(context.Background(), "System prompt", history, 1000)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == "System prompt" {
		t.Error("expected history to be included")
	}
}

func TestMockCompactor_RespectsLimit(t *testing.T) {
	c := NewMockCompactor()
	history := []string{
		"User: message one that is very long and takes up a lot of space in the prompt buffer",
		"Assistant: response one that is also very long and continues to fill up space",
		"User: message two also very long and takes up even more space in the buffer",
	}
	result, err := c.Compact(context.Background(), "Sys", history, 50)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// should stop adding history once limit is exceeded
	if len(result) > 60 {
		t.Errorf("expected result to respect limit, got length %d", len(result))
	}
}

func TestMockExecutor(t *testing.T) {
	e := NewMockExecutor()
	result, err := e.ExecuteCommand(context.Background(), "test-session", "ls", []string{"-la"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestMockLLMClient_Streaming(t *testing.T) {
	llm := NewMockLLMClient()
	tokens, err := llm.Generate(context.Background(), "test prompt", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var count int
	var done bool
	for token := range tokens {
		if token.Done {
			done = true
			break
		}
		count++
	}
	if count == 0 {
		t.Error("expected at least one token")
	}
	if !done {
		t.Error("expected Done=true at end of stream")
	}
}

func TestMockLLMClient_Cancellation(t *testing.T) {
	llm := NewMockLLMClient()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokens, err := llm.Generate(ctx, "test", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// cancel after first token
	for range tokens {
		cancel()
		break
	}

	// channel should close without hanging
	done := make(chan struct{})
	go func() {
		for range tokens {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-make(chan struct{}, 1):
		// OK
	}
}

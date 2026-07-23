package chatsummary

import (
	"context"
	"fmt"
	"strings"
	"time"

	llm "github.com/upperxcode/ada-llm-client"
)

const summarizePrompt = `You are a conversation summarizer. Your task is to condense the conversation below into a concise summary of at most 300 tokens.

Preserve in the summary:
- Important decisions that were made
- Problem context and requirements discussed
- Agreed next steps or action items

Focus only on information that would be relevant for continuing the conversation.
Output the summary in plain text. Do not add a preamble, title, or commentary.`

// generateSummary calls the LLM to condense a list of messages into a brief summary.
//
// Messages are passed in order as chat messages. A system instruction asking for
// a concise summary (max 300 tokens) is prepended automatically.
//
// The caller's context is used as base, but a 30-second timeout is enforced
// internally so that long-running calls never block indefinitely.
func generateSummary(ctx context.Context, client LLMClient, messages []RawMessage) (string, error) {
	llmMessages := make([]llm.Message, 0, len(messages)+1)

	// System-level instruction
	llmMessages = append(llmMessages, llm.NewSystemMessage(summarizePrompt))

	// Conversation history
	for _, msg := range messages {
		llmMessages = append(llmMessages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	summary, err := client.Chat(ctx, llmMessages)
	if err != nil {
		return "", fmt.Errorf("llm summarization failed: %w", err)
	}

	return strings.TrimSpace(summary), nil
}

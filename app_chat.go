package main

import (
	"context"
	"fmt"

	"ada-love-ide/internal/core"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SendMessage dispara o turno de conversa via pacote chat.
// Em paralelo, events chat:delta/chat:turnEnd são emitidos.
func (a *App) SendMessage(sessionID, text, modelOverride, thinkingLevel, mode string) (string, error) {
	fmt.Printf("[SendMessage] ENTER session=%s text=%q model=%s thinking=%s mode=%s\n", sessionID, text[:min(50, len(text))], modelOverride, thinkingLevel, mode)
	ctx := context.Background()
	resp, err := a.eng.Chat.Send(ctx, sessionID, text, modelOverride, thinkingLevel, mode)
	fmt.Printf("[SendMessage] EXIT session=%s respLen=%d err=%v\n", sessionID, len(resp), err)
	return resp, err
}

// RetryMessage reprocessa o último turno do usuário. Mock: mesmo que Send.
func (a *App) RetryMessage(sessionID, text string) (string, error) {
	ctx := context.Background()
	return a.eng.Chat.Send(ctx, sessionID, text, "", "", "")
}

// AnswerQuestion responde a uma pergunta pendente. Mock: noop.
func (a *App) AnswerQuestion(sessionID, answer string) error {
	a.eng.Saver.AppendMessage(sessionID, core.RawMessage{
		Role: "user", Content: answer,
	})
	return nil
}

// AnswerApproval aprova/rejeita uso de ferramenta. Mock: noop.
func (a *App) AnswerApproval(requestID string, approved bool, reason string) error {
	return nil
}

// StopGeneration interrompe geração em andamento.
func (a *App) StopGeneration(sessionID string) error {
	a.eng.Chat.Stop(sessionID)
	return nil
}

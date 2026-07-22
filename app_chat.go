package main

import (
	"context"
	"fmt"

	core "ada-love-core"
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

	// Resolve context size from session's model (modelOverride = "provider/model")
	ctxSize := 0
	if modelOverride != "" {
		if cs, ok := a.eng.GetModelSettings(modelOverride); ok && cs > 0 {
			ctxSize = cs
		}
	}

	resp, err := a.eng.Chat.Send(ctx, sessionID, text, modelOverride, thinkingLevel, mode, ctxSize)
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

// RespondPermission processa a resposta do usuário a um pedido de permissão.
func (a *App) RespondPermission(requestID, decision string) {
	if a.eng.Chat == nil {
		return
	}
	a.eng.Chat.RespondPermission(requestID, decision)
}

// StopGeneration interrompe geração em andamento.
func (a *App) StopGeneration(sessionID string) error {
	a.eng.Chat.Stop(sessionID)
	return nil
}

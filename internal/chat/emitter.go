package chat

// EmitDecisionEvent envia um evento de decisão do orchestrador
// com a cadeia de pensamento do modelo para o frontend.
// func (we *WailsEmitter) EmitDecisionEvent(sessionID, reasoning, nextAgent, task string, subTasks int) {
//	// Envia para o frontend via Wails runtime.EventsEmit
//	runtime.EventsEmit(runtime.Context(), "orchestrator:decision", map[string]interface{}{
//		"session_id": sessionID,
//		"reasoning":  reasoning,
//		"nextAgent":  nextAgent,
//		"task":       task,
//		"subTasks":   subTasks,
//	})
//	fmt.Printf("[WailsEmitter] Emitting orchestrator:decision for session=%s\n", sessionID)
//}

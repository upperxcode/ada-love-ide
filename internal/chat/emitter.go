package chat

import (
	"fmt"
)

// EmitDecisionEvent envia um evento de decisão do orchestrator
// com a cadeia de pensamento do modelo.
func (fe *FrontendEmitter) EmitDecisionEvent(sessionID, reasoning, nextAgent, task string, subTasks int) {
	fmt.Printf("[FrontendEmitter] Emitting orchestrator:decision for session=%s\n", sessionID)
}

package core

import (
	"context"
	"strings"

	commands "github.com/upperxcode/ada-commands"
)

// ExecutorAdapter adapta o ada-commands.CommandRouter para a interface core.Executor.
type ExecutorAdapter struct {
	router    *commands.CommandRouter
	workspace string
}

// NewExecutorAdapter cria um novo adapter com o router fornecido.
func NewExecutorAdapter(router *commands.CommandRouter) *ExecutorAdapter {
	return &ExecutorAdapter{
		router:    router,
		workspace: "",
	}
}

// NewExecutorAdapterWithWorkspace cria um adapter com workspace customizado.
func NewExecutorAdapterWithWorkspace(router *commands.CommandRouter, workspace string) *ExecutorAdapter {
	return &ExecutorAdapter{
		router:    router,
		workspace: workspace,
	}
}

// ExecuteCommand implementa core.Executor.
// Recebe o sessionID e delega ao CommandRouter.
func (e *ExecutorAdapter) ExecuteCommand(ctx context.Context, sessionID, cmd string, args []string) (string, error) {
	// Constrói o rawInput: "/cmd arg1 arg2 ..."
	var rawInput string
	if len(args) > 0 {
		rawInput = "/" + cmd + " " + strings.Join(args, " ")
	} else {
		rawInput = "/" + cmd
	}

	// Executa via router com o sessionID correto
	return e.router.Execute(ctx, sessionID, e.workspace, rawInput)
}

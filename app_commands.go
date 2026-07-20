package main

import (
	"ada-love-ide/internal/config/command"
)

// ListCommands retorna os slash commands conhecidos (apenas comandos de chat).
func (a *App) ListCommands() []command.CommandInfo {
	var result []command.CommandInfo
	for _, cmd := range a.eng.Router.ListChatCommands() {
		result = append(result, command.CommandInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Usage:       "/" + cmd.Name(),
			Aliases:     []string{},
		})
	}
	return result
}

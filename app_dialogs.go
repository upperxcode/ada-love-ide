package main

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OpenDirectoryDialog abre um seletor de diretório nativo.
func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar diretório",
	})
}

// OpenFileDialog abre um seletor de arquivo nativo.
func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar arquivo",
	})
}

// GetEnvProviderKeys lista env vars passíveis de serem chaves.
func (a *App) GetEnvProviderKeys() []string {
	return []string{
		"OPENAI_API_KEY",
		"OPENROUTER_API_KEY",
		"ANTHROPIC_API_KEY",
		"OPENAI_BASE_URL",
	}
}

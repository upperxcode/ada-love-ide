package chat

import stream "github.com/upperxcode/ada-stream"

type ChatMode string

const (
	ModeAsk  ChatMode = "ASK"
	ModeEdit ChatMode = "EDIT"
	ModePlan ChatMode = "PLAN"
	ModeFull ChatMode = "FULL"
)

type ModeConfig struct {
	Mode            ChatMode
	SystemPrompt    string
	AllowedTools    []string
	NeedsPermission bool
	CanOverrideOnce bool
}

func GetModeConfig(mode ChatMode) ModeConfig {
	switch mode {
	case ModeAsk:
		return ModeConfig{
			Mode: ModeAsk,
			SystemPrompt: `Você é um assistente puramente informativo.
Você pode LER arquivos e PESQUISAR (web/local) para responder perguntas.
Você NÃO pode modificar, criar ou executar nada.
Mantenha as respostas claras e diretas.`,
			AllowedTools:    []string{"read", "search"},
			NeedsPermission: false,
			CanOverrideOnce: true,
		}
	case ModePlan:
		return ModeConfig{
			Mode: ModePlan,
			SystemPrompt: `Você é um arquiteto de software.
Seu objetivo é ANALISAR o código, EXPLORAR arquivos e CRIAR PLANOS detalhados.
Você NÃO executa alterações — apenas documenta o que precisa ser feito.
Use a pasta plan/ para salvar os planos.`,
			AllowedTools:    []string{"read", "search"},
			NeedsPermission: false,
			CanOverrideOnce: true,
		}
	case ModeEdit:
		return ModeConfig{
			Mode: ModeEdit,
			SystemPrompt: `Você é um desenvolvedor editor de código.
Você pode ler, pesquisar, editar e criar arquivos no diretório de trabalho.
COMANDOS DE TERMINAL e ESCRITA FORA DO WORKSPACE precisam de confirmação do usuário.
Seja preciso nos patches e alterações.`,
			AllowedTools:    []string{"read", "search", "write"},
			NeedsPermission: true,
			CanOverrideOnce: true,
		}
	case ModeFull:
		return ModeConfig{
			Mode: ModeFull,
			SystemPrompt: `Você é um agente autônomo completo.
Pode planejar, explorar, editar arquivos e executar comandos no terminal.
Todas as ferramentas estão disponíveis. Seja eficiente.`,
			AllowedTools:    []string{"read", "search", "write", "exec", "plan"},
			NeedsPermission: false,
			CanOverrideOnce: false,
		}
	default:
		return GetModeConfig(ModeAsk)
	}
}

func GetSystemPrompt(mode ChatMode, basePrompt string) string {
	if basePrompt != "" {
		return basePrompt
	}
	return GetModeConfig(mode).SystemPrompt
}

func (m ChatMode) IsValid() bool {
	switch m {
	case ModeAsk, ModeEdit, ModePlan, ModeFull:
		return true
	}
	return false
}

func AllowedChunkTypes(mode ChatMode) []stream.ChunkType {
	switch mode {
	case ModeAsk:
		return []stream.ChunkType{stream.ChunkContent, stream.ChunkExplore, stream.ChunkRead, stream.ChunkThought}
	case ModePlan:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkRead, stream.ChunkThought}
	case ModeEdit:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkRead, stream.ChunkDiff, stream.ChunkThought, stream.ChunkContent}
	case ModeFull:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkExec, stream.ChunkRead, stream.ChunkDiff, stream.ChunkThought, stream.ChunkContent}
	}
	return nil
}

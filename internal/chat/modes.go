package chat

import (
	"encoding/json"
	"strings"

	stream "github.com/upperxcode/ada-stream"
)

// ChatMode representa os modos de operação do assistente.
type ChatMode string

const (
	ModeAsk    ChatMode = "ASK"
	ModeEdit   ChatMode = "EDIT"
	ModePlan   ChatMode = "PLAN"
	ModeFull   ChatMode = "FULL"
	ModeExec   ChatMode = "EXECUTE"
	ModeAdmin  ChatMode = "ADMIN"
)

// RiskLevel classifica o nível de risco de uma ação.
type RiskLevel int

const (
	RiskNone   RiskLevel = 0
	RiskLow    RiskLevel = 1
	RiskMedium RiskLevel = 2
	RiskHigh   RiskLevel = 3
	RiskCritical RiskLevel = 4
)

func (r RiskLevel) String() string {
	switch r {
	case RiskNone:
		return "none"
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	case RiskCritical:
		return "critical"
	}
	return "unknown"
}

// TTLPolicy define por quanto tempo uma permissão concedida é válida.
type TTLPolicy string

const (
	TTLSession   TTLPolicy = "session"    // dura até o fim da sessão do usuário
	TTLTask      TTLPolicy = "task"       // dura até o fim da tarefa atual
	TTLAction    TTLPolicy = "action"     // válida para uma única ação (diff/arquivo)
	TTLTemporary TTLPolicy = "temporary"  // dura N minutos (configurável)
	TTLPermanent TTLPolicy = "permanent"  // nunca expira (com confirmação extra)
)

// ToolCapability categoriza as ferramentas por capacidade.
type ToolCapability string

const (
	CapRead   ToolCapability = "read"
	CapSearch ToolCapability = "search"
	CapWrite  ToolCapability = "write"
	CapExec   ToolCapability = "exec"
	CapPlan   ToolCapability = "plan"
	CapAdmin  ToolCapability = "admin"
	CapConfig ToolCapability = "config"
)

// ActionClass classifica uma ação para fins de permissão.
type ActionClass string

const (
	ActionRead         ActionClass = "read"
	ActionSearch       ActionClass = "search"
	ActionWriteProject ActionClass = "write_project"
	ActionWriteEnv     ActionClass = "write_env"
	ActionWriteOutside ActionClass = "write_outside"
	ActionExec         ActionClass = "exec"
	ActionExecHighRisk ActionClass = "exec_high_risk"
	ActionMkdir        ActionClass = "mkdir"
	ActionAdmin        ActionClass = "admin"
	ActionConfigEdit   ActionClass = "config_edit"
	ActionNetwork      ActionClass = "network"
)

// ModeConfig carrega toda a configuração de um modo.
type ModeConfig struct {
	Mode                ChatMode
	SystemPrompt        string
	AllowedCapabilities []ToolCapability
	NeedsPermission     bool          // se precisa de consentimento do usuário
	DefaultTTL          TTLPolicy     // TTL padrão para grants deste modo
	CanOverrideOnce     bool          // permite ao usuário permitir uma vez fora do escopo
	AllowSessionGrant   bool          // permite "always allow in this session"
	RequiresExplicitActivation bool   // exige confirmação para entrar no modo (ex: FULL, ADMIN)
	DeniedActions       []ActionClass // ações bloqueadas mesmo com permissão
	DeniedCommands      []string      // comandos shell bloqueados (palavras-chave)
}

// RiskConfig associa nível de risco a cada ação em cada modo.
type RiskConfig struct {
	Action      ActionClass
	RiskByMode  map[ChatMode]RiskLevel
	Description string
}

// DefaultRiskMatrix define o risco de cada ação por modo.
var DefaultRiskMatrix = []RiskConfig{
	{Action: ActionRead, Description: "Leitura de arquivos do projeto",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskNone, ModePlan: RiskNone, ModeEdit: RiskLow, ModeExec: RiskLow, ModeFull: RiskNone, ModeAdmin: RiskNone}},
	{Action: ActionSearch, Description: "Busca no código/base de conhecimento",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskNone, ModePlan: RiskNone, ModeEdit: RiskLow, ModeExec: RiskLow, ModeFull: RiskNone, ModeAdmin: RiskNone}},
	{Action: ActionWriteProject, Description: "Escrita de arquivos do projeto (dentro do workspace)",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskHigh, ModePlan: RiskHigh, ModeEdit: RiskMedium, ModeExec: RiskLow, ModeFull: RiskLow, ModeAdmin: RiskMedium}},
	{Action: ActionWriteEnv, Description: "Escrita em arquivos de ambiente/config (.env, chaves)",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskCritical, ModeExec: RiskCritical, ModeFull: RiskHigh, ModeAdmin: RiskMedium}},
	{Action: ActionWriteOutside, Description: "Escrita fora do workspace",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskCritical, ModeExec: RiskCritical, ModeFull: RiskHigh, ModeAdmin: RiskMedium}},
	{Action: ActionExec, Description: "Execução de comandos no terminal",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskHigh, ModeExec: RiskMedium, ModeFull: RiskLow, ModeAdmin: RiskMedium}},
	{Action: ActionExecHighRisk, Description: "Comandos destrutivos (rm -rf, git push --force, sudo, docker rm -f, etc.)",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskCritical, ModeExec: RiskCritical, ModeFull: RiskHigh, ModeAdmin: RiskCritical}},
	{Action: ActionMkdir, Description: "Criação de diretórios",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskMedium, ModePlan: RiskMedium, ModeEdit: RiskLow, ModeExec: RiskLow, ModeFull: RiskLow, ModeAdmin: RiskNone}},
	{Action: ActionAdmin, Description: "Alteração de configurações do sistema/provedor",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskCritical, ModeExec: RiskCritical, ModeFull: RiskCritical, ModeAdmin: RiskHigh}},
	{Action: ActionConfigEdit, Description: "Alteração de configuração interna da IA (system prompts, MCPs, API keys)",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskCritical, ModePlan: RiskCritical, ModeEdit: RiskCritical, ModeExec: RiskCritical, ModeFull: RiskCritical, ModeAdmin: RiskHigh}},
	{Action: ActionNetwork, Description: "Requisições de rede externas (web search, API calls)",
		RiskByMode: map[ChatMode]RiskLevel{ModeAsk: RiskLow, ModePlan: RiskLow, ModeEdit: RiskLow, ModeExec: RiskLow, ModeFull: RiskLow, ModeAdmin: RiskMedium}},
}

var highRiskPrefixes = []string{
	"rm ", "rm -rf", "rm -r", "rm -f",
	"git push --force", "git push -f", "git reset --hard",
	"sudo ", "chmod 777", "chown ",
	"docker rm -f", "docker system prune", "docker rmi",
	"dd ", "mkfs.", "fdisk", "parted",
	":(){ :|:& };:",   // fork bomb
	"> /dev/", "> /dev/sd", "> /dev/nvme",
	"mv /", "cp -rf /",
	"curl -s http", "wget -q ",
	"eval ", "source /", ". /",
}

// GetModeConfig retorna a configuração completa para um modo.
func GetModeConfig(mode ChatMode) ModeConfig {
	switch mode {
	case ModeAsk:
		return ModeConfig{
			Mode: ModeAsk,
			SystemPrompt: `Você é um assistente puramente informativo.
Você pode LER arquivos e PESQUISAR (web/local) para responder perguntas.
Você NÃO pode modificar, criar ou executar nada.
Mantenha as respostas claras e diretas.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch},
			NeedsPermission:         false,
			DefaultTTL:              TTLSession,
			CanOverrideOnce:         true,
			AllowSessionGrant:       true,
			RequiresExplicitActivation: false,
			DeniedActions:           []ActionClass{ActionExec, ActionExecHighRisk, ActionWriteProject, ActionWriteOutside, ActionWriteEnv, ActionAdmin, ActionConfigEdit},
		}
	case ModePlan:
		return ModeConfig{
			Mode: ModePlan,
			SystemPrompt: `Você é um arquiteto de software.
Seu objetivo é ANALISAR o código, EXPLORAR arquivos e CRIAR PLANOS detalhados.
Você NÃO executa alterações — apenas documenta o que precisa ser feito.
Use a pasta plan/ para salvar os planos.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch},
			NeedsPermission:         false,
			DefaultTTL:              TTLTask,
			CanOverrideOnce:         true,
			AllowSessionGrant:       false,
			RequiresExplicitActivation: false,
			DeniedActions:           []ActionClass{ActionExec, ActionExecHighRisk, ActionWriteProject, ActionWriteOutside, ActionWriteEnv, ActionAdmin, ActionConfigEdit},
		}
	case ModeEdit:
		return ModeConfig{
			Mode: ModeEdit,
			SystemPrompt: `Você é um desenvolvedor editor de código.
Você pode ler, pesquisar, editar e criar arquivos.
COMANDOS DE TERMINAL e ESCRITA FORA DO WORKSPACE precisam de confirmação do usuário.
Use as ferramentas normally (read, write, exec, search, plan). Se um comando precisar de permissão, o sistema de segurança vai pedir automaticamente. Não pergunte antes — apenas execute.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch, CapWrite, CapPlan},
			NeedsPermission:         true,
			DefaultTTL:              TTLAction,
			CanOverrideOnce:         true,
			AllowSessionGrant:       true,
			RequiresExplicitActivation: false,
			DeniedActions:           []ActionClass{ActionExec, ActionExecHighRisk, ActionAdmin, ActionConfigEdit},
		}
	case ModeExec:
		return ModeConfig{
			Mode: ModeExec,
			SystemPrompt: `Você é um assistente de teste e execução controlada.
Pode LER, EDITAR arquivos e EXECUTAR uma lista restrita de comandos seguros:
- go test / npm test / cargo check / pytest
- go build / npm run build / cargo build
- git status / git diff / git log
- ls, cat, head, tail, grep, find
COMANDOS DESTRUTIVOS (rm -rf, git push --force, sudo) são bloqueados automaticamente.
Use as ferramentas normally. Comandos seguros serão executados automaticamente. Comandos de alto risco serão bloqueados ou pedirão confirmação. Não pergunte antes — apenas execute.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan},
			NeedsPermission:         true,
			DefaultTTL:              TTLTask,
			CanOverrideOnce:         false,
			AllowSessionGrant:       false,
			RequiresExplicitActivation: false,
			DeniedActions:           []ActionClass{ActionExecHighRisk, ActionWriteEnv, ActionWriteOutside, ActionAdmin, ActionConfigEdit},
			DeniedCommands:          highRiskPrefixes,
		}
	case ModeFull:
		return ModeConfig{
			Mode: ModeFull,
			SystemPrompt: `Você é um agente autônomo completo.
Pode planejar, explorar, editar arquivos e executar comandos no terminal.
Todas as ferramentas estão disponíveis.
ATENÇÃO: Comandos de alto risco (rm -rf, git push --force, sudo) exigem confirmação.
Seja eficiente e responsável.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan},
			NeedsPermission:         false,
			DefaultTTL:              TTLTemporary,
			CanOverrideOnce:         false,
			AllowSessionGrant:       false,
			RequiresExplicitActivation: true,
			DeniedActions:           []ActionClass{ActionAdmin, ActionConfigEdit},
			DeniedCommands:          nil, // FULL permite tudo, mas high-risk pede confirmação
		}
	case ModeAdmin:
		return ModeConfig{
			Mode: ModeAdmin,
			SystemPrompt: `Você é um administrador do sistema ADA.
Pode gerenciar configurações de provedores, modelos, MCPs, skills e permissões.
Ações DESTRUTIVAS (excluir provedores, resetar configurações) exigem confirmação dupla.
CUIDADO: alterações em system prompts e API keys afetam todo o sistema.`,
			AllowedCapabilities:     []ToolCapability{CapRead, CapSearch, CapWrite, CapExec, CapPlan, CapAdmin, CapConfig},
			NeedsPermission:         false,
			DefaultTTL:              TTLTemporary,
			CanOverrideOnce:         false,
			AllowSessionGrant:       false,
			RequiresExplicitActivation: true,
			DeniedActions:           nil,
		}
	default:
		return GetModeConfig(ModeAsk)
	}
}

// GetSystemPrompt retorna o system prompt para um modo.
// Se basePrompt não for vazio, usa ele; senão usa o default do modo.
func GetSystemPrompt(mode ChatMode, basePrompt string) string {
	if basePrompt != "" {
		return basePrompt
	}
	return GetModeConfig(mode).SystemPrompt
}

// IsValid verifica se o modo é reconhecido.
func (m ChatMode) IsValid() bool {
	switch m {
	case ModeAsk, ModeEdit, ModePlan, ModeFull, ModeExec, ModeAdmin:
		return true
	}
	return false
}

// AllowedChunkTypes retorna os tipos de chunk de streaming permitidos no modo.
func AllowedChunkTypes(mode ChatMode) []stream.ChunkType {
	switch mode {
	case ModeAsk:
		return []stream.ChunkType{stream.ChunkContent, stream.ChunkExplore, stream.ChunkRead, stream.ChunkThought}
	case ModePlan:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkRead, stream.ChunkThought}
	case ModeEdit:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkRead, stream.ChunkDiff, stream.ChunkThought, stream.ChunkContent}
	case ModeExec:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkExec, stream.ChunkRead, stream.ChunkDiff, stream.ChunkThought, stream.ChunkContent}
	case ModeFull:
		return []stream.ChunkType{stream.ChunkPlan, stream.ChunkExplore, stream.ChunkExec, stream.ChunkRead, stream.ChunkDiff, stream.ChunkThought, stream.ChunkContent}
	case ModeAdmin:
		return []stream.ChunkType{stream.ChunkContent, stream.ChunkExplore, stream.ChunkRead, stream.ChunkThought}
	}
	return nil
}

// GetCapabilityForTool mapeia um nome de ferramenta para sua capacidade.
func GetCapabilityForTool(toolName string) ToolCapability {
	switch toolName {
	case "read", "read_file", "cat":
		return CapRead
	case "search", "grep", "find", "explore":
		return CapSearch
	case "write", "write_file", "edit", "patch", "create":
		return CapWrite
	case "exec", "execute", "run", "shell", "terminal", "command":
		return CapExec
	case "plan", "planning":
		return CapPlan
	case "admin", "admin_action", "manage":
		return CapAdmin
	case "config", "configure", "settings":
		return CapConfig
	default:
		return CapExec // fallback seguro: tratar como execução
	}
}

// ClassifyAction classifica toolName + args em uma ActionClass.
func ClassifyAction(toolName, argsJSON string, mode ChatMode) (ActionClass, RiskLevel) {
	cap := GetCapabilityForTool(toolName)
	targetPath := extractPath(argsJSON)
	command := extractCommand(argsJSON)

	// Se a ferramenta não for exec, mapeamento direto
	switch cap {
	case CapRead:
		return ActionRead, GetRisk(ActionRead, mode)
	case CapSearch:
		return ActionSearch, GetRisk(ActionSearch, mode)
	case CapPlan:
		return ActionRead, GetRisk(ActionRead, mode) // plan usa read+write de arquivos .md
	case CapAdmin:
		return ActionAdmin, GetRisk(ActionAdmin, mode)
	case CapConfig:
		return ActionConfigEdit, GetRisk(ActionConfigEdit, mode)
	case CapWrite:
		if isEnvFile(targetPath) {
			return ActionWriteEnv, GetRisk(ActionWriteEnv, mode)
		}
		if isOutsideWorkspace(targetPath) {
			return ActionWriteOutside, GetRisk(ActionWriteOutside, mode)
		}
		return ActionWriteProject, GetRisk(ActionWriteProject, mode)
	case CapExec:
		if command != "" && isHighRiskCommand(command) {
			return ActionExecHighRisk, GetRisk(ActionExecHighRisk, mode)
		}
		return ActionExec, GetRisk(ActionExec, mode)
	}

	return ActionExec, GetRisk(ActionExec, mode)
}

// GetRisk retorna o nível de risco de uma ação no modo atual.
func GetRisk(action ActionClass, mode ChatMode) RiskLevel {
	for _, rc := range DefaultRiskMatrix {
		if rc.Action == action {
			if risk, ok := rc.RiskByMode[mode]; ok {
				return risk
			}
			return RiskMedium // default seguro
		}
	}
	return RiskMedium
}

// isHighRiskCommand verifica se um comando shell é de alto risco.
func isHighRiskCommand(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, prefix := range highRiskPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// isEnvFile verifica se o caminho é um arquivo de ambiente/config sensível.
func isEnvFile(path string) bool {
	if path == "" {
		return false
	}
	name := strings.ToLower(path)
	sensitive := []string{
		".env", ".env.", ".envrc",
		".gitconfig", ".ssh/", "id_rsa", "id_ed25519",
		"credentials", "secrets", "token", "apikey",
		".npmrc", ".netrc",
		"config.json", "settings.json", ".vscode/",
		"ci/", ".github/", "gitlab-ci", ".circleci",
	}
	for _, s := range sensitive {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}

// isOutsideWorkspace verifica se o caminho está fora do workspace.
// Por simplicidade, verifica se começa com / (caminho absoluto fora do projeto).
func isOutsideWorkspace(path string) bool {
	if path == "" {
		return false
	}
	// Caminhos relativos são considerados dentro do workspace
	if strings.HasPrefix(path, "/") {
		return true
	}
	return false
}

// extractCommand extrai o comando de uma string JSON de argumentos.
func extractCommand(argsJSON string) string {
	if argsJSON == "" {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &data); err != nil {
		return ""
	}
	if cmd, ok := data["command"].(string); ok {
		return cmd
	}
	return ""
}

package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	commands "github.com/upperxcode/ada-commands"
	llm "github.com/upperxcode/ada-llm-client"
	wiki "github.com/upperxcode/ada-llm-wiki"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Orchestrator é o pipeline central de processamento de mensagens.
type Orchestrator struct {
	Storage   StorageEngine
	LLMClient LLMClient
	Compactor Compactor
	Executor  Executor
	Wiki      *wiki.WikiManager
	emitter   Emitter
}

// NewOrchestrator cria um orchestrator com injeção de dependências.
func NewOrchestrator(storage StorageEngine, llm LLMClient, compactor Compactor, executor Executor) *Orchestrator {
	return &Orchestrator{
		Storage:   storage,
		LLMClient: llm,
		Compactor: compactor,
		Executor:  executor,
		// O emitter será configurado posteriormente via SetEmitter
	}
}

// SetEmitter permite configurar o emitter após a criação do Orchestrator.
func (o *Orchestrator) SetEmitter(emitter Emitter) {
	o.emitter = emitter
}

// Emit envia eventos através do emitter configurado.
func (o *Orchestrator) Emit(event string, data ...any) {
	if o.emitter != nil {
		o.emitter.Emit(event, data...)
	}
}

// CompilePrompt monta o prompt final integrando wiki, histórico compactado e input do usuário.
// Respeita o orçamento de tokens descontando o overhead da wiki do budget disponível para histórico.
func (o *Orchestrator) CompilePrompt(ctx context.Context, sessionID, userInput string, history []Message) (string, error) {
	var wikiBlock strings.Builder

	if o.Wiki != nil {
		articles := o.Wiki.Search(ctx, userInput)
		if len(articles) > 0 {
			wikiBlock.WriteString("\n--- START INTERNAL WIKI CONTEXT ---\n")
			for _, art := range articles {
				wikiBlock.WriteString(fmt.Sprintf("### Wiki: %s\nTags: %s\n%s\n\n", art.Title, strings.Join(art.Tags, ", "), art.Content))
			}
			wikiBlock.WriteString("--- END INTERNAL WIKI CONTEXT ---\n")
		}
	}
	wikiText := wikiBlock.String()

	wikiOverhead := 0
	if wikiText != "" {
		wikiOverhead = o.Compactor.CountTokens(wikiText)
	}

	compactedHistory, err := o.Compactor.CompactWithOverhead(ctx, history, wikiOverhead)
	if err != nil {
		return "", fmt.Errorf("compact with overhead: %w", err)
	}

	finalPrompt := compactedHistory + wikiText + "\nUser: " + userInput + "\nAssistant:"

	finalTokens := o.Compactor.CountTokens(finalPrompt)
	if maxTokens := 8000; finalTokens > maxTokens {
		charsPerToken := 4
		maxChars := maxTokens * charsPerToken
		if len(finalPrompt) > maxChars {
			finalPrompt = finalPrompt[:maxChars-3] + "..."
		}
	}

	return finalPrompt, nil
}

// ProcessMessage processa uma mensagem completa (sem streaming).
func (o *Orchestrator) ProcessMessage(ctx context.Context, sessionID, userInput, model string) (string, error) {
	// Check for command prefix first
	if strings.HasPrefix(strings.TrimSpace(userInput), "/") {
		cmdName, args := commands.ParseCommand(userInput)
		if cmdName != "" {
			response, err := o.Executor.ExecuteCommand(ctx, sessionID, cmdName, args)
			if err != nil {
				return "", err
			}
			if response != "" {
				return response, nil
			}
		}
	}

	if response, matched := o.checkStaticRoutes(userInput); matched {
		return response, nil
	}

	history, err := o.Storage.GetMessagesBySession(sessionID)
	if err != nil {
		return "", fmt.Errorf("history: %w", err)
	}

	fullPrompt, err := o.CompilePrompt(ctx, sessionID, userInput, history)
	if err != nil {
		return "", fmt.Errorf("compile: %w", err)
	}

	response, err := o.generateWithLLM(ctx, fullPrompt, model)
	if err != nil {
		return "", fmt.Errorf("llm: %w", err)
	}

	o.SaveMessages(sessionID, userInput, response)
	return response, nil
}

// ProcessMessageStream processa uma mensagem com streaming.
// Não persiste — o caller é responsável por salvar.
func (o *Orchestrator) ProcessMessageStream(ctx context.Context, sessionID, userInput, model string) (<-chan LLMToken, error) {
	fmt.Printf("[Orchestrator.ProcessMessageStream] ENTER session=%s input=%q model=%s\n", sessionID, userInput[:min(50, len(userInput))], model)

	// Check for command prefix first
	if strings.HasPrefix(strings.TrimSpace(userInput), "/") {
		cmdName, args := commands.ParseCommand(userInput)
		if cmdName != "" {
			response, err := o.Executor.ExecuteCommand(ctx, sessionID, cmdName, args)
			if err != nil {
				return nil, err
			}
			if response != "" {
				fmt.Printf("[Orchestrator] Command matched: %s\n", response[:min(50, len(response))])
				ch := make(chan LLMToken, 1)
				go func() {
					defer close(ch)
					ch <- LLMToken{Token: response, Done: true}
				}()
				return ch, nil
			}
		}
	}

	if response, matched := o.checkStaticRoutes(userInput); matched {
		fmt.Printf("[Orchestrator] Static route matched: %s\n", response[:min(50, len(response))])
		ch := make(chan LLMToken, 1)
		go func() {
			defer close(ch)
			ch <- LLMToken{Token: response, Done: true}
		}()
		return ch, nil
	}

	history, err := o.Storage.GetMessagesBySession(sessionID)
	if err != nil {
		fmt.Printf("[Orchestrator] GetMessagesBySession ERROR: %v\n", err)
		return nil, fmt.Errorf("history: %w", err)
	}

	fullPrompt, err := o.CompilePrompt(ctx, sessionID, userInput, history)
	if err != nil {
		fmt.Printf("[Orchestrator] CompilePrompt ERROR: %v\n", err)
		return nil, fmt.Errorf("compile: %w", err)
	}
	fmt.Printf("[Orchestrator] Prompt length: %d\n", len(fullPrompt))

	// Call LLM client
	ch, err := o.LLMClient.Generate(ctx, fullPrompt, model)
	if err != nil {
		fmt.Printf("[Orchestrator] LLMClient.Generate ERROR: %v\n", err)
		return nil, err
	}

	fmt.Printf("[Orchestrator] LLMClient.Generate returned channel\n")
	return ch, nil
}

func (o *Orchestrator) checkStaticRoutes(userInput string) (string, bool) {
	// Fallback: use the embedded static router from ada-llm-client
	if response, ok := llm.CheckStaticResponse(userInput); ok {
		return response, true
	}

	return "", false
}

// ExecuteCommand intercepta e executa comandos que começam com "/"
// Este método é público para ser chamado pelo Chat antes de processar via LLM.
func (o *Orchestrator) ExecuteCommand(ctx context.Context, sessionID, cmdName string, args []string) (string, error) {
	result, err := o.Executor.ExecuteCommand(ctx, sessionID, cmdName, args)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (o *Orchestrator) formatHistory(history []Message) []string {
	out := make([]string, 0, len(history))
	for _, msg := range history {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		out = append(out, fmt.Sprintf("%s: %s", role, msg.Content))
	}
	return out
}

func (o *Orchestrator) generateWithLLM(ctx context.Context, prompt string, model string) (string, error) {
	tokens, err := o.LLMClient.Generate(ctx, prompt, model)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for token := range tokens {
		if !token.Done {
			b.WriteString(token.Token)
		}
	}
	return b.String(), nil
}

func (o *Orchestrator) SaveMessages(sessionID, userInput, response string) {
	now := time.Now().Format(time.RFC3339)
	_ = o.Storage.SaveMessage(Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   userInput,
		CreatedAt: now,
	})
	_ = o.Storage.SaveMessage(Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   response,
		CreatedAt: now,
	})
}

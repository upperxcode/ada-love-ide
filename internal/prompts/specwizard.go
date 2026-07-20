package prompts

import "fmt"

var promptBuilders = map[string]PromptBuilder{
	"PRD":               buildPRD,
	"FUNCTIONAL":        buildFunctional,
	"NON_FUNCTIONAL":    buildNonFunctional,
	"API_CONTRACT":      buildAPIContract,
	"CUSTOMIZATION":     buildCustomization,
	"FINAL_ADJUSTMENTS": buildFinalAdjustments,
}

func Build(target string, ctx FieldContext, currentValue string) (Prompt, error) {
	builder, ok := promptBuilders[target]
	if !ok {
		return Prompt{}, fmt.Errorf("unknown prompt target: %s", target)
	}
	return builder(ctx, currentValue), nil
}

const aiTargetSuffix = `
CRITICAL: Output must be concise, structured, and machine-readable. This text will be consumed by another AI model — not by humans. Avoid fluff, marketing language, adjectives, and verbose explanations. Use short sentences, bullet points, and clear section headers.`

func buildPRD(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no initial description provided)"
	}

	return Prompt{
		SystemPrompt: `You are an Elite Product Manager. Produce a concise, engineering-focused Product Requirements Document for another AI to consume.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
EXPERT PLUGIN: %s

USER'S INITIAL IDEA:
%s

Instructions:
- Define the software's ultimate goal and domain boundary
- List core user personas and their primary goals
- Establish strict business rules and constraints
- Focus on engineering clarity; avoid vague product descriptions
- Output in structured markdown sections
%s`,
			ctx.SpecName, ctx.ExpertLanguagePlugin, userVal, aiTargetSuffix),
		Temperature: 0.2,
		MaxTokens:   2000,
	}
}

func buildFunctional(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no initial draft provided)"
	}

	return Prompt{
		SystemPrompt: `You are a System Analyst. Expand requirements into concrete, testable functional pipelines with clear state transitions.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
EXPERT PLUGIN: %s

PRD:
%s

USER'S INITIAL DRAFT:
%s

Instructions:
- Each requirement must be testable (pass/fail)
- Define success and failure states for each flow
- Use a structured checklist format
- Focus only on functional behavior, not implementation details
%s`,
			ctx.SpecName, ctx.ExpertLanguagePlugin, ctx.PRD, userVal, aiTargetSuffix),
		Temperature: 0.1,
		MaxTokens:   2000,
	}
}

func buildNonFunctional(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no initial draft provided)"
	}

	return Prompt{
		SystemPrompt: `You are a Technical Architect. Define precise non-functional constraints that the system must satisfy.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
EXPERT PLUGIN: %s

PRD:
%s

FUNCTIONAL REQUIREMENTS:
%s

USER'S INITIAL DRAFT:
%s

Instructions:
- Cover: security, performance, scalability, maintainability, availability
- Specify platform and environment constraints
- List compliance and regulatory requirements if applicable
- Each constraint must be measurable or verifiable
%s`,
			ctx.SpecName, ctx.ExpertLanguagePlugin, ctx.PRD, ctx.FunctionalReqs, userVal, aiTargetSuffix),
		Temperature: 0.1,
		MaxTokens:   2000,
	}
}

func buildAPIContract(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no initial draft provided)"
	}

	return Prompt{
		SystemPrompt: `You are a Principal Software Architect. Define strict communication contracts adhering 100% to the chosen architecture, patterns, and stack. Be deterministic and precise.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
EXPERT PLUGIN: %s

ARCHITECTURE: %s
PERSISTENCE: %s
ENGINEERING PHILOSOPHIES: %s
DESIGN PATTERNS: %s | DATA PATTERNS: %s
DEPENDENCIES: %s

SCOPE:
PRD: %s
FUNCTIONAL REQS: %s

USER'S INITIAL DRAFT:
%s

Instructions:
- Define JSON/gRPC payloads with exact field names and types
- Specify status codes and error response formats
- Follow naming conventions aligned with chosen engineering philosophies
- Do not deviate from the declared architecture and patterns
%s`,
			ctx.SpecName, ctx.ExpertLanguagePlugin,
			ctx.Architecture, ctx.Persistence,
			ctx.EngineeringPhilosophies, ctx.DesignPatterns, ctx.DataPatterns,
			ctx.DependencyManifest,
			ctx.PRD, ctx.FunctionalReqs,
			userVal, aiTargetSuffix),
		Temperature: 0.0,
		MaxTokens:   2000,
	}
}

func buildCustomization(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no initial notes provided)"
	}

	return Prompt{
		SystemPrompt: `You are a UI/UX Technical Designer. Detail visual behaviors, component reuse rules, and edge-case handling based on the chosen stack.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
STACK PLUGIN: %s
STATE MANAGEMENT: %s

API CONTRACT:
%s

USER'S NOTES:
%s

Instructions:
- Define component reuse rules (e.g. repeated patterns must become shared components)
- Detail edge cases, loading, empty, and error states
- Specify customization behavior for the chosen state management approach
- Keep recommendations tight to the chosen stack
%s`,
			ctx.SpecName, ctx.StackPlugin, ctx.StateManagement,
			ctx.APIContract,
			userVal, aiTargetSuffix),
		Temperature: 0.2,
		MaxTokens:   2000,
	}
}

func buildFinalAdjustments(ctx FieldContext, currentValue string) Prompt {
	userVal := currentValue
	if userVal == "" {
		userVal = "(no specific instructions provided)"
	}

	return Prompt{
		SystemPrompt: `You are a Senior Implementation Engineer. Generate a structured implementation roadmap as a plan for another AI to execute.`,
		UserPrompt: fmt.Sprintf(`SPEC NAME: %s
EXPERT PLUGIN: %s

FULL CONFIGURATION:
  Architecture: %s
  Persistence: %s
  Engineering Philosophies: %s
  Design Patterns: %s | Data Patterns: %s
  Dependencies: %s
  State Management: %s

API CONTRACT:
%s

CUSTOMIZATION DETAILS:
%s

USER'S INSTRUCTIONS:
%s

Instructions:
- Break down implementation into ordered phases
- Highlight critical integration points and ordering constraints
- Include setup, core domain, infrastructure, and delivery phases
- Keep it practical and actionable
%s`,
			ctx.SpecName, ctx.ExpertLanguagePlugin,
			ctx.Architecture, ctx.Persistence,
			ctx.EngineeringPhilosophies, ctx.DesignPatterns, ctx.DataPatterns,
			ctx.DependencyManifest,
			ctx.StateManagement,
			ctx.APIContract,
			ctx.CustomizationDetails,
			userVal, aiTargetSuffix),
		Temperature: 0.2,
		MaxTokens:   2000,
	}
}

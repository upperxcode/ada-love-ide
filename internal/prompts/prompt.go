package prompts

type Prompt struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

type FieldContext struct {
	SpecName              string
	ExpertLanguagePlugin  string
	PRD                   string
	FunctionalReqs        string
	NonFunctionalReqs     string
	Architecture          string
	Persistence           string
	EngineeringPhilosophies string
	DesignPatterns        string
	DataPatterns          string
	StackPlugin           string
	DependencyManifest    string
	StateManagement       string
	APIContract           string
	CustomizationDetails  string
	FinalAdjustments      string
}

type PromptBuilder func(ctx FieldContext, currentValue string) Prompt

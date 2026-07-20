package specwizard

import "time"

type StackItem struct {
	Name    string `json:"name"`
	Example string `json:"example"`
}

	type Dependency struct {
		Name      string `json:"lib"`
		Version   string `json:"ver"`
		Mandatory bool   `json:"mandatory"`
	}

	type SpecWizardConfig struct {
		ID                        string       `json:"id"`
		Name                      string       `json:"name"`
		Description               string       `json:"description"`
		ExpertLanguagePlugin      string       `json:"expert_language_plugin"`
		PRD                       string       `json:"prd"`
		FunctionalRequirements    []string     `json:"functional_requirements"`
		NonFunctionalRequirements []string     `json:"non_functional_requirements"`
		Persistence               string       `json:"persistence"`
		Architecture              string       `json:"architecture"`
		EngineeringPhilosophies   []string     `json:"engineering_philosophies"`
		DesignPatterns            []string     `json:"design_patterns"`
		DataPatterns              []string     `json:"data_patterns"`
		StackConfig               []StackItem  `json:"stack_config"`
		Business                  Business     `json:"business"`
		Color                     string       `json:"color"`
		Icon                      string       `json:"icon"`
		ArchitectureHealth        int          `json:"architecture_health"`
		DependencyManifest        []Dependency `json:"dependency_manifest"`
		StackPlugin               string       `json:"stack_plugin"`
		CreatedAt                 time.Time    `json:"created_at"`
		UpdatedAt                 time.Time    `json:"updated_at"`
	}

type Business struct {
	StateManagement             string `json:"state_management"`
	APIContract                 string `json:"api_contract"`
	CustomizationDetails        string `json:"customization_details"`
	FinalAdjustments            string `json:"final_adjustments"`
	ArchitectureRecommendations string `json:"architecture_recommendations"`
}

func New(name string) SpecWizardConfig {
	now := time.Now()
	return SpecWizardConfig{
		ID:                        now.Format("20060102150405"),
		Name:                      name,
		Color:                     "#3b82f6",
		Icon:                      "📝",
		StackConfig:               []StackItem{},
		FunctionalRequirements:    []string{},
		NonFunctionalRequirements: []string{},
		EngineeringPhilosophies:   []string{},
		DesignPatterns:            []string{},
		DataPatterns:              []string{},
			Business:                  Business{},
			DependencyManifest:        []Dependency{},
			CreatedAt:                 now,
		UpdatedAt:                 now,
	}
}

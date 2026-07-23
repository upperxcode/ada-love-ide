package workspace

type WorkspaceConfig struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Path             string   `json:"path"`
	Folders          []string `json:"folders"`
	Personality      string   `json:"personality"`
	RoutingRules     string   `json:"routing_rules"`
	Knowledge        []string `json:"knowledge"`
	WorkerNames      []string `json:"worker_names"`
	Skills           []string `json:"skills"`
	Tools            []string `json:"tools"`
	Enabled          bool     `json:"enabled"`
	Color            string   `json:"color"`
	Icon             string   `json:"icon"`
	MaxPrompt        int      `json:"max_prompt"`
	MaxContent       int      `json:"max_content"`
	Commit           bool     `json:"commit"`
	MaxPromptSend    int      `json:"max_prompt_send"`
	CommitChanges    bool     `json:"commit_changes"`
	MaxContextLength int      `json:"max_context_length"`
	SpecWizard       string   `json:"spec_wizard"`
	SpecWizardID     string   `json:"spec_wizard_id"`
	Agents           []string `json:"agents"`
	Summary          string   `json:"summary"`
	SummaryHash      string   `json:"summary_hash"`
}

func New(title, path string) WorkspaceConfig {
	return WorkspaceConfig{
		Title:         title,
		Path:          path,
		Enabled:       true,
		WorkerNames:   []string{},
		Folders:       []string{},
		Knowledge:     []string{},
		Skills:        []string{},
		Tools:         []string{},
		Agents:        []string{},
		MaxPrompt:     4096,
		MaxContent:    8192,
		Commit:        true,
		CommitChanges: true,
	}
}

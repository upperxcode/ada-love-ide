package skill

// SkillConfig represents the configuration/state of a skill
type SkillConfig struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Content     string `json:"content"`
	Color       string `json:"color"`  // Header background color (hex)
	Icon        string `json:"icon"`   // Emoji/icon for the skill
	Active      bool   `json:"active"` // Enabled/disabled state
}

// SkillState represents the enabled/disabled state of a skill (for UI)
type SkillState struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

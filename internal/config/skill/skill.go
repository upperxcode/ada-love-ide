package skill

type SearchResult struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	RegistryName string  `json:"registry_name"`
	Summary      string  `json:"summary"`
	Description  string  `json:"description"`
	Slug         string  `json:"slug"`
	Version      string  `json:"version"`
	Score        float64 `json:"score"`
}

type SkillFullInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Registry    string   `json:"registry"`
	URL         string   `json:"url"`
	Markdown    string   `json:"markdown"`
	Raw         string   `json:"raw"`
	LineCount   int      `json:"line_count"`
	CharCount   int      `json:"char_count"`
	Tags        []string `json:"tags"`
}

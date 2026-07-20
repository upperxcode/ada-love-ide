package worker

	type WorkerConfig struct {
		ID               int64  `json:"id"`
		Name             string `json:"name"`
		Persona          string `json:"persona"`
	Language         string `json:"language"`
	Icon             string `json:"icon"`
	Color            string `json:"color"`
	ConnectionType   string `json:"connection_type"`
	ConnectionName   string `json:"connection_name"`
	ConnectionConfig string `json:"connection_config"`
	InheritFolders   bool   `json:"inherit_folders"`
	InheritKnowledge bool   `json:"inherit_knowledge"`
	InheritSkills    bool   `json:"inherit_skills"`
	InheritTools     bool   `json:"inherit_tools"`
	InheritPersona   bool   `json:"inherit_persona"`
}

func New(name string) WorkerConfig {
	return WorkerConfig{
		Name:             name,
		Icon:             "🤖",
		ConnectionType:   "ada",
		ConnectionName:   "Ada",
		InheritFolders:   true,
		InheritKnowledge: true,
		InheritSkills:    true,
		InheritTools:     true,
		InheritPersona:   true,
	}
}

package plugins

type TestConfig struct {
	Command    string `yaml:"command" json:"command"`
	FailPrompt string `yaml:"fail_prompt" json:"fail_prompt"`
}

type ExpertPlugin struct {
	ID           string      `yaml:"id" json:"id"`
	Name         string      `yaml:"name" json:"name"`
	Description  string      `yaml:"description" json:"description"`
	Language     string      `yaml:"language" json:"language"`
	StartCommand string      `yaml:"start_command" json:"start_command"`
	Triggers     []string    `yaml:"triggers" json:"triggers"`
	TestConfig   *TestConfig `yaml:"test_config" json:"test_config"`
	Endpoint     string      `json:"endpoint,omitempty"`
}

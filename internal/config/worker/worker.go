package worker

import (
	"encoding/json"
)

type WorkerConfig struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Persona          string `json:"persona"`
	Language         string `json:"language"`
	Icon             string `json:"icon"`
	Color            string `json:"color"`
	ConnectionType   string `json:"connection_type"`
	ConnectionName   string `json:"connection_name"`
	Command          string `json:"command"`
	Arguments        string `json:"arguments"`
	ModelsCommand    string `json:"models_command"`
	Environment      string `json:"environment"`
	InheritFolders   bool   `json:"inherit_folders"`
	InheritKnowledge bool   `json:"inherit_knowledge"`
	InheritSkills    bool   `json:"inherit_skills"`
	InheritTools     bool   `json:"inherit_tools"`
	InheritPersona   bool   `json:"inherit_persona"`
}

type URLPaths struct {
	ChatPath         string `json:"chat_path"`
	ChatBodyTemplate string `json:"chat_body_template"`
	ModelsPath       string `json:"models_path"`
	ModelsFormat     string `json:"models_format"`
	Stream           bool   `json:"stream"`
	StartCommand     string `json:"start_command"`
}

// EncodeArguments serializa Arguments e ModelsCommand em uma única string para o DB.
// Convenção:
//
//	CLI: primeira linha = Arguments, segunda linha = ModelsCommand.
//	URL: Arguments = JSON com URLPaths, ModelsCommand = path/models URL.
func (w WorkerConfig) EncodeArguments() string {
	if w.ConnectionType == "url" {
		return w.encodeURLArguments()
	}
	if w.ModelsCommand == "" {
		return w.Arguments
	}
	return w.Arguments + "\n" + w.ModelsCommand
}

func (w WorkerConfig) encodeURLArguments() string {
	cfg := URLPaths{
		ChatPath:         "/v1/chat/completions",
		ChatBodyTemplate: `{"model":"{{.Model}}","messages":[{"role":"user","content":"{{.Message}}"}],"stream":{{.Stream}}}`,
		ModelsPath:       "/config/providers",
		ModelsFormat:     "providers_obj",
		Stream:           false,
		StartCommand:     "",
	}
	if w.Arguments != "" && w.Arguments != "{}" {
		var existing URLPaths
		if err := json.Unmarshal([]byte(w.Arguments), &existing); err == nil {
			if existing.ChatPath != "" {
				cfg.ChatPath = existing.ChatPath
			}
			if existing.ChatBodyTemplate != "" {
				cfg.ChatBodyTemplate = existing.ChatBodyTemplate
			}
			if existing.ModelsPath != "" {
				cfg.ModelsPath = existing.ModelsPath
			}
			if existing.ModelsFormat != "" {
				cfg.ModelsFormat = existing.ModelsFormat
			}
			cfg.Stream = existing.Stream
			cfg.StartCommand = existing.StartCommand
		}
	}
	b, _ := json.Marshal(cfg)
	return string(b)
}

// DecodeArguments extrai Arguments e ModelsCommand de uma string codificada.
func DecodeArguments(encoded string) (arguments, modelsCommand string) {
	for i, c := range encoded {
		if c == '\n' {
			return encoded[:i], encoded[i+1:]
		}
	}
	return encoded, ""
}

// DecodeURLPaths extrai URLPaths de uma string JSON codificada.
func DecodeURLPaths(encoded string) URLPaths {
	if encoded == "" {
		return defaultURLPaths()
	}
	var cfg URLPaths
	if err := json.Unmarshal([]byte(encoded), &cfg); err != nil {
		return defaultURLPaths()
	}
	if cfg.ChatPath == "" {
		cfg.ChatPath = "/v1/chat/completions"
	}
	if cfg.ChatBodyTemplate == "" {
		cfg.ChatBodyTemplate = `{"model":"{{.Model}}","messages":[{"role":"user","content":"{{.Message}}"}],"stream":{{.Stream}}}`
	}
	if cfg.ModelsPath == "" {
		cfg.ModelsPath = "/config/providers"
	}
	if cfg.ModelsFormat == "" {
		cfg.ModelsFormat = "providers_obj"
	}
	return cfg
}

func defaultURLPaths() URLPaths {
	return URLPaths{
		ChatPath:         "/v1/chat/completions",
		ChatBodyTemplate: `{"model":"{{.Model}}","messages":[{"role":"user","content":"{{.Message}}"}],"stream":{{.Stream}}}`,
		ModelsPath:       "/config/providers",
		ModelsFormat:     "providers_obj",
		Stream:           false,
	}
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

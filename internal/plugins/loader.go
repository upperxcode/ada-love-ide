package plugins

import (
	"os"

	"gopkg.in/yaml.v3"
)

type expertsConfig struct {
	Experts []*ExpertPlugin `yaml:"experts"`
}

func LoadExperts(path string) ([]*ExpertPlugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config expertsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	experts := make([]*ExpertPlugin, 0, len(config.Experts))
	for _, e := range config.Experts {
		if e != nil && e.ID != "" {
			experts = append(experts, e)
		}
	}
	return experts, nil
}

func FindExpertByLanguage(language string, plugins []*ExpertPlugin) (*ExpertPlugin, bool) {
	for _, p := range plugins {
		if p.Language == language {
			return p, true
		}
	}
	return nil, false
}

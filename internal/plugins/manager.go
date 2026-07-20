package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type PluginManager struct {
	plugins    []*ExpertPlugin
	pluginsDir string
	mu         sync.RWMutex
}

func NewManager(pluginsDir string) (*PluginManager, error) {
	experts, err := LoadExperts(filepath.Join(pluginsDir, "experts.yaml"))
	if err != nil {
		return nil, err
	}

	return &PluginManager{
		plugins:    experts,
		pluginsDir: pluginsDir,
	}, nil
}

func (m *PluginManager) List() []*ExpertPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins
}

func (m *PluginManager) FindByLanguage(lang string) (*ExpertPlugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return FindExpertByLanguage(lang, m.plugins)
}

// CallExpert invokes the plugin binary in CLI mode via STDIO: it spawns
// `<StartCommand> <action>`, pipes `input` to stdin, and parses the JSON
// response from stdout. No ports are allocated and no process is left running
// after the call returns.
func (m *PluginManager) CallExpert(plugin *ExpertPlugin, action string, input string) (map[string]interface{}, error) {
	binPath := filepath.Join(m.pluginsDir, plugin.StartCommand)

	cmd := exec.Command(binPath, action)
	cmd.Stdin = strings.NewReader(input)

	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		msg := errb.String()
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("plugin %s falhou em '%s': %s", plugin.ID, action, strings.TrimSpace(msg))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("plugin %s retornou resposta inválida: %v", plugin.ID, err)
	}
	return result, nil
}

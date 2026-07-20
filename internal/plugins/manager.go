package plugins

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type runningPlugin struct {
	plugin *ExpertPlugin
	cmd    *exec.Cmd
}

type PluginManager struct {
	plugins    []*ExpertPlugin
	running    map[string]*runningPlugin
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
		running:    make(map[string]*runningPlugin),
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

func (m *PluginManager) EnsureRunning(plugin *ExpertPlugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.running[plugin.ID]; ok {
		if m.isHealthy(plugin.Endpoint) {
			return nil
		}
		m.stopLocked(plugin.ID)
	}

	port, err := m.findFreePort(8080, 8200)
	if err != nil {
		return err
	}

	plugin.Endpoint = fmt.Sprintf("http://localhost:%d", port)

	binPath := filepath.Join(m.pluginsDir, plugin.StartCommand)
	cmd := exec.Command(binPath, fmt.Sprintf("%d", port))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Dir = m.pluginsDir

	logPath := filepath.Join(m.pluginsDir, "logs")
	os.MkdirAll(logPath, 0755)
	logFile, _ := os.OpenFile(filepath.Join(logPath, plugin.ID+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin %s: %v", plugin.ID, err)
	}

	m.running[plugin.ID] = &runningPlugin{plugin: plugin, cmd: cmd}

	for i := 0; i < 30; i++ {
		if m.isHealthy(plugin.Endpoint) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	m.stopLocked(plugin.ID)
	return fmt.Errorf("timeout waiting for plugin %s", plugin.ID)
}

func (m *PluginManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id := range m.running {
		m.stopLocked(id)
	}
}

func (m *PluginManager) stopLocked(id string) {
	if rp, ok := m.running[id]; ok {
		if rp.cmd != nil && rp.cmd.Process != nil {
			rp.cmd.Process.Signal(os.Interrupt)
			pgid, err := syscall.Getpgid(rp.cmd.Process.Pid)
			if err == nil {
				syscall.Kill(-pgid, syscall.SIGKILL)
			} else {
				rp.cmd.Process.Kill()
			}
			rp.cmd.Wait()
		}
		delete(m.running, id)
	}
}

func (m *PluginManager) IsHealthy(endpoint string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isHealthy(endpoint)
}

func (m *PluginManager) isHealthy(endpoint string) bool {
	if endpoint == "" {
		return false
	}
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (m *PluginManager) findFreePort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			continue
		}
		l, err := net.ListenTCP("tcp", addr)
		if err == nil {
			l.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port in range %d-%d", start, end)
}

func (m *PluginManager) CallExpert(plugin *ExpertPlugin, action string) (map[string]interface{}, error) {
	if plugin.Endpoint == "" {
		return nil, fmt.Errorf("plugin %s not running", plugin.ID)
	}

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(plugin.Endpoint + "/" + action)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to expert: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expert returned error: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return result, nil
}
